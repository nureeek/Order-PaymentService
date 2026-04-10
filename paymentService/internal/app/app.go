package app

import (
	"github.com/gin-gonic/gin"
	httpHandler "paymentService/internal/transport/http"
)

func NewRouter(handler *httpHandler.Handler) *gin.Engine {
	r := gin.Default()

	r.POST("/payments", handler.ProcessPayment)
	r.GET("/payments/:order_id", handler.GetPayment)

	return r
}
