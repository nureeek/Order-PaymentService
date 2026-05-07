package infrastructure

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"orderService/internal/domain"

	"github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewOrderCache(redisURL string) *OrderCache {
	opt, _ := redis.ParseURL(redisURL)
	client := redis.NewClient(opt)
	return &OrderCache{client: client, ttl: 5 * time.Minute}
}

func (c *OrderCache) Get(ctx context.Context, orderID string) (*domain.Order, error) {
	val, err := c.client.Get(ctx, "order:"+orderID).Result()
	if err != nil {
		return nil, err
	}
	log.Println("Cache HIT for order:", orderID)
	var order domain.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (c *OrderCache) Set(ctx context.Context, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "order:"+order.ID, data, c.ttl).Err()
}

func (c *OrderCache) Delete(ctx context.Context, orderID string) error {
	return c.client.Del(ctx, "order:"+orderID).Err()
}
func (c *OrderCache) Client() *redis.Client {
	return c.client
}
