---
name: liquiswiss-release
description: Cut a LiquiSwiss release - suggest the next version tag, tag and push after confirmation, watch the deploy, verify prod and write the GitHub release. Use whenever the user wants to release, deploy to prod, ship to production, tag a version, or talks about tags, releases or "rausbringen" for this repo.
user-invocable: true
---

# LiquiSwiss release

Deploys are tag-driven since 2026-09-02. Pushing a `v*` tag builds both images
(`backend` and `nuxt`, each tagged `:latest`, `:<sha>` and `:vX.Y.Z`) and deploys
them to production. **A push to `main` runs tests only and deploys nothing.**
The GitHub release is created HERE via the gh CLI, not by CI.

Manual escape hatch without a tag: `gh workflow run "CI/CD Pipeline" --ref main`
builds and deploys HEAD, with the version resolved to the last tag.

## Steps

1. **Preflight**
   - Branch must be `main`, `git status` clean, nothing unpushed. The tag has to
     point at a commit that exists on the remote. Sort that out first, and never
     tag around it.
   - Show what would ship: `git log --oneline $(git describe --tags --abbrev=0)..HEAD`.
   - **If the release contains a goose migration, take a dump first**:
     `ssh liquiswiss "cd /root/liquiswiss && make dump-db"`. The automatic
     rollback swaps the image and does nothing to the database, so a bad
     migration leaves both versions unable to start. Migrations must be forward
     only and safe against the live schema.

2. **Suggest the next version**
   - Last tag: `git describe --tags --abbrev=0`.
   - Bump heuristic from the commits since that tag: anything user-facing or new
     gives a minor bump, pure fixes and chores give a patch bump. Major stays
     manual.
   - Present the suggestion together with the commit list and ASK for
     confirmation. The user may name a different version.

3. **Tag and push** (only after explicit confirmation)
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```

4. **Watch the release run**
   - The tag push creates its own run: find it with
     `gh run list --limit 3 --json databaseId,headBranch,status` and pick the one
     whose `headBranch` is the tag.
   - `gh run watch <id> --exit-status --interval 20`. Note that a `main` run may
     be in flight at the same time, do not watch that one.
   - Jobs in order: Detect Changes, Resolve Version, Lint Nuxt, Test Backend,
     E2E Tests, Build Nuxt, Build Backend, Deploy to Production.
   - On failure: STOP, report, no GitHub release. Do not delete the tag unless
     the user asks. `gh run rerun <id> --failed` is usually the fix.

5. **Verify production**
   ```bash
   curl -s https://api.liquiswiss.ch/api/config
   ```
   `version` must be the new tag and `sha` the released commit. Also confirm both
   services actually carry the new tag:
   ```bash
   ssh liquiswiss 'cat /root/liquiswiss/.env; docker ps --format "{{.Names}}\t{{.Image}}\t{{.Status}}" | grep -E "nuxt|backend"'
   ```
   `BACKEND_TAG` and `NUXT_TAG` must both be the new sha and both containers
   `(healthy)`.

6. **If a service rolled back**
   The deploy is per service: the webhook pulls, recreates with `--no-deps`,
   waits for the container healthcheck and rolls the image back on failure, so
   one service can ship while the other does not. Symptom is a deploy job that
   fails after one service succeeded.
   ```bash
   ssh liquiswiss "tail -40 /root/liquiswiss/webhook.log"
   ```
   Fix the cause, then redeploy. Do not hand-patch prod.

7. **GitHub release** (only after CI success and a verified prod)
   - Write the notes YOURSELF. Not `--generate-notes`, not a pasted commit list.
   - Follow the existing house style in this repo (see `gh release view v1.2.2`):
     a `## Highlights` section with bold lead-ins per item, then `## Deploy notes`
     naming migrations, new env vars and anything operational, then `## Commits`
     with the one-line log.
   - Language: English. No em-dashes.
   ```bash
   gh release create vX.Y.Z --title vX.Y.Z --notes-file <path-to-notes>
   ```

## Rules that bite here

- **Never push without asking**, tags included. This skill asks in step 2 and
  that is the only consent that counts.
- Commit messages: single line, capitalised, no footers.
- The pre-commit hook runs `npm run lint:fix` and `go test ./...`, and the Go
  tests need the local test database (`docker compose up -d database-testing`,
  port 3318). It also runs `git add -u`, so a grouped commit quietly picks up
  every other modified file. Check `git show --stat` afterwards.
- Compose changes under `_deployment/` do not travel with a release. They ship
  via `_deployment/sync.sh` (preview with `--check` first), and only take effect
  with `--up` or the next deploy.
- Secrets for the deploy live in the GitHub Environment `production`
  (`DEPLOY_URL`, `DEPLOY_SECRET`). Repository-level deploy secrets are gone on
  purpose, do not recreate them.
- Production runs on `ssh liquiswiss` at `/root/liquiswiss`, which is not a git
  checkout. The pipeline details are in the `dg:deploy-setup` skill and in
  `_deployment/README.md`.
