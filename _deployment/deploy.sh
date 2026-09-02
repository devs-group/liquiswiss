#!/bin/bash
# CI-side trigger for the production deploy webhook.
# Usage: deploy.sh <image-tag> <service> [scope]
set -euo pipefail

IMAGE_TAG="${1:-latest}"
SERVICE="${2:-backend}"
# What this deploy carries, for the Slack message: "release vX.Y.Z <parts>"
# or "manual" (workflow_dispatch).
SCOPE="${3:-}"

# -f: the webhook answers non-2xx on a failed deploy, which must fail the job.
curl -fsS --show-error -X POST \
  -H "X-Access-Token: ${DEPLOY_SECRET}" \
  -H "Content-Type: application/json" \
  --data "{\"tag\": \"${IMAGE_TAG}\", \"service\": \"${SERVICE}\", \"scope\": \"${SCOPE}\"}" \
  "${DEPLOY_URL}"

echo
echo "Deployment succeeded for ${SERVICE} (${IMAGE_TAG}) scope=${SCOPE:-unspecified}"
