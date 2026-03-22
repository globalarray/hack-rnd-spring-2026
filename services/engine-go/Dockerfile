# Используем 1.23, чтобы удовлетворить требования библиотек
FROM golang:1.23-alpine3.20 AS builder

RUN apk add --no-cache protoc protobuf-dev tzdata ca-certificates

# Указываем конкретные версии, которые гарантированно работают с Go 1.23
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
RUN go install github.com/envoyproxy/protoc-gen-validate@v1.0.4

WORKDIR /app

# Сначала копируем только модули для кэширования слоев
COPY services/engine-go/go.mod services/engine-go/go.sum ./
RUN go mod download

# Копируем остальной код
COPY services/engine-go/ .
COPY proto/ /app/proto/

# Генерируем gRPC-контракт engine-go
RUN mkdir -p internal/gen/pb && \
    protoc \
    -I /app/proto \
    --go_out=paths=source_relative:internal/gen/pb \
    --go-grpc_out=paths=source_relative:internal/gen/pb \
    --validate_out=paths=source_relative:internal/gen/pb \
    /app/proto/test_engine.proto

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /engine cmd/engine/main.go

FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /engine /engine

ENTRYPOINT ["/engine"]
