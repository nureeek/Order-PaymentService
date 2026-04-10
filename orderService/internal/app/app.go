package app

import (
	"github.com/gin-gonic/gin"
	httpHandler "orderService/internal/transport/http"
)

func NewRouter(handler *httpHandler.Handler) *gin.Engine {
	r := gin.Default()

	r.POST("/orders", handler.CreateOrder)
	r.GET("/orders/revenue", handler.GetRevenue)
	r.GET("/orders/:id", handler.GetOrder)
	r.PATCH("/orders/:id/cancel", handler.CancelOrder)

	return r
}
