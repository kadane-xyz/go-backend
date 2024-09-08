FROM golang:1.22 AS builder

COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o server src/cmd/main.go

FROM scratch

COPY --from=builder /go/server /server

EXPOSE 8080

CMD ["/server"]