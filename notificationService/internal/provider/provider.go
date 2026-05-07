package provider

import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type EmailSender interface {
	Send(to, orderID string, amount float64) error
}

type SimulatedProvider struct{}

func NewSimulatedProvider() *SimulatedProvider {
	return &SimulatedProvider{}
}

func (p *SimulatedProvider) Send(to, orderID string, amount float64) error {
	time.Sleep(500 * time.Millisecond)
	if rand.Intn(3) == 0 {
		return errors.New("simulated provider failure")
	}
	log.Printf("[Email] Sent to %s for order %s amount %.2f", to, orderID, amount)
	return nil
}
