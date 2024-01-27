#!/bin/sh

SUCCEEDED=$(curl -X GET "${DEPLOY_URL}" -H "X-Access-Token: ${DEPLOY_TOKEN}" | grep "upn" | wc -l)

if [ "${SUCCEEDED}" != 1 ] ; then
  echo "Deployment failed!"
  exit 1
fi

echo "Deployment succeeded"
