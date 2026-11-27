#!/usr/bin/env bash
# Replace the local dev database with production data, keeping local login and MCP tokens.
# Usage: make db-pull-prod   (host comes from the gitignored _deployment/.sync.env)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

[ -f _deployment/.sync.env ] && . _deployment/.sync.env
: "${LS_HOST:?LS_HOST missing (see _deployment/.sync.env)}"

DB=liquiswiss
KEEP_TABLES=(refresh_tokens oauth_clients oauth_auth_codes oauth_refresh_tokens)
STAMP="$(date +%Y%m%d-%H%M%S)"
DIR="$ROOT/.dumps"
mkdir -p "$DIR"

local_db() { docker compose exec -T database "$@"; }
LOCAL_AUTH=(-uliquiswiss -ppassword)

echo "1/6 Backing up local database to .dumps/local-$STAMP.sql"
local_db mariadb-dump "${LOCAL_AUTH[@]}" --single-transaction --routines --triggers "$DB" > "$DIR/local-$STAMP.sql"

echo "2/6 Saving local tokens (${KEEP_TABLES[*]})"
local_db mariadb-dump "${LOCAL_AUTH[@]}" --no-create-info --complete-insert --skip-triggers "$DB" "${KEEP_TABLES[@]}" > "$DIR/local-tokens-$STAMP.sql"

echo "3/6 Dumping production from $LS_HOST (token tables schema only)"
IGNORE=""
for t in "${KEEP_TABLES[@]}"; do IGNORE="$IGNORE --ignore-table=$DB.$t"; done
# Credentials stay on the host: the dump runs with the container's own env (prod compose needs BWSM to parse, so plain docker exec)
REMOTE_DUMP="docker exec \$(docker ps -qf label=com.docker.compose.service=database-app) sh -c"
ssh "$LS_HOST" "$REMOTE_DUMP 'mariadb-dump -u\"\$MARIADB_USER\" -p\"\$MARIADB_PASSWORD\" --single-transaction --routines --triggers $IGNORE $DB'" > "$DIR/prod-$STAMP.sql"
ssh "$LS_HOST" "$REMOTE_DUMP 'mariadb-dump -u\"\$MARIADB_USER\" -p\"\$MARIADB_PASSWORD\" --no-data --skip-triggers $DB ${KEEP_TABLES[*]}'" >> "$DIR/prod-$STAMP.sql"

echo "4/6 Importing production dump"
local_db mariadb "${LOCAL_AUTH[@]}" -e "DROP DATABASE IF EXISTS $DB; CREATE DATABASE $DB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
local_db mariadb "${LOCAL_AUTH[@]}" "$DB" < "$DIR/prod-$STAMP.sql"

echo "5/6 Applying pending migrations and restoring local tokens"
make --no-print-directory migrate
{
  echo "SET FOREIGN_KEY_CHECKS=0;"
  cat "$DIR/local-tokens-$STAMP.sql"
  # Tokens of users that production does not know would point nowhere
  for t in refresh_tokens oauth_auth_codes oauth_refresh_tokens; do
    echo "DELETE FROM $t WHERE user_id NOT IN (SELECT id FROM users);"
  done
  echo "SET FOREIGN_KEY_CHECKS=1;"
} | local_db mariadb "${LOCAL_AUTH[@]}" "$DB"

echo "6/6 Restarting backend (recreates dynamic migrations)"
make --no-print-directory dc CMD="restart backend"

echo "Done. Backups in .dumps/ (local-$STAMP.sql, prod-$STAMP.sql)"
