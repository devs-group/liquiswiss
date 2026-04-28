.PHONY: up down restart logs ps build dc migrate

# Optional: pull access token from macOS keychain. If absent, compose runs with
# its in-file defaults (external integrations like email and currency disabled).
# Setup (team members with vault access):
#   security add-generic-password -a "$USER" -s "bws-liquiswiss-dev" -w
BWS_ACCESS_TOKEN := $(shell security find-generic-password -a "$$USER" -s "bws-liquiswiss-dev" -w 2>/dev/null)

# Use bws when token available, otherwise fall through to bare docker compose.
ifeq ($(strip $(BWS_ACCESS_TOKEN)),)
  RUN := docker compose
else
  RUN := BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose
endif

# Generic compose passthrough: `make dc CMD="logs -f backend"`
dc:      ; @$(RUN) $(CMD)
up:      ; @$(RUN) up -d
down:    ; @$(RUN) down
restart: ; @$(RUN) restart
logs:    ; @$(RUN) logs -f
ps:      ; @$(RUN) ps
build:   ; @$(RUN) up -d --build

# Apply pending static schema migrations against the compose database.
migrate: ; @$(MAKE) -C backend goose-static-up
