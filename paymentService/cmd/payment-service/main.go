package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"paymentService/internal/app"
	infrastructure "paymentService/internal/inftastructure"
	"paymentService/internal/repository"
	grpcHandler "paymentService/internal/transport/grpc"
	httpHandler "paymentService/internal/transport/http"
	"paymentService/internal/usecase"
	"time"

	_ "github.com/lib/pq"
	pb "github.com/nureeek/Generated-order-payment-grpc/payment"
	"google.golang.org/grpc"
)

func main() {
	dbURL := os.Getenv("PAYMENT_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:kometa0707@localhost:5432/payment_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("cannot connect to payment_db:", err)
	}

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	pub, err := infrastructure.NewPublisher(amqpURL)
	if err != nil {
		log.Printf("Warning: could not connect to RabbitMQ: %v", err)
	}

	repo := repository.NewPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo, pub)
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9091"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			start := time.Now()
			resp, err := handler(ctx, req)
			log.Printf("gRPC method: %s duration: %s", info.FullMethod, time.Since(start))
			return resp, err
		}),
	)
	pb.RegisterPaymentServiceServer(grpcServer, grpcHandler.NewPaymentGRPCHandler(uc))

	go func() {
		log.Println("gRPC server started on :" + grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("gRPC serve error:", err)
		}
	}()

	handler := httpHandler.NewHandler(uc)
	router := app.NewRouter(handler)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	log.Println("HTTP server started on :" + httpPort)
	router.Run(":" + httpPort)
}
