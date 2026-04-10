package usecase

import (
	"context"
	"errors"
	"time"

	"orderService/internal/domain"
	"orderService/internal/repository"

	"github.com/google/uuid"
	pb "github.com/nureeek/Generated-order-payment-grpc/payment"
)

type OrderUseCase struct {
	repo          repository.OrderRepository
	paymentClient pb.PaymentServiceClient
}

func NewOrderUseCase(repo repository.OrderRepository, paymentClient pb.PaymentServiceClient) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
	}
}

func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	if idempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(idempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	order := &domain.Order{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     "Pending",
		CreatedAt:  time.Now(),
	}

	if err := uc.repo.Create(order, idempotencyKey); err != nil {
		return nil, err
	}
	paymentResp, err := uc.paymentClient.ProcessPayment(context.Background(), &pb.PaymentRequest{
		OrderId:  order.ID,
		Amount:   float64(order.Amount),
		Currency: "KZT",
	})
	if err != nil {
		uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
		return order, errors.New("payment service unavailable")
	}

	if paymentResp.Status == "Authorized" {
		uc.repo.UpdateStatus(order.ID, "Paid")
		order.Status = "Paid"
	} else {
		uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	return uc.repo.GetByID(id)
}

func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	if order.Status != "Pending" {
		return errors.New("only pending orders can be cancelled")
	}

	return uc.repo.UpdateStatus(id, "Cancelled")
}

func (uc *OrderUseCase) GetRevenue(customerID string) (*repository.CustomerRevenue, error) {
	return uc.repo.GetRevenueByCustomerID(customerID)
}
