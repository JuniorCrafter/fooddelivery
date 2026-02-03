# Этап 1: Сборка (используем легкий образ Go)
FROM golang:1.25-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код проекта
COPY . .

# Аргумент, который скажет, какой именно сервис собирать
ARG SERVICE_PATH

# Собираем бинарный файл
RUN go build -o service ${SERVICE_PATH}

# Этап 2: Финальный образ (минимальный размер)
FROM alpine:latest
WORKDIR /root/

# Копируем только собранный файл из первого этапа
COPY --from=builder /app/service .

# Запускаем сервис
CMD ["./service"]
