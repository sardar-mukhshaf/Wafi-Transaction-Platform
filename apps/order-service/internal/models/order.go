package models

import (
	"time"
)

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	StatusPending            OrderStatus = "PENDING"
	StatusPaymentRequested   OrderStatus = "PAYMENT_REQUESTED"
	StatusPaymentConfirmed   OrderStatus = "PAYMENT_CONFIRMED"
	StatusPaymentFailed      OrderStatus = "PAYMENT_FAILED"
	StatusInventoryReserved  OrderStatus = "INVENTORY_RESERVED"
	StatusCompleted          OrderStatus = "COMPLETED"
	StatusCancelled          OrderStatus = "CANCELLED"
)

// Order represents a commerce order.
type Order struct {
	ID            string      `json:"id" db:"id"`
	CustomerID    string      `json:"customer_id" db:"customer_id"`
	CorrelationID string      `json:"correlation_id" db:"correlation_id"`
	Status        OrderStatus `json:"status" db:"status"`
	Currency      string      `json:"currency" db:"currency"`
	TotalAmount   float64     `json:"total_amount" db:"total_amount"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
}

// OrderItem represents a line item in an order.
type OrderItem struct {
	ID        string  `json:"id" db:"id"`
	OrderID   string  `json:"order_id" db:"order_id"`
	ProductID string  `json:"product_id" db:"product_id"`
	SKU       string  `json:"sku" db:"sku"`
	Quantity  int32   `json:"quantity" db:"quantity"`
	UnitPrice float64 `json:"unit_price" db:"unit_price"`
}

// CreateOrderRequest is the incoming HTTP payload.
type CreateOrderRequest struct {
	CustomerID string           `json:"customer_id" validate:"required"`
	Currency   string           `json:"currency" validate:"required"`
	Items      []OrderItemInput `json:"items" validate:"required,min=1"`
}

// OrderItemInput is a line item in the create request.
type OrderItemInput struct {
	ProductID string  `json:"product_id" validate:"required"`
	SKU       string  `json:"sku" validate:"required"`
	Quantity  int32   `json:"quantity" validate:"required,min=1"`
	UnitPrice float64 `json:"unit_price" validate:"required,gt=0"`
}

// OrderResponse is the outgoing HTTP payload.
type OrderResponse struct {
	ID            string      `json:"id"`
	CustomerID    string      `json:"customer_id"`
	Status        OrderStatus `json:"status"`
	Currency      string      `json:"currency"`
	TotalAmount   float64     `json:"total_amount"`
	CorrelationID string      `json:"correlation_id"`
	CreatedAt     time.Time   `json:"created_at"`
}
