.PHONY: dc up down restart logs ps deploy goose-up goose-down goose-down-to-0 goose-down-to goose-down-to-and-up dump-db rotate-db-app rotate-db-wp

# Source BWSM token + project id from .credentials, run docker compose with secrets injected.
# Usage: make dc CMD="exec database-app mariadb -u root -p..."
WRAP = . ./.credentials && export BWS_ACCESS_TOKEN && bws run --project-id "$$BWS_PROJECT_ID" -- docker compose

dc:
	@$(WRAP) $(CMD)

up:
	@$(WRAP) up -d --remove-orphans

down:
	@$(WRAP) down

restart:
	@$(WRAP) restart

logs:
	@$(WRAP) logs -f

ps:
	@$(WRAP) ps

deploy:
	@bash ./webhook.sh

# Goose migrations (DSN built from injected BWSM secrets)
DB_DSN = $$APP_DB_USER:$$APP_DB_PASSWORD@tcp(database-app:3306)/liquiswiss

goose-up:
	@$(WRAP) run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" up

goose-down:
	@$(WRAP) run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down

goose-down-to-0:
	@$(WRAP) run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down-to 0

goose-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		$(WRAP) run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down-to $(word 2, $(MAKECMDGOALS)); \
	fi

goose-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		$(WRAP) run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down-to $(word 2, $(MAKECMDGOALS)); \
		$(WRAP) run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" up; \
	fi

# Dumps the running app DB with the container's own MARIADB_* credentials into dumps/
# Plain docker exec on purpose: bws run re-parses the command and breaks the quoted sh -c
dump-db:
	@mkdir -p dumps
	@f=dumps/liquiswiss-$$(date +%Y%m%d-%H%M%S).sql; \
	docker exec $$(docker ps -qf label=com.docker.compose.service=database-app) sh -c 'mariadb-dump -u"$$MARIADB_USER" -p"$$MARIADB_PASSWORD" --single-transaction --routines --triggers "$$MARIADB_DATABASE"' > $$f \
		&& echo "Dump written to $$f ($$(du -h $$f | cut -f1))" \
		|| { rm -f $$f; echo "Dump failed" >&2; exit 1; }

# Interactive password rotation. Prompts hide input. Pipes SQL via stdin to avoid
# quoting issues through bws run / docker compose exec.
# After running, update matching BWSM secrets in web vault, then `make deploy`.
MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

rotate-db-app:
	@bash $(MAKEFILE_DIR)rotate-db.sh app

rotate-db-wp:
	@bash $(MAKEFILE_DIR)rotate-db.sh wp
