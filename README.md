# AP2 Assignment 1 – Clean Architecture based Microservices (Order & Payment)

## Overview

This project implements a small two-service platform in Go using Clean Architecture and REST communication.

The system consists of:
- Order Service
- Payment Service

Each service is implemented as an independent microservice with its own database, internal models, repository layer, use case layer, and HTTP transport layer.

The project demonstrates:
- separation of concerns
- dependency inversion
- bounded contexts
- database per service
- synchronous REST communication with timeout handling

## Services

### 1. Order Service
The Order Service manages customer orders and their states.

Supported endpoints:
- `POST /orders`
- `GET /orders/{id}`
- `PATCH /orders/{id}/cancel`

Main responsibilities:
- create a new order with status `Pending`
- call Payment Service to authorize payment
- update order status to `Paid` or `Failed`
- return order details
- allow cancellation only for `Pending` orders

### 2. Payment Service
The Payment Service processes payments and stores transaction information.

Supported endpoints:
- `POST /payments`
- `GET /payments/{order_id}`

Main responsibilities:
- receive payment request from Order Service
- authorize or decline payment
- store payment result in its own database
- return payment information

## Architecture

Each service follows Clean Architecture principles.

### Layers
- domain: contains core business entities
- usecase: contains business logic and application rules
- repository: contains persistence logic and outbound infrastructure logic
- transport/http: contains thin HTTP handlers
- cmd: composition root where dependencies are wired manually

### Why this structure?
This design keeps HTTP, database code, and business logic separated.  
Handlers remain thin, business rules stay in use cases, and repositories handle database access.

## Bounded Contexts

The system is divided into two bounded contexts:

- Order Context – owns order data and order lifecycle
- Payment Context – owns payment authorization and transaction storage

There is:
- no shared database
- no shared entity/model package
- no direct SQL access from Order Service to Payment data

This makes service boundaries clear and reduces coupling.

## Databases

Each service has its own PostgreSQL database:

- order_db
- payment_db

This follows the database-per-service rule and ensures separate data ownership.

## Business Rules

The system implements the following rules:

1. Money is stored as int64
2. Order amount must be greater than 0
3. Only `Pending` orders can be cancelled
4. `Paid` orders cannot be cancelled
5. If payment amount is greater than 100000, the payment is declined
6. Order Service communicates with Payment Service using REST only
7. Order Service uses a custom `http.Client` with a timeout of 2 seconds

## Failure Handling

If Payment Service is unavailable:
- Order Service does not wait indefinitely
- the HTTP client timeout is triggered
- Order Service returns `503 Service Unavailable`
- the order is marked as `Failed`

I chose to mark the order as `Failed` in this scenario because it avoids leaving the order in an uncertain intermediate state and makes the behavior deterministic and easier to explain.

## REST Communication

Order Service calls Payment Service through:

- `POST /payments`

This is synchronous communication over HTTP using a custom client with timeout protection.

## Project Structure

### order-service
```text
order-service/
├── cmd/order-service/main.go
├── internal/domain/
├── internal/usecase/
├── internal/repository/
├── internal/transport/http/
└── migrations/
```

### payment-service
```text
payment-service/
├── cmd/payment-service/main.go
├── internal/domain/
├── internal/usecase/
├── internal/repository/
├── internal/transport/http/
└── migrations/
```
