package app

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	httpHandler "orderService/internal/transport/http"
)

func NewRouter(handler *httpHandler.Handler, redisClient *redis.Client) *gin.Engine {
	r := gin.Default()
	r.Use(httpHandler.RateLimiter(redisClient))
	
	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/revenue", handler.GetRevenue)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)

	return r
}
