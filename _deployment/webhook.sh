#!/bin/bash
# Deploy webhook script. Place on the deploy host next to a `.credentials` file
# (see `.credentials.example`) and trigger via the host's webhook listener.

set -o pipefail

# Required env vars supplied by .credentials:
#   CI_DEPLOY_USER     - registry user
#   CI_DEPLOY_PASSWORD - registry token
#   REGISTRY_HOST      - container registry hostname
#   SLACK_HOOK_URL     - Slack incoming webhook URL
#   APP_URL            - URL announced in the Slack notification
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "${SCRIPT_DIR}/.credentials" ]; then
  # shellcheck disable=SC1091
  . "${SCRIPT_DIR}/.credentials"
else
  echo "❌ Missing ${SCRIPT_DIR}/.credentials (see .credentials.example)"
  exit 1
fi

{
  if ! echo "$CI_DEPLOY_PASSWORD" | docker login -u "$CI_DEPLOY_USER" --password-stdin "$REGISTRY_HOST"; then
    echo "❌ Docker login failed. Check or rotate the registry token."
    exit
  fi
  echo "✅ Docker login succeeded"

  if ! docker compose pull --policy "always"; then
    echo "❌ ERROR pulling images"
    exit
  fi
  echo "✅ Pulled new images"

  # Only recreate containers whose image or config changed; leave the rest running.
  if ! docker compose up -d --remove-orphans; then
    echo "❌ ERROR recreating containers"
    exit
  fi
  echo "✅ Recreated changed containers"

  echo "🚀 Deployed successfully"
} | xargs -0 -I output curl -X POST -H 'Content-type: application/json' \
  --data "{\"text\": \"New Deployment to ${APP_URL}:\n\noutput\"}" \
  "${SLACK_HOOK_URL}"
