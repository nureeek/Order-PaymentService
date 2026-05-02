package usecase

import (
	"context"
	"errors"

	"paymentService/internal/domain"
	"paymentService/internal/repository"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo      repository.PaymentRepository
	publisher Publisher
}

type Publisher interface {
	Publish(ctx context.Context, event PaymentEvent) error
}

type PaymentEvent struct {
	PaymentID     string
	OrderID       string
	Amount        float64
	CustomerEmail string
	Status        string
}

func NewPaymentUseCase(repo repository.PaymentRepository, publisher Publisher) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, publisher: publisher}
}
func (uc *PaymentUseCase) ProcessPayment(orderID string, amount int64) (*domain.Payment, error) {

	if amount > 100000 {
		return nil, errors.New("declined")
	}

	payment := &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        "Authorized",
	}

	err := uc.repo.Create(payment)
	if err != nil {
		return nil, err
	}
	if uc.publisher != nil {
		uc.publisher.Publish(context.Background(), PaymentEvent{
			PaymentID:     payment.ID,
			OrderID:       payment.OrderID,
			Amount:        float64(payment.Amount),
			CustomerEmail: "user@example.com",
			Status:        payment.Status,
		})
	}
	return payment, nil

}

func (uc *PaymentUseCase) GetPaymentByOrderID(orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}

func (uc *PaymentUseCase) ListPayments(min, max int64) ([]*domain.Payment, error) {
	if min > 0 && max > 0 && min > max {
		return nil, errors.New("min_amount cannot be greater than max_amount")
	}
	return uc.repo.FindByAmountRange(min, max)
}
