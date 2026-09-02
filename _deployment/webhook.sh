#!/bin/bash
# Deploy webhook script. Lives on the production host next to `.credentials`
# (see `.credentials.example`) and `.env` (image tags, see `.env.example`), and
# is triggered by the host's adnanh/webhook listener.
#
# Modes:
#   ./webhook.sh                       start/refresh every service from .env
#   ./webhook.sh <tag> <service> [scope]  deploy one service at one image tag
#
# Deploy mode pulls the tag, recreates only that service, waits for it to become
# healthy and rolls the image back if it does not. The new tag reaches .env only
# after a successful health check, so a failed deploy leaves no trace behind.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"
LOG_FILE="$SCRIPT_DIR/webhook.log"
exec > >(tee -a "$LOG_FILE") 2>&1

# Credentials early so notify_slack works in every mode.
# Required in .credentials:
#   CI_DEPLOY_USER / CI_DEPLOY_PASSWORD - registry login
#   REGISTRY_HOST                       - container registry hostname
#   SLACK_HOOK_URL                      - Slack incoming webhook
#   APP_URL                             - URL named in the Slack message
#   BWS_ACCESS_TOKEN / BWS_PROJECT_ID   - Bitwarden Secrets Manager (prod project)
if [ ! -f ./.credentials ]; then
  echo "❌ Missing ${SCRIPT_DIR}/.credentials (see .credentials.example)"
  exit 1
fi
# shellcheck disable=SC1091
. ./.credentials
export BWS_ACCESS_TOKEN

# All compose calls need the BWSM secrets injected: the compose file marks the
# production values as required (${X:?}), so bare `docker compose` cannot even
# interpolate, let alone start anything.
dc() {
  bws run --project-id "$BWS_PROJECT_ID" -- docker compose "$@"
}

notify_slack() {
  local msg="$1"
  [ -z "${SLACK_HOOK_URL:-}" ] && return 0
  local tail_log
  tail_log=$(tail -15 "$LOG_FILE" 2>/dev/null | sed 's/"/\\"/g' | sed ':a;N;$!ba;s/\n/\\n/g')
  curl -sf -m 10 -X POST -H 'Content-Type: application/json' \
    -d "{\"text\":\":rotating_light: *LiquiSwiss deploy webhook failed* ($(hostname))\\n${msg}\\n\`\`\`${tail_log}\`\`\`\"}" \
    "$SLACK_HOOK_URL" >/dev/null 2>&1 || true
}

notify_slack_success() {
  [ -z "${SLACK_HOOK_URL:-}" ] && return 0
  curl -sf -m 10 -X POST -H 'Content-Type: application/json' \
    -d "{\"text\":\":white_check_mark: $1\"}" \
    "$SLACK_HOOK_URL" >/dev/null 2>&1 || true
}

# Default trap; deploy mode replaces it with one that rolls back first.
trap 'notify_slack "mode=${2:-all} args=${1:-none} line=$LINENO"; exit 1' ERR

registry_login() {
  echo "$CI_DEPLOY_PASSWORD" | docker login -u "$CI_DEPLOY_USER" --password-stdin "$REGISTRY_HOST"
}

# ── No-args mode: bring up all services using the tags in .env ──
if [ -z "${1:-}" ]; then
  echo "$(date '+%Y-%m-%d %H:%M:%S') [$$] Starting all services from .env"

  for VAR in BACKEND_TAG NUXT_TAG; do
    if ! grep -q "^${VAR}=" .env 2>/dev/null; then
      echo "ERROR: ${VAR} not set in .env (see .env.example)"
      notify_slack "startup: ${VAR} not set in .env"
      exit 1
    fi
  done

  registry_login
  dc pull
  dc up -d --remove-orphans

  echo "$(date '+%Y-%m-%d %H:%M:%S') [$$] All services started"
  exit 0
fi

# ── Deploy mode: update a single service to a new image tag ──
IMAGE_TAG="$1"
SERVICE="${2:-backend}"
# Deploy scope from CI ("release vX.Y.Z backend+nuxt" | "manual" | empty):
# cosmetics for the Slack message only.
SCOPE="${3:-}"

case "$SERVICE" in
  backend|nuxt) ;;
  *) echo "ERROR: unknown service '${SERVICE}' (expected backend or nuxt)"; exit 1 ;;
esac

# Service name to tag var: backend -> BACKEND_TAG, nuxt -> NUXT_TAG.
TAG_VAR="$(echo "${SERVICE}" | tr '[:lower:]-' '[:upper:]_')_TAG"

touch .env
PREV_TAG=$(grep -oP "^${TAG_VAR}=\K.*" .env || echo "latest")

# The new tag is exported only; .env is written after the health check passes.
export "$TAG_VAR"="$IMAGE_TAG"

persist_tag() {
  if grep -q "^${TAG_VAR}=" .env; then
    sed -i "s/^${TAG_VAR}=.*/${TAG_VAR}=${IMAGE_TAG}/" .env
  else
    echo "${TAG_VAR}=${IMAGE_TAG}" >> .env
  fi
}

rollback() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') [$$] Rolling back ${TAG_VAR} to ${PREV_TAG}"
  if grep -q "^${TAG_VAR}=" .env; then
    sed -i "s/^${TAG_VAR}=.*/${TAG_VAR}=${PREV_TAG}/" .env
  fi
  export "$TAG_VAR"="$PREV_TAG"
  dc up -d --no-deps "$SERVICE" || true
}

trap 'rollback; notify_slack "deploy ${SERVICE}@${IMAGE_TAG} failed, rolled back to ${PREV_TAG}"; exit 1' ERR

# One deployment at a time.
DEPLOY_LOCK="/tmp/liquiswiss-deploy.lock"
exec 200>"$DEPLOY_LOCK"
if ! flock -n 200; then
  echo "$(date '+%Y-%m-%d %H:%M:%S') [$$] Waiting for deploy lock (SERVICE=$SERVICE IMAGE_TAG=$IMAGE_TAG)"
  flock 200 || { echo "ERROR: Failed to acquire deploy lock"; exit 1; }
fi
echo "$(date '+%Y-%m-%d %H:%M:%S') [$$] Deployment started (SERVICE=$SERVICE IMAGE_TAG=$IMAGE_TAG)"

registry_login

echo "Pulling $SERVICE image"
dc pull "$SERVICE"

echo "Recreating $SERVICE container"
dc up -d --no-deps "$SERVICE"

# Read the container's own healthcheck result rather than exec'ing a probe:
# `bws run` joins its arguments and hands them to `sh -c`, so anything with
# shell metacharacters (the nuxt probe's parentheses) dies with a syntax error
# before it ever reaches the container. Both services declare a healthcheck in
# the compose file, so docker has the answer already.
echo "Waiting for $SERVICE to be healthy..."
CID="$(dc ps -q "$SERVICE")"
if [ -z "$CID" ]; then
  echo "ERROR: no container id for $SERVICE"
  rollback
  notify_slack "deploy ${SERVICE}@${IMAGE_TAG}: no container after up, rolled back to ${PREV_TAG}"
  exit 1
fi
MAX_ATTEMPTS=30
ATTEMPT=0
HEALTHY=0
while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
  STATE="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$CID" 2>/dev/null || echo unknown)"
  case "$STATE" in
    healthy) HEALTHY=1 ;;
    # An image without its own HEALTHCHECK and no compose override cannot be
    # waited on; treat a running container as good rather than rolling back.
    none) docker inspect --format '{{.State.Status}}' "$CID" 2>/dev/null | grep -q '^running$' && HEALTHY=1 ;;
  esac
  [ "$HEALTHY" -eq 1 ] && { echo "$SERVICE is healthy"; break; }
  ATTEMPT=$((ATTEMPT + 1))
  echo "Waiting... (attempt $ATTEMPT/$MAX_ATTEMPTS)"
  sleep 2
done

if [ "$HEALTHY" -ne 1 ]; then
  echo "ERROR: $SERVICE not healthy after 60s"
  rollback
  notify_slack "deploy ${SERVICE}@${IMAGE_TAG} failed health check, rolled back to ${PREV_TAG}"
  exit 1
fi

# Deployed: persist the tag, then stop rolling back on later errors.
persist_tag
trap - ERR

# Bring up anything that is not running, but never recreate the service just
# deployed (that would drop its first-boot logs).
dc up -d --no-recreate --remove-orphans

echo "$(date '+%Y-%m-%d %H:%M:%S') [$$] Deployment complete"

# SCOPE from CI: "release vX.Y.Z <parts>" or just <parts>, where parts is
# backend+nuxt | backend | nuxt.
REL=""; PARTS="$SCOPE"
if [[ "$SCOPE" == release* ]]; then
  REL="$(echo "$SCOPE" | awk '{print $2}')"
  PARTS="$(echo "$SCOPE" | awk '{print $3}')"
fi
case "$PARTS" in
  backend+nuxt) WHAT="backend and nuxt" ;;
  backend)      WHAT="backend" ;;
  nuxt)         WHAT="nuxt" ;;
  *)            WHAT="$SERVICE" ;;
esac
MSG="LiquiSwiss${REL:+ *${REL}*} ${WHAT} deployed to ${APP_URL}"
notify_slack_success "${MSG} (\`${IMAGE_TAG:0:7}\`)"
