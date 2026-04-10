package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"paymentService/internal/app"
	"paymentService/internal/repository"
	httpHandler "paymentService/internal/transport/http"
	"paymentService/internal/usecase"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:kometa0707@localhost:5432/payment_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("cannot connect to payment_db:", err)
	}

	repo := repository.NewPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := httpHandler.NewHandler(uc)
	router := app.NewRouter(handler)

	router.Run(":8081")
}
