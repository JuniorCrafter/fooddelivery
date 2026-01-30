package service

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type NotificationConsumer struct {
	conn *amqp091.Connection
}

func NewConsumer(url string) (*NotificationConsumer, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	return &NotificationConsumer{conn: conn}, nil
}

// Listen начинает слушать очередь "order_status_updates"
func (c *NotificationConsumer) Listen() {
	ch, _ := c.conn.Channel()
	defer ch.Close()

	// Объявляем очередь (если её нет, RabbitMQ её создаст)
	q, _ := ch.QueueDeclare("order_status_updates", true, false, false, false, nil)

	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for d := range msgs {
			// Имитация отправки уведомления
			log.Printf(" Получено событие: %s. Отправляем Push клиенту...", d.Body)
		}
	}()

	log.Println("Notification Service ждет сообщений...")
	select {} // Вечный цикл, чтобы сервис не закрылся
}
