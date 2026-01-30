package main

import (
	"log"

	"github.com/JuniorCrafter/fooddelivery/internal/notifications/service"
	"github.com/JuniorCrafter/fooddelivery/internal/platform/config"
)

func main() {
	// 1. Загружаем конфиг
	cfg := config.Load()

	// 2. Инициализируем потребителя (Consumer), используя данные из cfg
	// Теперь мы РЕАЛЬНО используем переменную cfg, и Go доволен
	consumer, err := service.NewConsumer(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к RabbitMQ по адресу %s: %v", cfg.RabbitMQURL, err)
	}

	log.Println("Notification Service успешно запущен...")

	// 3. Начинаем слушать очередь
	consumer.Listen()
}
