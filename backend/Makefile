goose-up:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss up
goose-up-1:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss up-by-one
goose-down:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss down
goose-down-to-0:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss down-to 0
goose-down-to:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to <version_number>"; \
	else \
		goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss down-to $(word 2, $(MAKECMDGOALS)); \
	fi
goose-down-to-and-up:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make goose-down-to-and-up <version_number>"; \
	else \
		goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss down-to $(word 2, $(MAKECMDGOALS)); \
		goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss up; \
	fi
