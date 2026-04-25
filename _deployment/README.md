# Deployment

## Layout

- `backend/Dockerfile`, `nuxt/Dockerfile` — production images built and pushed by CI.
- `deploy.sh` — CI-side script that calls the deploy webhook on the production host.
- `webhook.sh` — host-side script run by `adnanh/webhook` to pull new images and recreate changed containers. Reads secrets from `.credentials` (untracked, kept next to the script on the host).
- `webhook/hooks.json.example` — template for the host's hooks file listing the registered webhook endpoint and access token.
- `webhook/webhooks.service` — systemd unit template that runs `adnanh/webhook`.
- `.credentials.example` — template for the untracked `.credentials` file consumed by `webhook.sh`.

## Production host setup

1. Place `webhook.sh` on the host at the chosen deploy directory and make it executable.
2. Copy `.credentials.example` to `.credentials` next to `webhook.sh` and fill in real values. Restrict to `chmod 600`.
3. Copy `webhook/hooks.json.example` to the host's hooks file and replace the token placeholder with a strong random value. Update `execute-command` and `command-working-directory` to point to the deploy directory chosen above.
4. Copy `webhook/webhooks.service` to the host's systemd unit directory and adjust paths/user to match the deploy directory.
5. `systemctl daemon-reload && systemctl enable --now webhooks`.
6. Allow the reverse proxy's bridge network to reach the webhook port through the host firewall.

## CI configuration

GitHub Actions repo secrets:

- `DEPLOY_URL` — full URL of the deploy webhook endpoint on the host.
- `DEPLOY_TOKEN` — must match the `value` set in the hooks file.

## Updating the webhook

Edit `webhook.sh` in this repo, push, then on the host pull the change and the next webhook trigger uses the new version. Credentials never need to be touched unless rotated.
