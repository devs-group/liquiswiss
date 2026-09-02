FROM golang:1.26.2-alpine AS dev

RUN apk add --no-cache git curl lsof
RUN go install github.com/air-verse/air@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.18.0

WORKDIR /app

CMD ["air"]

FROM golang:1.26.2-alpine AS build

# Build identifiers stamped into the binary and served by GET /api/config.
ARG GIT_SHA=dev
ARG VERSION=dev
ARG APP_SHA=

WORKDIR /app
COPY ./backend /app
RUN go get
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w \
      -X liquiswiss/pkg/buildinfo.GitSHA=${GIT_SHA} \
      -X liquiswiss/pkg/buildinfo.Version=${VERSION} \
      -X liquiswiss/pkg/buildinfo.AppSHA=${APP_SHA}" \
    -o liquiswiss

CMD ["./liquiswiss"]

FROM golang:1.26.2-alpine AS prod

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.18.0
COPY --from=build /app/liquiswiss /bin/liquiswiss

# Probes its own /api/health; also what the deploy webhook waits on.
HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=5 \
  CMD ["/bin/liquiswiss", "healthcheck"]

CMD ["/bin/liquiswiss"]
