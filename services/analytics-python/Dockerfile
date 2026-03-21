FROM golang:1.25-alpine3.20 AS builder

# Установка системных зависимостей
RUN apk add --no-cache protoc protobuf-dev tzdata ca-certificates

# Установка плагинов генерации
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /app

# Копируем модули
COPY services/engine-go/go.mod services/engine-go/go.sum ./
RUN go mod download

# Копируем остальной код
COPY services/engine-go/ .

# Генерация кода из прото-файла
RUN mkdir -p internal/gen/pb && \
    protoc --go_out=internal/gen/pb --go_opt=paths=source_relative \
    --go-grpc_out=internal/gen/pb --go-grpc_opt=paths=source_relative \
    analytics.proto

# Сборка
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /engine cmd/engine/main.go

FROM scratch
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /engine /engine

ENTRYPOINT ["/engine"]