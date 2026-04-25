#!/bin/sh
set -x

SUCCEEDED=$(curl -X GET "${DEPLOY_URL}" -H "X-Access-Token: ${DEPLOY_TOKEN}" | grep "OK" | wc -l)

if [ "${SUCCEEDED}" != 1 ] ; then
  echo "Deployment failed!"
  exit 1
fi

echo "Deployment succeeded"
