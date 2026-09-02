# Deployment

Follows the devs-group standard (kikal is the reference implementation), with
two deviations noted under "Deviations from the standard".

## Release model

Tag-driven. A push to `main` runs tests only; the release is the tag:

```bash
git tag v0.3.0
git push --tags
```

That builds both images with the version baked in and deploys backend and nuxt.
`workflow_dispatch` on `main` does the same from HEAD, with the version resolved
to the last tag.

The running version is public: `curl https://api.liquiswiss.ch/api/config` returns
`sha`, `version` and `appVersion`. CI reads the same endpoint before a deploy to
work out which parts a release actually changes (for the Slack message).

## Layout

- `backend/Dockerfile`, `nuxt/Dockerfile` — production images. Both take
  `GIT_SHA` / `VERSION` / `APP_SHA` build args; the backend stamps them into the
  binary via ldflags and serves them on `/api/config`.
- `deploy.sh` — CI-side trigger: `deploy.sh <tag> <service> [scope]`.
- `webhook.sh` — host-side deploy script. Pulls the tag, recreates one service,
  waits for it to become healthy, rolls the image back if it does not, and only
  then writes the tag to `.env`. Serialised with a flock, logs to `webhook.log`,
  reports success and failure to Slack.
- `sync.sh` — pushes `docker-compose.yml`, `webhook.sh`, `rotate-db.sh`,
  `Makefile` and `.env.example` to the host. Never touches `.env`/`.credentials`.
- `.env.example` — the image tag vars the deployment understands.
- `.credentials.example` — registry, Slack and BWSM values consumed by `webhook.sh`.
- `webhook/hooks.json.example`, `webhook/webhooks.service` — host webhook listener.

## Files on the production host

| File | Source | Notes |
|---|---|---|
| `docker-compose.yml` | `sync.sh` | |
| `webhook.sh` | `sync.sh` | `chmod +x` |
| `Makefile`, `rotate-db.sh` | `sync.sh` | |
| `.env.example` | `sync.sh` | reference only |
| `.env` | hand-written once | `BACKEND_TAG` / `NUXT_TAG`, rewritten by deploys |
| `.credentials` | hand-written once | `chmod 600`, BWSM token + registry + Slack |
| `webhook.log` | `webhook.sh` | deploy and rollback history |

`sync.sh` never writes `.env` or `.credentials`, so a deploy can never clobber a
secret. The flip side: a new setting reaches the host in two steps, the template
with the sync and the value by hand.

## Production host setup

1. `sync.sh` the deploy files into the deploy directory, `chmod +x webhook.sh`.
2. Copy `.credentials.example` to `.credentials`, fill in real values, `chmod 600`.
3. Copy `.env.example` to `.env` and seed `BACKEND_TAG` / `NUXT_TAG` (`latest` for
   the very first start).
4. Copy `webhook/hooks.json.example` into the host's hooks file, replace the token
   placeholder with `openssl rand -hex 32`, and point `execute-command` and
   `command-working-directory` at the deploy directory. `pass-arguments-to-command`
   must list `tag`, `service` and `scope` in that order or they are silently dropped.
5. Copy `webhook/webhooks.service` into the systemd unit directory, adjust paths,
   `systemctl daemon-reload && systemctl enable --now webhooks`.
6. Allow the reverse proxy's bridge network to reach the webhook port.
7. Start everything once with the no-args mode: `./webhook.sh`. Do this before the
   first CI deploy, which uses `--no-deps` and never creates the databases.

## CI configuration

Deploy secrets are scoped to the `production` GitHub Environment, never to the
repository:

```bash
gh secret set DEPLOY_SECRET --env production   # matches the hooks file token
gh secret set DEPLOY_URL    --env production   # https://<webhook-host>/hooks/<id>
```

## Migrations

The rollback swaps the image; it does nothing to the database. Before a release
that migrates, take a dump, and keep migrations forward-only and safe against the
live schema (add nullable, backfill, constrain in separate releases). A migration
that fails halfway leaves goose's `goose_db_version` behind the code, and the
rolled-back image can be just as unable to start as the new one.

```bash
ssh <host> "cd <deploy-dir> && make dump-db"
```

## Deviations from the standard

- **Secrets come from Bitwarden Secrets Manager, not the server `.env`.** Every
  compose call goes through `bws run --project-id`, so `webhook.sh` wraps compose
  in a `dc()` helper and a bare `docker compose` on the host fails on the required
  `${X:?}` vars. The server `.env` holds image tags only.
- **Service names stay `backend` and `nuxt`** (the standard suggests
  `<name>-backend`), so the tag vars are `BACKEND_TAG` and `NUXT_TAG`. Renaming
  them would rename the containers under the already-`liquiswiss` compose project.

## Debugging

```bash
ssh <host> "tail -30 <deploy-dir>/webhook.log"
ssh <host> "cd <deploy-dir> && make ps && make dc CMD='logs backend --tail 20'"
curl -s https://api.liquiswiss.ch/api/config
```

- Deploy "succeeded" but nothing changed: check `BACKEND_TAG` / `NUXT_TAG` in the
  host `.env` against the deployed sha.
- A changed `.env` that seems ignored: `docker restart` does not reload it, use
  `make dc CMD="up -d --force-recreate <service>"`.
