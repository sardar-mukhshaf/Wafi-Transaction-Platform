package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("kafka-client")

// Producer wraps a confluent Kafka producer with OTel tracing.
type Producer struct {
	producer *kafka.Producer
	topic    string
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"client.id":         "fabric-producer",
		"acks":              "all",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	return &Producer{producer: p, topic: topic}, nil
}

// Produce sends a message to Kafka with tracing context.
func (p *Producer) Produce(ctx context.Context, key []byte, value []byte) error {
	ctx, span := tracer.Start(ctx, "kafka.Produce", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()

	headers := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, headers)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
	}
	for k, v := range headers {
		msg.Headers = append(msg.Headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	deliveryChan := make(chan kafka.Event, 1)
	if err := p.producer.Produce(msg, deliveryChan); err != nil {
		span.RecordError(err)
		return err
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		span.RecordError(m.TopicPartition.Error)
		return m.TopicPartition.Error
	}
	return nil
}

// Close flushes and closes the producer.
func (p *Producer) Close() {
	p.producer.Flush(15000)
	p.producer.Close()
}
