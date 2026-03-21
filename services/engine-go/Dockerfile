FROM golang:1.26.1-alpine3.23 AS builder

RUN apk add --no-cache tzdata ca-certificates

WORKDIR /build

COPY services/engine-go/go.mod services/engine-go/go.sum ./

RUN go mod download

COPY . .

COPY services/engine-go/ ./services/engine-go/

WORKDIR /build/services/engine-go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o engine cmd/engine/main.go

FROM scratch

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /build/services/engine-go/engine .

EXPOSE 50036

CMD ["./engine"]