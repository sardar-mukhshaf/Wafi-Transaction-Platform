package kafka

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OutboxEvent represents an event stored in the outbox table.
type OutboxEvent struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Key       string    `json:"key"`
	Payload   []byte    `json:"payload"`
	Headers   []byte    `json:"headers"`
	CreatedAt time.Time `json:"created_at"`
}

// OutboxStore handles persisting events to the outbox table.
type OutboxStore struct {
	db *sql.DB
}

// NewOutboxStore creates a new OutboxStore.
func NewOutboxStore(db *sql.DB) *OutboxStore {
	return &OutboxStore{db: db}
}

// Save stores an event in the outbox table within the given transaction.
func (o *OutboxStore) Save(ctx context.Context, tx *sql.Tx, topic, key string, payload, headers []byte) error {
	event := OutboxEvent{
		ID:        uuid.New().String(),
		Topic:     topic,
		Key:       key,
		Payload:   payload,
		Headers:   headers,
		CreatedAt: time.Now().UTC(),
	}

	headersJSON, err := json.Marshal(event.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO outbox_events (id, topic, key, payload, headers, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.ID, event.Topic, event.Key, event.Payload, headersJSON, event.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return nil
}

// Poll retrieves unprocessed outbox events.
func (o *OutboxStore) Poll(ctx context.Context, limit int) ([]OutboxEvent, error) {
	rows, err := o.db.QueryContext(ctx, `
		SELECT id, topic, key, payload, headers, created_at
		FROM outbox_events
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var ev OutboxEvent
		if err := rows.Scan(&ev.ID, &ev.Topic, &ev.Key, &ev.Payload, &ev.Headers, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// MarkProcessed marks an outbox event as processed.
func (o *OutboxStore) MarkProcessed(ctx context.Context, id string) error {
	_, err := o.db.ExecContext(ctx, `
		UPDATE outbox_events SET processed_at = $1 WHERE id = $2
	`, time.Now().UTC(), id)
	return err
}

// EnsureOutboxTable creates the outbox table if it does not exist.
func EnsureOutboxTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS outbox_events (
			id UUID PRIMARY KEY,
			topic VARCHAR(255) NOT NULL,
			key VARCHAR(255) NOT NULL,
			payload BYTEA NOT NULL,
			headers JSONB,
			created_at TIMESTAMP NOT NULL,
			processed_at TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed ON outbox_events(processed_at, created_at);
	`)
	return err
}
