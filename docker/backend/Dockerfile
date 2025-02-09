FROM golang:1.23 AS builder

WORKDIR /app

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN sqlc generate -f src/sql/sqlc.yaml && \
    CGO_ENABLED=0 go build -ldflags "-s -w" -o server src/cmd/main.go

FROM debian:12-slim 

RUN apt-get update && apt-get install -y ca-certificates curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /server

EXPOSE 80

CMD ["/server"]