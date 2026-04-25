#!/usr/bin/env bash
# Interactive MariaDB password rotation for app/wp DBs.
# Sources .credentials for BWSM token, runs ALTER USER via stdin pipe.
# Usage: bash rotate-db.sh app|wp
set -euo pipefail

WHICH="${1:-}"
case "$WHICH" in
  app) SVC="database-app";       APP_USER="liquiswiss";    LABEL="APP";;
  wp)  SVC="database-wordpress"; APP_USER="liquiswiss-wp"; LABEL="WP";;
  *)   echo "Usage: $0 app|wp" >&2; exit 1;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "${SCRIPT_DIR}/.credentials"
export BWS_ACCESS_TOKEN

read -r -s -p "Old ${LABEL} root password: "    OLD_ROOT; echo
read -r -s -p "New ${LABEL} user (${APP_USER}) password: " NEW_USER; echo
read -r -s -p "New ${LABEL} root password: "    NEW_ROOT; echo

SQL=$(cat <<EOF
ALTER USER '${APP_USER}'@'%' IDENTIFIED BY '${NEW_USER}';
ALTER USER 'root'@'%' IDENTIFIED BY '${NEW_ROOT}';
ALTER USER 'root'@'localhost' IDENTIFIED BY '${NEW_ROOT}';
FLUSH PRIVILEGES;
EOF
)

echo "$SQL" | bws run --project-id "$BWS_PROJECT_ID" -- docker compose exec -T "$SVC" mariadb -u root -p"$OLD_ROOT"

case "$WHICH" in
  app) echo "✅ App DB rotated. Update BWSM: APP_DB_PASSWORD, APP_DB_ROOT_PASSWORD, then: make deploy";;
  wp)  echo "✅ WP DB rotated. Update BWSM: WP_DB_PASSWORD, WP_DB_ROOT_PASSWORD, then: make deploy";;
esac
