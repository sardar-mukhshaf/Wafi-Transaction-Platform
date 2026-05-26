package saga

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/saudi-fabric/kafka-clients"
	"github.com/saudi-fabric/order-service/internal/models"
)

// Orchestrator handles the saga lifecycle for order processing.
type Orchestrator struct {
	db        *sql.DB
	outbox    *kafka.OutboxStore
}

// NewOrchestrator creates a new saga orchestrator.
func NewOrchestrator(db *sql.DB) *Orchestrator {
	return &Orchestrator{
		db:     db,
		outbox: kafka.NewOutboxStore(db),
	}
}

// StartPaymentSaga initiates the payment phase of the order saga.
func (o *Orchestrator) StartPaymentSaga(ctx context.Context, orderID, correlationID string, amount float64, currency string) error {
	paymentID := uuid.New().String()

	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	// Update order status
	_, err = tx.ExecContext(ctx, `
		UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3
	`, models.StatusPaymentRequested, time.Now().UTC(), orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Publish payment requested event via outbox
	payload := map[string]interface{}{
		"order_id":       orderID,
		"payment_id":     paymentID,
		"correlation_id": correlationID,
		"amount":         amount,
		"currency":       currency,
		"payment_method": "MADA", // default for demo
	}
	payloadBytes, _ := json.Marshal(payload)
	headers := map[string]string{"event-type": "OrderPaymentRequested", "correlation-id": correlationID}
	headersBytes, _ := json.Marshal(headers)

	if err := o.outbox.Save(ctx, tx, "order-events", orderID, payloadBytes, headersBytes); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return tx.Commit()
}

// HandlePaymentConfirmed processes the payment confirmation event.
func (o *Orchestrator) HandlePaymentConfirmed(ctx context.Context, orderID, correlationID string) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3
	`, models.StatusPaymentConfirmed, time.Now().UTC(), orderID)
	if err != nil {
		return err
	}

	// Next saga step: request inventory reservation
	payload := map[string]interface{}{
		"order_id":       orderID,
		"correlation_id": correlationID,
		"reserved_at":    time.Now().UTC(),
	}
	payloadBytes, _ := json.Marshal(payload)
	headers := map[string]string{"event-type": "OrderInventoryReserved", "correlation-id": correlationID}
	headersBytes, _ := json.Marshal(headers)

	if err := o.outbox.Save(ctx, tx, "order-events", orderID, payloadBytes, headersBytes); err != nil {
		return err
	}

	return tx.Commit()
}

// HandlePaymentFailed triggers compensating transaction.
func (o *Orchestrator) HandlePaymentFailed(ctx context.Context, orderID, correlationID, reason string) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3
	`, models.StatusCancelled, time.Now().UTC(), orderID)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"order_id":       orderID,
		"correlation_id": correlationID,
		"reason":         reason,
		"cancelled_at":   time.Now().UTC(),
	}
	payloadBytes, _ := json.Marshal(payload)
	headers := map[string]string{"event-type": "OrderCancelled", "correlation-id": correlationID}
	headersBytes, _ := json.Marshal(headers)

	if err := o.outbox.Save(ctx, tx, "order-events", orderID, payloadBytes, headersBytes); err != nil {
		return err
	}

	log.Printf("[SAGA] Order %s cancelled due to payment failure: %s", orderID, reason)
	return tx.Commit()
}

// HandleInventoryReserved completes the saga.
func (o *Orchestrator) HandleInventoryReserved(ctx context.Context, orderID, correlationID string) error {
	_, err := o.db.ExecContext(ctx, `
		UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3
	`, models.StatusCompleted, time.Now().UTC(), orderID)
	if err != nil {
		return err
	}

	log.Printf("[SAGA] Order %s completed successfully", orderID)
	return nil
}
