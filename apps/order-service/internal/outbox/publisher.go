package outbox

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/saudi-fabric/kafka-clients"
	"github.com/saudi-fabric/order-service/internal/models"
)

// Publisher publishes domain events via the outbox pattern.
type Publisher struct {
	store *kafka.OutboxStore
}

// NewPublisher creates a new Publisher.
func NewPublisher(db *sql.DB) *Publisher {
	return &Publisher{store: kafka.NewOutboxStore(db)}
}

// PublishOrderCreated stores an OrderCreated event in the outbox.
func (p *Publisher) PublishOrderCreated(ctx context.Context, tx *sql.Tx, order models.Order, items []models.OrderItemInput) error {
	type orderItemPayload struct {
		ProductID string  `json:"product_id"`
		SKU       string  `json:"sku"`
		Quantity  int32   `json:"quantity"`
		UnitPrice float64 `json:"unit_price"`
	}

	var payloadItems []orderItemPayload
	for _, item := range items {
		payloadItems = append(payloadItems, orderItemPayload{
			ProductID: item.ProductID,
			SKU:       item.SKU,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	payload := map[string]interface{}{
		"order_id":       order.ID,
		"customer_id":    order.CustomerID,
		"correlation_id": order.CorrelationID,
		"created_at":     order.CreatedAt,
		"items":          payloadItems,
		"currency":       order.Currency,
		"total_amount":   order.TotalAmount,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"event-type":      "OrderCreated",
		"correlation-id":  order.CorrelationID,
		"service":         "order-service",
	}
	headersBytes, _ := json.Marshal(headers)

	return p.store.Save(ctx, tx, "order-events", order.ID, payloadBytes, headersBytes)
}

// PublishPaymentRequested stores a PaymentRequested event in the outbox.
func (p *Publisher) PublishPaymentRequested(ctx context.Context, tx *sql.Tx, orderID, paymentID, correlationID string, amount float64, currency, method string) error {
	payload := map[string]interface{}{
		"order_id":        orderID,
		"payment_id":      paymentID,
		"correlation_id":  correlationID,
		"amount":          amount,
		"currency":        currency,
		"payment_method":  method,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"event-type":     "OrderPaymentRequested",
		"correlation-id": correlationID,
		"service":        "order-service",
	}
	headersBytes, _ := json.Marshal(headers)

	return p.store.Save(ctx, tx, "order-events", orderID, payloadBytes, headersBytes)
}
