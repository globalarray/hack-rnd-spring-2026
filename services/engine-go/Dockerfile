FROM golang:1.26.1-alpine3.23 AS builder

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /build/services/engine-go

COPY services/engine-go/go.mod services/engine-go/go.sum ./

RUN go mod download

COPY services/engine-go/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/engine ./cmd/engine/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/engine ./engine

EXPOSE 50036

CMD ["./engine"]
