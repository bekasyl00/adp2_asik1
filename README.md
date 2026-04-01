# AP2 Assignment 1 – Clean Architecture Microservices (Order & Payment)

## Architecture Overview

This project implements a two-service platform: **Order Service** and **Payment Service**, built using Clean Architecture principles in Go with the Gin HTTP framework.

### Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                          System Architecture                                 │
│                                                                              │
│   ┌─────────────────────────────┐      REST (HTTP)     ┌───────────────────────────┐
│   │     ORDER SERVICE (:8080)   │  ──────────────────► │   PAYMENT SERVICE (:8081)  │
│   │                             │   POST /payments     │                           │
│   │  ┌───────────────────────┐  │   Timeout: 2s        │  ┌─────────────────────┐  │
│   │  │   Transport (HTTP)    │  │                      │  │  Transport (HTTP)   │  │
│   │  │   - Gin Handlers      │  │                      │  │  - Gin Handlers     │  │
│   │  │   - Payment Client    │  │                      │  │                     │  │
│   │  └───────┬───────────────┘  │                      │  └─────────┬───────────┘  │
│   │          │                  │                      │            │              │
│   │  ┌───────▼───────────────┐  │                      │  ┌─────────▼───────────┐  │
│   │  │   Use Cases           │  │                      │  │  Use Cases          │  │
│   │  │   - CreateOrder       │  │                      │  │  - AuthorizePayment │  │
│   │  │   - GetOrder          │  │                      │  │  - GetPayment       │  │
│   │  │   - CancelOrder       │  │                      │  │                     │  │
│   │  └───────┬───────────────┘  │                      │  └─────────┬───────────┘  │
│   │          │                  │                      │            │              │
│   │  ┌───────▼───────────────┐  │                      │  ┌─────────▼───────────┐  │
│   │  │   Domain              │  │                      │  │  Domain             │  │
│   │  │   - Order Entity      │  │                      │  │  - Payment Entity   │  │
│   │  │   - Ports (Interfaces)│  │                      │  │  - Ports (Interfaces│  │
│   │  └───────┬───────────────┘  │                      │  └─────────┬───────────┘  │
│   │          │                  │                      │            │              │
│   │  ┌───────▼───────────────┐  │                      │  ┌─────────▼───────────┐  │
│   │  │   Repository          │  │                      │  │  Repository         │  │
│   │  │   - PostgreSQL Impl   │  │                      │  │  - PostgreSQL Impl  │  │
│   │  └───────┬───────────────┘  │                      │  └─────────┬───────────┘  │
│   │          │                  │                      │            │              │
│   └──────────┼──────────────────┘                      └────────────┼──────────────┘
│              │                                                      │              │
│      ┌───────▼────────┐                                    ┌────────▼───────┐      │
│      │  order_db       │                                    │  payment_db    │      │
│      │  (PostgreSQL)   │                                    │  (PostgreSQL)  │      │
│      │  Port: 5432     │                                    │  Port: 5433    │      │
│      └────────────────┘                                    └────────────────┘      │
│                                                                                    │
│              ⚠  Separate databases — NO shared storage                             │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

### Dependency Flow (Clean Architecture)

```
Transport (HTTP Handlers) → Use Cases → Domain (Entities + Ports)
                                              ↑
                        Repository (Impl) ────┘  (Dependency Inversion)
```

## Bounded Contexts

### Order Service (Bounded Context: Order Management)
- **Responsibility**: Manages customer orders and their lifecycle states
- **Entity**: `Order` (ID, CustomerID, ItemName, Amount, Status, CreatedAt)
- **States**: Pending → Paid / Failed / Cancelled
- **Database**: `order_db` (PostgreSQL, port 5440)

### Payment Service (Bounded Context: Payment Processing)
- **Responsibility**: Processes payments and validates transaction limits
- **Entity**: `Payment` (ID, OrderID, TransactionID, Amount, Status)
- **States**: Authorized / Declined
- **Database**: `payment_db` (PostgreSQL, port 5441)

Each service owns its own:
- Domain models (no shared code)
- Database (database-per-service pattern)
- Business rules
- Repository implementations

## Architecture Decisions

### 1. Clean Architecture Layers
Each service follows a strict layered architecture:

| Layer | Responsibility | Dependencies |
|-------|---------------|-------------|
| **Domain** | Entities + Port interfaces | None (innermost) |
| **Use Case** | Business logic & orchestration | Domain interfaces only |
| **Repository** | Database persistence (PostgreSQL) | Domain interfaces |
| **Transport** | HTTP handlers (Gin), external clients | Use Case layer |
| **App** | Composition Root, DI wiring | All layers (outermost) |

### 2. Dependency Inversion
- Use cases depend on **interfaces** (Ports), not concrete implementations
- `OrderRepository` interface in domain, implemented by `PostgresOrderRepository`
- `PaymentClient` interface in domain, implemented by `PaymentHTTPClient`
- This makes the system testable and decoupled

### 3. Manual Dependency Injection
All dependencies are wired manually in `internal/app/app.go` (Composition Root):
- No DI containers or frameworks
- Dependencies flow: DB → Repository → UseCase → Handler → Router

### 4. Financial Accuracy
- All monetary amounts use `int64` (cents) — **never** `float64`
- Example: `$10.00` is stored as `1000` cents
- Database constraint: `amount > 0`

### 5. Database per Service
- Order Service → `order_db` (port 5432)
- Payment Service → `payment_db` (port 5433)
- No shared tables, schemas, or connections

## Failure Handling

### Payment Service Unavailable Scenario
When the Payment Service is down or unreachable:

1. **Timeout Protection**: Order Service uses `http.Client` with a **2-second timeout**
2. **No Hanging**: The timeout prevents the Order Service from waiting indefinitely
3. **Error Response**: Returns HTTP `503 Service Unavailable`
4. **Order Status**: The order is marked as **"Failed"**

**Design Decision**: We chose to mark orders as "Failed" (rather than keeping them "Pending") because:
- It provides clear feedback to the user that the payment attempt failed
- The user can create a new order when the Payment Service recovers
- Keeping orders "Pending" indefinitely could create confusion about the order's state
- A "Pending" order with a failed payment attempt might mislead users into thinking payment is still being processed

### Payment Limit Exceeded
When payment amount exceeds 100000 cents ($1000.00):
- Payment Service returns status "Declined" (not an error)
- Order Service marks the order as "Failed"
- returns HTTP `402 Payment Required`

## Project Structure

```
asik1/
├── docker-compose.yml          # Database infrastructure
├── README.md                   # This file
│
├── order-service/
│   ├── cmd/
│   │   └── order-service/
│   │       └── main.go         # Entry point (Composition Root)
│   ├── internal/
│   │   ├── domain/
│   │   │   ├── order.go        # Order entity + business rules
│   │   │   └── ports.go        # Repository & PaymentClient interfaces
│   │   ├── usecase/
│   │   │   └── order_usecase.go # Business logic orchestration
│   │   ├── repository/
│   │   │   └── postgres_order.go # PostgreSQL implementation
│   │   ├── transport/
│   │   │   └── http/
│   │   │       ├── handler.go       # Gin HTTP handlers
│   │   │       └── payment_client.go # REST client for Payment Service
│   │   └── app/
│   │       └── app.go          # Application wiring, config, DI
│   ├── migrations/
│   │   └── 001_create_orders.sql
│   ├── go.mod
│   └── go.sum
│
└── payment-service/
    ├── cmd/
    │   └── payment-service/
    │       └── main.go         # Entry point (Composition Root)
    ├── internal/
    │   ├── domain/
    │   │   ├── payment.go      # Payment entity + business rules
    │   │   └── ports.go        # Repository interface
    │   ├── usecase/
    │   │   └── payment_usecase.go # Business logic
    │   ├── repository/
    │   │   └── postgres_payment.go # PostgreSQL implementation
    │   ├── transport/
    │   │   └── http/
    │   │       └── handler.go     # Gin HTTP handlers
    │   └── app/
    │       └── app.go          # Application wiring, config, DI
    ├── migrations/
    │   └── 001_create_payments.sql
    ├── go.mod
    └── go.sum
```

## How to Run

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL client (optional, for direct DB access)

### 1. Start Databases
```bash
docker-compose up -d
```

This starts two separate PostgreSQL instances:
- `order_db` on port **5440**
- `payment_db` on port **5441**

Migrations are automatically applied on first start.

### 2. Start Payment Service (Terminal 1)
```bash
cd payment-service
go mod tidy
go run cmd/payment-service/main.go
```
Payment Service runs on **http://localhost:8081**

### 3. Start Order Service (Terminal 2)
```bash
cd order-service
go mod tidy
go run cmd/order-service/main.go
```
Order Service runs on **http://localhost:8080**

## API Examples

### Create an Order (with successful payment)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust-001", "item_name": "Laptop", "amount": 15000}'
```

**Response** (201 Created):
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": "cust-001",
  "item_name": "Laptop",
  "amount": 15000,
  "status": "Paid",
  "created_at": "2026-04-01T10:30:00Z"
}
```

### Create an Order (payment declined — exceeds limit)
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust-002", "item_name": "Luxury Watch", "amount": 150000}'
```

**Response** (402 Payment Required):
```json
{
  "id": "...",
  "customer_id": "cust-002",
  "item_name": "Luxury Watch",
  "amount": 150000,
  "status": "Failed",
  "created_at": "..."
}
```

### Get Order Details
```bash
curl http://localhost:8080/orders/{order-id}
```

### Cancel an Order
```bash
curl -X PATCH http://localhost:8080/orders/{order-id}/cancel
```

**Response** (200 OK):
```json
{
  "id": "...",
  "status": "Cancelled"
}
```

**Error** (409 Conflict — trying to cancel a paid order):
```json
{
  "error": "paid orders cannot be cancelled"
}
```

### Create a Payment Directly
```bash
curl -X POST http://localhost:8081/payments \
  -H "Content-Type: application/json" \
  -d '{"order_id": "order-123", "amount": 5000}'
```

### Get Payment by Order ID
```bash
curl http://localhost:8081/payments/{order-id}
```

### Test Failure Scenario (Payment Service Down)
1. Stop the Payment Service
2. Try creating an order:
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust-003", "item_name": "Phone", "amount": 5000}'
```

**Response** (503 Service Unavailable):
```json
{
  "error": "payment service unavailable",
  "order": {
    "id": "...",
    "status": "Failed"
  }
}
```

## Business Rules Summary

| Rule | Location | Enforcement |
|------|----------|------------|
| Amount > 0 | Domain layer (both services) | `NewOrder()` / `NewPayment()` validation |
| Amount as int64 | Domain layer | Struct field type + DB column type `BIGINT` |
| Paid orders can't be cancelled | Domain layer (Order) | `Order.Cancel()` method |
| Only Pending orders can be cancelled | Domain layer (Order) | `Order.Cancel()` method |
| Payment limit ($1000) | Domain layer (Payment) | `NewPayment()` → amount > 100000 → "Declined" |
| 2-second timeout | App layer (Order Service) | `http.Client{Timeout: 2s}` |
| 503 on Payment unavailable | Transport layer (Order) | Handler checks for `ErrPaymentUnavailable` |

## Environment Variables

### Order Service
| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5440 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | order_db | Database name |
| `SERVER_PORT` | 8080 | HTTP server port |
| `PAYMENT_BASE_URL` | http://localhost:8081 | Payment Service URL |

### Payment Service
| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5441 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | payment_db | Database name |
| `SERVER_PORT` | 8081 | HTTP server port |
