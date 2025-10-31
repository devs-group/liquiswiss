FROM golang:1.25.3-alpine AS build

WORKDIR /app
COPY ./backend /app
RUN go get
RUN go build

CMD ["./liquiswiss"]

FROM golang:1.25.3-alpine AS prod

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.18.0
COPY --from=build /app/liquiswiss /bin/liquiswiss

CMD ["/bin/liquiswiss"]
