goose-static-up:
	goose --dir internal/db/migrations/static mysql liquiswiss:password@/liquiswiss up
goose-dynamic-up:
	goose -no-versioning --dir internal/db/migrations/dynamic mysql liquiswiss:password@/liquiswiss up
goose-static-up-1:
	goose --dir internal/db/migrations/static mysql liquiswiss:password@/liquiswiss up-by-one
goose-dynamic-up-1:
	goose -no-versioning --dir internal/db/migrations/dynamic mysql liquiswiss:password@/liquiswiss up-by-one
goose-static-down:
	goose --dir internal/db/migrations/static mysql liquiswiss:password@/liquiswiss down
goose-dynamic-down:
	goose -no-versioning --dir internal/db/migrations/dynamic mysql liquiswiss:password@/liquiswiss down
goose-static-down-to-0:
	goose --dir internal/db/migrations/static mysql liquiswiss:password@/liquiswiss down-to 0
goose-dynamic-down-to-0:
	goose -no-versioning --dir internal/db/migrations/dynamic mysql liquiswiss:password@/liquiswiss down-to 0
goose-static-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		goose --dir internal/db/migrations/static mysql liquiswiss:password@/liquiswiss down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-dynamic-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		goose -no-versioning --dir internal/db/migrations/dynamic mysql liquiswiss:password@/liquiswiss down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-static-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		goose --dir internal/db/migrations/static mysql liquiswiss:password@/liquiswiss down-to $(word 2, $(MAKECMDGOALS)); \
		goose --dir internal/db/migrations/static mysql liquiswiss:password@/liquiswiss up; \
	fi
goose-dynamic-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		goose -no-versioning --dir internal/db/migrations/dynamic mysql liquiswiss:password@/liquiswiss down-to $(word 2, $(MAKECMDGOALS)); \
		goose -no-versioning --dir internal/db/migrations/dynamic mysql liquiswiss:password@/liquiswiss up; \
	fi
