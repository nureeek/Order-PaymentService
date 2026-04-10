package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"orderService/internal/app"
	"orderService/internal/repository"
	httpHandler "orderService/internal/transport/http"
	"orderService/internal/usecase"
)

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:kometa0707@localhost:5432/order_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("cannot connect to order_db:", err)
	}

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	repo := repository.NewOrderRepository(db)
	uc := usecase.NewOrderUseCase(repo, httpClient, "http://localhost:8081")
	handler := httpHandler.NewHandler(uc)
	router := app.NewRouter(handler)

	router.Run(":8080")
}
