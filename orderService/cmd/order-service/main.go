package main

import (
	"log"
	"net"
	"orderService/internal/infrastructure"
	"os"

	"database/sql"
	"orderService/internal/app"
	"orderService/internal/repository"
	ordergrpc "orderService/internal/transport/grpc"
	httpHandler "orderService/internal/transport/http"
	"orderService/internal/usecase"

	_ "github.com/lib/pq"
	orderpb "github.com/nureeek/Generated-order-payment-grpc/order"
	pb "github.com/nureeek/Generated-order-payment-grpc/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	dbURL := os.Getenv("ORDER_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:kometa0707@localhost:5432/order_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("cannot connect to order_db:", err)
	}

	grpcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:9091"
	}

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed to connect to payment service:", err)
	}
	defer conn.Close()

	paymentClient := pb.NewPaymentServiceClient(conn)

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	cache := infrastructure.NewOrderCache(redisURL)
	repo := repository.NewOrderRepository(db)
	uc := usecase.NewOrderUseCase(repo, paymentClient, cache)
	handler := httpHandler.NewHandler(uc)
	router := app.NewRouter(handler, cache.Client())
	
	orderGrpcPort := os.Getenv("ORDER_GRPC_PORT")
	if orderGrpcPort == "" {
		orderGrpcPort = "9090"
	}

	lis, err := net.Listen("tcp", ":"+orderGrpcPort)
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, ordergrpc.NewOrderGRPCHandler(uc))

	go func() {
		log.Println("Order gRPC server started on :" + orderGrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	router.Run(":" + httpPort)
}
