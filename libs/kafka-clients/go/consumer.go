package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Handler is a function that processes a Kafka message.
type Handler func(ctx context.Context, key, value []byte) error

// Consumer wraps a confluent Kafka consumer with OTel tracing.
type Consumer struct {
	consumer *kafka.Consumer
	handler  Handler
	topics   []string
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(brokers, groupID string, topics []string, handler Handler) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}
	return &Consumer{consumer: c, handler: handler, topics: topics}, nil
}

// Run starts consuming messages in a blocking loop.
func (c *Consumer) Run(ctx context.Context) error {
	if err := c.consumer.SubscribeTopics(c.topics, nil); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			continue
		}

		// Extract trace context from headers
		carrier := propagation.MapCarrier{}
		for _, h := range msg.Headers {
			carrier[h.Key] = string(h.Value)
		}
		msgCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
		msgCtx, span := tracer.Start(msgCtx, "kafka.Consume", trace.WithSpanKind(trace.SpanKindConsumer))

		if err := c.handler(msgCtx, msg.Key, msg.Value); err != nil {
			span.RecordError(err)
			span.End()
			continue
		}

		span.End()
		c.consumer.CommitMessage(msg)
	}
}

// Close shuts down the consumer.
func (c *Consumer) Close() {
	c.consumer.Close()
}
