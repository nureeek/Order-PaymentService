package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"orderService/internal/domain"
	"orderService/internal/repository"

	"github.com/google/uuid"
	pb "github.com/nureeek/Generated-order-payment-grpc/payment"
)

type Cache interface {
	Get(ctx context.Context, orderID string) (*domain.Order, error)
	Set(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, orderID string) error
}
type OrderUseCase struct {
	repo          repository.OrderRepository
	paymentClient pb.PaymentServiceClient
	cache         Cache
}

func NewOrderUseCase(repo repository.OrderRepository, paymentClient pb.PaymentServiceClient, cache Cache) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
		cache:         cache,
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
		if uc.cache != nil {
			uc.cache.Delete(context.Background(), order.ID)
		}
		order.Status = "Paid"
	} else {
		uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	if uc.cache != nil {
		if order, err := uc.cache.Get(context.Background(), id); err == nil {
			return order, nil
		}
		log.Println("Cache MISS for order:", id)
	}
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if uc.cache != nil {
		uc.cache.Set(context.Background(), order)
	}
	return order, nil
}

func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}
	if order.Status != "Pending" {
		return errors.New("only pending orders can be cancelled")
	}
	if err := uc.repo.UpdateStatus(id, "Cancelled"); err != nil {
		return err
	}
	if uc.cache != nil {
		uc.cache.Delete(context.Background(), id)
	}
	return nil
}

func (uc *OrderUseCase) GetRevenue(customerID string) (*repository.CustomerRevenue, error) {
	return uc.repo.GetRevenueByCustomerID(customerID)
}
