# Order-Payment Microservices

## Repositories
- Proto files: https://github.com/nureeek/Protos-order-payment-grpc
- Generated code: https://github.com/nureeek/Generated-order-payment-grpc


## Architecture Diagram
![Architecture](assignment3_architecture%20(2).png)

### Communication Flow
- Client → Order Service: REST HTTP (:8080)
- Order Service → Payment Service: gRPC (:9091)
- stream-client → Order Service: gRPC Server-side Streaming (:9090)
- Protos repo → Generated repo: GitHub Actions auto-generation


### Test Streaming
```bash
cd orderService
go run cmd/stream-client/main.go <order_id>
```


# Assignment 3

## Architecture
- Order Service (REST :8080, gRPC :9090)
- Payment Service (REST :8081, gRPC :9091) — publishes events to RabbitMQ
- Notification Service — consumes events from RabbitMQ
- RabbitMQ — message broker on :5672 (management UI :15672)

## Event Flow
1. Client sends POST /orders to Order Service
2. Order Service calls Payment Service via gRPC
3. Payment Service processes payment and publishes event to queue `payment.completed`
4. Notification Service receives event and logs email notification



### Start RabbitMQ
### Payment Service
```bash
cd paymentService
go run cmd/payment-service/main.go
```

### Order Service
```bash
cd orderService
go run cmd/order-service/main.go
```

### Notification Service
```bash
cd notificationService
go run cmd/notification-service/main.go
```

## Idempotency Strategy
Notification Service uses `sync.Map` to store processed `payment_id` values.
Before processing each message, it checks if the ID was already handled.
If duplicate — acknowledges and skips without logging twice.

## ACK Logic
- Auto-ACK is disabled (`autoAck: false`)
- Message is acknowledged only AFTER successful log print
- If service crashes before ACK — RabbitMQ redelivers the message
- If processing fails — message is Nacked and dropped (not requeued)
- Queue is declared as `durable: true` — messages survive broker restart

## Environment Variables
| Variable | Default |
|---|---|
| AMQP_URL | amqp://guest:guest@localhost:5672/ |
| PAYMENT_DB_URL | postgres://...payment_db |
| ORDER_DB_URL | postgres://...order_db |
| GRPC_PORT | 9091 |
| HTTP_PORT | 8080/8081 |