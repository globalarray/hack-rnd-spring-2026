FROM golang:1.25-alpine AS builder

WORKDIR /build/services/auth-go

COPY services/auth-go/go.mod services/auth-go/go.sum ./
RUN go mod download

COPY services/auth-go/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/auth ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/auth /app/auth
COPY services/auth-go/internal/migrations/init.sql /app/internal/migrations/init.sql

ENTRYPOINT ["/app/auth"]
