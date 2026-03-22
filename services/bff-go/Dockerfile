FROM golang:1.25-alpine AS builder

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /build/services/bff-go
COPY services/bff-go/go.mod services/bff-go/go.sum ./
RUN go mod download

COPY services/bff-go/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/bff ./cmd/bff

FROM scratch

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /out/bff .

EXPOSE 8080

CMD ["./bff"]
