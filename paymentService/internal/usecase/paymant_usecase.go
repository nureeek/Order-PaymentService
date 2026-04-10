package usecase

import (
	"errors"

	"paymentService/internal/domain"
	"paymentService/internal/repository"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo repository.PaymentRepository
}

func NewPaymentUseCase(repo repository.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
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

	return payment, nil
}

func (uc *PaymentUseCase) GetPaymentByOrderID(orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}
