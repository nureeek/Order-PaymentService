package consumer

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentEvent struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	PaymentID     string  `json:"payment_id"`
}

type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	processed sync.Map
}

func NewConsumer(amqpURL string) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		"payment.completed",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return nil, err
	}

	return &Consumer{conn: conn, channel: ch}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		"payment.completed",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Notification Service started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer...")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			var event PaymentEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Failed to parse message: %v", err)
				msg.Nack(false, false)
				continue
			}

			if _, exists := c.processed.Load(event.PaymentID); exists {
				log.Printf("Duplicate message for payment_id: %s, skipping", event.PaymentID)
				msg.Ack(false)
				continue
			}

			log.Printf("[Notification] Sent email to %s for Order #%s. Amount: %.2f Status: %s",
				event.CustomerEmail, event.OrderID, event.Amount, event.Status)

			c.processed.Store(event.PaymentID, true)
			msg.Ack(false)
		}
	}
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
