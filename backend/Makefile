DB_DSN ?= liquiswiss:password@tcp(127.0.0.1:3317)/liquiswiss
STATIC_DIR := internal/db/migrations/static
DYNAMIC_DIR := internal/db/migrations/dynamic
GOOSE_STATIC := goose --dir $(STATIC_DIR) mysql "$(DB_DSN)"
GOOSE_DYNAMIC := goose -no-versioning --dir $(DYNAMIC_DIR) mysql "$(DB_DSN)"

modernize:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix ./...

goose-static-create:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-static-create <type_of_creation>_<name_of_migration> (eg. create_table_my_table)"; \
	else \
		goose -s --dir $(STATIC_DIR) create $(word 2, $(MAKECMDGOALS)) sql; \
	fi
goose-dynamic-create:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-dynamic-create <type_of_creation>_<name_of_migration> (eg. create_function_calc_sum)"; \
	else \
		goose -s --dir $(DYNAMIC_DIR) create $(word 2, $(MAKECMDGOALS)) sql; \
	fi

goose-static-up:         ; $(GOOSE_STATIC) up
goose-dynamic-up:        ; $(GOOSE_DYNAMIC) up
goose-static-up-1:       ; $(GOOSE_STATIC) up-by-one
goose-dynamic-up-1:      ; $(GOOSE_DYNAMIC) up-by-one
goose-static-down:       ; $(GOOSE_STATIC) down
goose-dynamic-down:      ; $(GOOSE_DYNAMIC) down
goose-static-down-to-0:  ; $(GOOSE_STATIC) down-to 0
goose-dynamic-down-to-0: ; $(GOOSE_DYNAMIC) down-to 0

goose-static-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-static-down-to <version_number>"; \
	else \
		$(GOOSE_STATIC) down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-dynamic-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-dynamic-down-to <version_number>"; \
	else \
		$(GOOSE_DYNAMIC) down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-static-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-static-down-to-and-up <version_number>"; \
	else \
		$(GOOSE_STATIC) down-to $(word 2, $(MAKECMDGOALS)) && $(GOOSE_STATIC) up; \
	fi
goose-dynamic-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-dynamic-down-to-and-up <version_number>"; \
	else \
		$(GOOSE_DYNAMIC) down-to $(word 2, $(MAKECMDGOALS)) && $(GOOSE_DYNAMIC) up; \
	fi
%:
	@:
