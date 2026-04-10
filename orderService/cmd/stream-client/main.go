package main

import (
	"context"
	"io"
	"log"
	"os"

	pb "github.com/nureeek/Generated-order-payment-grpc/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	orderID := os.Args[1]

	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	stream, err := client.SubscribeToOrderUpdates(context.Background(), &pb.OrderRequest{OrderId: orderID})
	if err != nil {
		log.Fatal(err)
	}

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Order %s status: %s at %s", update.OrderId, update.Status, update.UpdatedAt.AsTime())
	}
}
