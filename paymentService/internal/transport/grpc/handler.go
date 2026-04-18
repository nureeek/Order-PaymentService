package grpc

import (
	"context"
	"paymentService/internal/usecase"

	pb "github.com/nureeek/Generated-order-payment-grpc/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentGRPCHandler struct {
	pb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentGRPCHandler(uc *usecase.PaymentUseCase) *PaymentGRPCHandler {
	return &PaymentGRPCHandler{uc: uc}
}

func (h *PaymentGRPCHandler) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	payment, err := h.uc.ProcessPayment(req.OrderId, int64(req.Amount))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "payment declined: %v", err)
	}

	return &pb.PaymentResponse{
		PaymentId: payment.ID,
		Status:    payment.Status,
		CreatedAt: timestamppb.Now(),
	}, nil
}
func (h *PaymentGRPCHandler) ListPayments(ctx context.Context, req *pb.ListPaymentsRequest) (*pb.ListPaymentsResponse, error) {
	payments, err := h.uc.ListPayments(req.MinAmount, req.MaxAmount)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid range: %v", err)
	}

	var result []*pb.PaymentResponse
	for _, p := range payments {
		result = append(result, &pb.PaymentResponse{
			PaymentId: p.ID,
			Status:    p.Status,
			CreatedAt: timestamppb.Now(),
			Amount:    p.Amount,
		})
	}

	return &pb.ListPaymentsResponse{Payments: result}, nil
}
