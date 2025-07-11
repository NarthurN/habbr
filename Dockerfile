# Dockerfile для Habbr GraphQL API
# Многоэтапная сборка для оптимизации размера образа

# Этап 1: Сборка приложения
FROM golang:1.23-alpine AS builder

# Установка необходимых инструментов
RUN apk add --no-cache git ca-certificates tzdata

# Создание пользователя для безопасности
RUN adduser -D -g '' appuser

# Установка рабочей директории
WORKDIR /build

# Копирование файлов зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей (кэшируется при изменении только кода)
RUN go mod download
RUN go mod verify

# Копирование исходного кода
COPY . .

# Сборка приложения с оптимизацией
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o habbr \
    ./cmd/server

# Этап 2: Тестирование (опционально)
FROM builder AS tester
RUN go test -v ./...

# Этап 3: Производственный образ
FROM scratch AS production

# Импорт пользователя и группы из builder
COPY --from=builder /etc/passwd /etc/passwd

# Импорт CA сертификатов
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Импорт временных зон
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копирование бинарного файла
COPY --from=builder /build/habbr /habbr

# Использование непривилегированного пользователя
USER appuser

# Открытие порта
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/habbr", "health"] || exit 1

# Запуск приложения
ENTRYPOINT ["/habbr"]

# Этап 4: Образ для разработки
FROM golang:1.23-alpine AS development

# Установка дополнительных инструментов для разработки
RUN apk add --no-cache \
    git \
    curl \
    postgresql-client \
    redis \
    ca-certificates

# Установка air для hot reload
RUN go install github.com/cosmtrek/air@latest

# Установка migrate для работы с миграциями
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Установка gqlgen
RUN go install github.com/99designs/gqlgen@latest

# Создание рабочей директории
WORKDIR /app

# Копирование файлов проекта
COPY . .

# Загрузка зависимостей
RUN go mod download

# Сборка для разработки
RUN go build -o habbr ./cmd/server

# Открытие порта
EXPOSE 8080

# Команда по умолчанию для разработки
CMD ["air", "-c", ".air.toml"]

# Этап 5: Образ для отладки
FROM alpine:latest AS debug

# Установка инструментов для отладки
RUN apk add --no-cache \
    ca-certificates \
    curl \
    netcat-openbsd \
    postgresql-client \
    redis \
    htop \
    strace

# Создание пользователя
RUN adduser -D -g '' appuser

# Копирование бинарного файла
COPY --from=builder /build/habbr /habbr

# Создание директории для логов
RUN mkdir -p /var/log/habbr && chown appuser:appuser /var/log/habbr

# Использование непривилегированного пользователя
USER appuser

# Открытие порта
EXPOSE 8080

# Команда по умолчанию
CMD ["/habbr"]
