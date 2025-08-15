# Stage 1: Builder
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Установка migrate CLI с postgres-драйвером
RUN CGO_ENABLED=0 go install -tags "postgres file" github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3

# Сборка основного приложения
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/app/main.go

# Stage 2: Runtime
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /go/bin/migrate .
COPY config.yaml .

CMD ["./main"]
