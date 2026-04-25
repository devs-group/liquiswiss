DB_DSN ?= $(APP_DB_USER):$(APP_DB_PASSWORD)@tcp($(APP_DB_HOST):$(APP_DB_PORT))/$(APP_DB_NAME)

goose-up:
	docker compose run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" up

goose-down:
	docker compose run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down

goose-down-to-0:
	docker compose run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down-to 0

goose-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		docker compose run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down-to $(word 2, $(MAKECMDGOALS)); \
	fi

goose-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		docker compose run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" down-to $(word 2, $(MAKECMDGOALS)); \
		docker compose run --rm backend goose --dir internal/db/migrations mysql "$(DB_DSN)" up; \
	fi

dump-db:
	docker compose run --rm database-app mysqldump -u $(APP_DB_USER) -p$(APP_DB_PASSWORD) $(APP_DB_NAME) > dumpfile.sql
