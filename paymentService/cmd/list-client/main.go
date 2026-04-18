package main

import (
	"context"
	"log"

	pb "github.com/nureeek/Generated-order-payment-grpc/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:9091", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)

	resp, err := client.ListPayments(context.Background(), &pb.ListPaymentsRequest{
		MinAmount: 100,
		MaxAmount: 50000,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range resp.Payments {
		log.Printf("payment_id: %s status: %s amount: %d", p.PaymentId, p.Status, p.Amount)
	}
}
