gobuild:
	cd cmd/liquiswiss && go build
goose-up:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss up
	echo "RUN THE mocks.sql as well"
goose-up-1:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss up-by-one
goose-down:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss down
goose-down-to-0:
	goose --dir internal/db/migrations mysql liquiswiss:password@/liquiswiss down-to 0