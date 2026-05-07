package consumer

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"notificationService/internal/provider"
)

type PaymentEvent struct {
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
	PaymentID     string  `json:"payment_id"`
}

type Consumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	processed   sync.Map
	redisClient *redis.Client
	emailSender provider.EmailSender
}

func NewConsumer(amqpURL string, redisURL string) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare("payment.dlq", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		"payment.completed", true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "payment.dlq",
			"x-message-ttl":             int32(30000),
			"x-max-delivery-count":      int32(3),
		},
	)
	if err != nil {
		return nil, err
	}

	ch.Qos(1, 0, false)

	opt, _ := redis.ParseURL(redisURL)
	redisClient := redis.NewClient(opt)

	mode := os.Getenv("PROVIDER_MODE")
	var sender provider.EmailSender
	sender = provider.NewSimulatedProvider()
	if mode == "REAL" {
		log.Println("Using REAL provider (not implemented, falling back to simulated)")
	}

	return &Consumer{
		conn:        conn,
		channel:     ch,
		redisClient: redisClient,
		emailSender: sender,
	}, nil
}

func (c *Consumer) sendWithRetry(event PaymentEvent) error {
	maxRetries := 4
	for i := 0; i < maxRetries; i++ {
		err := c.emailSender.Send(event.CustomerEmail, event.OrderID, event.Amount)
		if err == nil {
			return nil
		}
		if i == maxRetries-1 {
			return err
		}
		wait := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Printf("Retry %d after %s: %v", i+1, wait, err)
		time.Sleep(wait)
	}
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume("payment.completed", "", false, false, false, false, nil)
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

			key := "notif:" + event.PaymentID
			exists, _ := c.redisClient.Exists(ctx, key).Result()
			if exists > 0 {
				log.Printf("Duplicate payment_id: %s, skipping", event.PaymentID)
				msg.Ack(false)
				continue
			}

			if err := c.sendWithRetry(event); err != nil {
				log.Printf("Failed after retries: %v", err)
				msg.Nack(false, false)
				continue
			}

			c.redisClient.Set(ctx, key, "processed", 24*time.Hour)
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
