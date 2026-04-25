.PHONY: up down restart logs ps build dc _check

# Pull token from macOS keychain (set via: security add-generic-password -a "$USER" -s "bws-liquiswiss-dev" -w)
BWS_ACCESS_TOKEN := $(shell security find-generic-password -a "$$USER" -s "bws-liquiswiss-dev" -w 2>/dev/null)

# Required BWSM secrets (no compose fallback). Reports ALL missing at once.
REQUIRED_SECRETS := SEND_GRID_TOKEN SEND_GRID_TEMPLATE_ID FIXER_IO_KEY

define PREFLIGHT
missing=""; \
for var in $(REQUIRED_SECRETS); do \
  if [ -z "$$(printenv $$var)" ]; then missing="$$missing $$var"; fi; \
done; \
if [ -n "$$missing" ]; then \
  echo "ERROR: missing required BWSM secrets:" >&2; \
  for v in $$missing; do echo "  - $$v" >&2; done; \
  exit 1; \
fi
endef

_check:
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- sh -c '$(PREFLIGHT)'

# Generic compose passthrough: `make dc CMD="logs -f backend"`
dc: _check
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose $(CMD)

up: _check
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose up -d

down:
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose down

restart: _check
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose restart

logs:
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose logs -f

ps:
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose ps

build: _check
	@BWS_ACCESS_TOKEN="$(BWS_ACCESS_TOKEN)" bws run -- docker compose up -d --build
