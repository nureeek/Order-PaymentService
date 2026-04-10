# Order Service 📦

A robust microservice for managing customer orders, built with **Go (Golang)** using the **Gin** framework and **PostgreSQL**. This service handles the full order lifecycle, including idempotency checks, payment service integration, and dynamic filtering.

## 🚀 Key Features

* **Order Management**: Create, retrieve, and cancel orders.
* **Idempotency Support**: Prevents duplicate orders using an `Idempotency-Key` header.
* **Payment Integration**: Communicates with external payment gateways to authorize transactions.
* **Dynamic Filtering**: Fetch order lists with optional `min_amount` and `max_amount` query parameters.
* **Clean Architecture**: Separation of concerns between Transport (HTTP), UseCase (Business Logic), and Repository (Data Access).

## 🛠 Tech Stack

* **Language:** Go 1.21+
* **Framework:** [Gin Gonic](https://github.com/gin-gonic/gin)
* **Database:** PostgreSQL
* **ID Generation:** `google/uuid`
* **Driver:** `database/sql` (standard library)

---

## 📂 Project Structure

```text
orderService/
├── cmd/
│   └── order-service/      # Main entry point
├── internal/
│   ├── app/                # Router and application setup
│   ├── domain/             # Domain models (Order, etc.)
│   ├── usecase/            # Business logic layers
│   ├── repository/         # Data persistence (PostgreSQL)
│   └── transport/
│       └── http/           # HTTP handlers and controllers
└── migrations/             # SQL schema files
```

---

## 🔌 API Endpoints

### 1. Create an Order
**POST** `/orders`
* **Required Header:** `X-Idempotency-Key` (to prevent duplicate processing)
* **Body:**
    ```json
    {
      "customer_id": "99",
      "item_name": "Mechanical Keyboard",
      "amount": 12500
    }
    ```

### 2. Get Order by ID
**GET** `/orders/:id`

### 3. List Orders (with Filters)
**GET** `/orders?min_amount=1000&max_amount=50000`
* *Both parameters are optional.*
* Returns an empty list `[]` instead of `null` if no records found.

### 4. Cancel Order
**PATCH** `/orders/:id/cancel`
* *Only "Pending" orders can be cancelled.*

---

## ⚙️ Installation & Running

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/your-username/order-service.git
    cd order-service
    ```

2.  **Database Setup**:
    Ensure your PostgreSQL instance is running and create the `orders` table using the provided schema in your repository.

3.  **Run the application**:
    ```bash
    go run cmd/order-service/main.go
    ```
    *The server will start on port `8080` by default.*

---

## 🧪 Testing with cURL

**Fetch all orders between 5,000 and 20,000:**
```bash
curl "http://localhost:8080/orders?min_amount=5000&max_amount=20000"
```

**Retrieve a specific order:**
```bash
curl "http://localhost:8080/orders/your-uuid-here"
```

