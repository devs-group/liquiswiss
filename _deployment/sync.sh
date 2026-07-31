#!/bin/bash
# Sync the deploy files (compose + webhook + rotate-db + Makefile) from this dir
# to the prod server, with a preview first. Secrets (.env, .credentials) live
# only on the server and are never touched.
#
# Usage:
#   ./sync.sh              preview the diff, then ask before applying
#   ./sync.sh --check      preview only, never apply (exit 0)
#   ./sync.sh --yes        apply without the confirm prompt
#   ./sync.sh --up         after applying, run `make dc CMD="up -d"` on prod
#                          (creates new services, recreates changed ones).
#                          Runs even when no files differ, so it doubles as a
#                          recreate-only command.
#
# This repo is public, so the target host and path are NOT hardcoded here.
# Provide them one of two ways:
#   1. a gitignored `_deployment/.sync.env` next to this script:
#        LS_HOST=<ssh alias>
#        LS_DEST=<remote dir>
#   2. environment variables: LS_HOST=... LS_DEST=... ./sync.sh
set -euo pipefail

cd "$(cd "$(dirname "$0")" && pwd)"

# shellcheck source=/dev/null
[ -f .sync.env ] && . ./.sync.env

LS_HOST="${LS_HOST:-}"
LS_DEST="${LS_DEST:-}"
if [ -z "$LS_HOST" ] || [ -z "$LS_DEST" ]; then
  echo "ERROR: LS_HOST and LS_DEST must be set (see _deployment/.sync.env or the header of this script)" >&2
  exit 2
fi
FILES=(docker-compose.yml webhook.sh rotate-db.sh Makefile)

APPLY_PROMPT=1
DO_APPLY=1
DO_UP=0
for arg in "$@"; do
  case "$arg" in
    --check) DO_APPLY=0 ;;
    --yes|-y) APPLY_PROMPT=0 ;;
    --up) DO_UP=1 ;;
    *) echo "unknown arg: $arg" >&2; exit 2 ;;
  esac
done

# Validate the compose file before it can reach prod. Production values come
# from BWSM / the server-side .env, which is not present here, and vars marked
# required (${X:?...}) fail interpolation without SOME value. So drop a
# throwaway .env with dummies for every required var: that isolates real
# YAML/schema errors from a merely absent .env.
if command -v docker >/dev/null 2>&1; then
  tmpenv=0
  if [ ! -f .env ]; then
    grep -oE '\$\{[A-Z0-9_]+:\?' docker-compose.yml | sed 's/[${]*//;s/:?//;s/$/=dummy/' | sort -u >.env
    tmpenv=1
  fi
  ok=1
  docker compose -f docker-compose.yml config -q 2>/dev/null || ok=0
  [ "$tmpenv" -eq 1 ] && rm -f .env
  if [ "$ok" -eq 0 ]; then
    echo "ERROR: docker-compose.yml is invalid (run: docker compose -f docker-compose.yml config)" >&2
    exit 1
  fi
fi

echo "== Preview: $(printf '%s ' "${FILES[@]}")-> ${LS_HOST}:${LS_DEST}/"
tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT
CHANGED=0
for f in "${FILES[@]}"; do
  [ -f "$f" ] || { echo "  (skip, missing locally: $f)"; continue; }
  # Fetch the remote copy to a file (not a var: command substitution strips
  # trailing newlines and would show phantom diffs), then diff file-to-file.
  scp -q "${LS_HOST}:${LS_DEST}/$f" "$tmpdir/$f" 2>/dev/null || : >"$tmpdir/$f"
  if diff -q "$tmpdir/$f" "$f" >/dev/null 2>&1; then
    echo "  = $f (up to date)"
  else
    CHANGED=$((CHANGED + 1))
    echo "  ~ $f (differs):"
    diff "$tmpdir/$f" "$f" | sed 's/^/      /' || true
  fi
done

if [ "$DO_APPLY" -eq 0 ]; then
  [ "$CHANGED" -eq 0 ] && echo "Nothing to sync." || echo "(--check) not applying."
  exit 0
fi

# No file changes still leaves --up meaningful: it recreates services whose
# config drifted from what is running, so fall through instead of exiting.
if [ "$CHANGED" -eq 0 ]; then
  echo "Nothing to sync."
  [ "$DO_UP" -eq 0 ] && exit 0
else
  if [ "$APPLY_PROMPT" -eq 1 ]; then
    printf "Apply these %d change(s) to %s? [y/N] " "$CHANGED" "$LS_HOST"
    read -r ans
    case "$ans" in y|Y|yes) ;; *) echo "Aborted."; exit 0 ;; esac
  fi

  for f in "${FILES[@]}"; do
    [ -f "$f" ] || continue
    scp -q "$f" "${LS_HOST}:${LS_DEST}/$f"
    echo "  synced $f"
  done
fi

if [ "$DO_UP" -eq 1 ]; then
  # Bare `docker compose` fails on prod: required vars come from BWSM, which the
  # Makefile injects via `bws run`. Always go through `make dc`.
  echo "== make dc CMD=\"up -d\" on ${LS_HOST}"
  ssh "$LS_HOST" "cd '${LS_DEST}' && make dc CMD='up -d'"
fi

echo "Done."
