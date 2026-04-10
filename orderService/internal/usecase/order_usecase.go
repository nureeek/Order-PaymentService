package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"orderService/internal/domain"
	"orderService/internal/repository"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo       repository.OrderRepository
	httpClient *http.Client
	paymentURL string
}

func NewOrderUseCase(repo repository.OrderRepository, client *http.Client, paymentURL string) *OrderUseCase {
	return &OrderUseCase{
		repo:       repo,
		httpClient: client,
		paymentURL: paymentURL,
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

	payload, _ := json.Marshal(map[string]interface{}{
		"order_id": order.ID,
		"amount":   order.Amount,
	})

	req, _ := http.NewRequest(http.MethodPost, uc.paymentURL+"/payments", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := uc.httpClient.Do(req)
	if err != nil {
		uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
		return order, errors.New("payment service unavailable")
	}
	defer resp.Body.Close()

	var paymentResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		uc.repo.UpdateStatus(order.ID, "Failed")
		order.Status = "Failed"
		return order, err
	}

	if status, ok := paymentResp["status"].(string); ok && status == "Authorized" {
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
