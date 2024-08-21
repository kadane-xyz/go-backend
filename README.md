# Kadane.XYZ Go-Backend

## Development:

Install [golang](https://go.dev/doc/install) binaries

Create keys for signing JWT
```bash
openssl genpkey -algorithm ed25519 -out ed25519-private.pem
openssl pkey -in ed25519-private.pem -pubout -out ed25519-public.pem
```

Configure Environment Variables

``.env``

Initialize environment
```bash
go mod tidy
```

Install air
```bash
go install github.com/air-verse/air@latest
```

Run go binary using air
```bash
air
```