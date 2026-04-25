FROM golang:1.26.2-alpine AS dev

RUN apk add --no-cache git curl lsof
RUN go install github.com/air-verse/air@latest
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.18.0

WORKDIR /app

CMD ["air"]

FROM golang:1.26.2-alpine AS build

WORKDIR /app
COPY ./backend /app
RUN go get
RUN go build

CMD ["./liquiswiss"]

FROM golang:1.26.2-alpine AS prod

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.18.0
COPY --from=build /app/liquiswiss /bin/liquiswiss

CMD ["/bin/liquiswiss"]
