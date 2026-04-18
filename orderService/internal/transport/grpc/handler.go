package grpc

import (
	"time"

	pb "github.com/nureeek/Generated-order-payment-grpc/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"orderService/internal/usecase"
)

type OrderGRPCHandler struct {
	pb.UnimplementedOrderServiceServer
	uc *usecase.OrderUseCase
}

func NewOrderGRPCHandler(uc *usecase.OrderUseCase) *OrderGRPCHandler {
	return &OrderGRPCHandler{uc: uc}
}

func (h *OrderGRPCHandler) SubscribeToOrderUpdates(req *pb.OrderRequest, stream pb.OrderService_SubscribeToOrderUpdatesServer) error {
	lastStatus := ""

	for {
		order, err := h.uc.GetOrder(req.OrderId)
		if err != nil {
			return status.Errorf(codes.NotFound, "order not found: %v", err)
		}

		if order.Status != lastStatus {
			lastStatus = order.Status
			err := stream.Send(&pb.OrderStatusUpdate{
				OrderId:   order.ID,
				Status:    order.Status,
				UpdatedAt: timestamppb.Now(),
			})
			if err != nil {
				return err
			}
		}

		time.Sleep(2 * time.Second)
	}
}
