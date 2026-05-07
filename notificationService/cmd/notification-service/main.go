package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"notificationService/internal/consumer"
)

func main() {
	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	c, err := consumer.NewConsumer(amqpURL, redisURL)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Signal received, shutting down...")
		cancel()
	}()

	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
