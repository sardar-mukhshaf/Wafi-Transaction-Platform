package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/saudi-fabric/kafka-clients"
	"github.com/saudi-fabric/order-service/internal/handlers"
	"github.com/saudi-fabric/order-service/internal/outbox"
	"github.com/saudi-fabric/order-service/internal/saga"
)

func main() {
	cfg := LoadConfig()
	ctx := context.Background()

	// Initialize tracing
	if err := initTracer(ctx, cfg.OTelEndpoint); err != nil {
		log.Printf("failed to init tracer: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	// Run migrations
	if err := runMigrations(ctx, db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Ensure outbox table
	if err := kafka.EnsureOutboxTable(ctx, db); err != nil {
		log.Fatalf("failed to ensure outbox table: %v", err)
	}

	// Initialize saga orchestrator
	orchestrator := saga.NewOrchestrator(db)
	_ = orchestrator // used in event handlers

	// Initialize outbox publisher
	outboxPublisher := outbox.NewPublisher(db)

	// Start outbox relay in background
	go startOutboxRelay(ctx, cfg.KafkaBrokers, db)

	// Start Kafka consumer for payment events
	go startEventConsumer(ctx, cfg.KafkaBrokers, db, orchestrator)

	// Setup HTTP router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(otelhttp.NewMiddleware("order-service"))

	httpHandler := handlers.NewHTTPHandler(db, outboxPublisher)
	httpHandler.RegisterRoutes(r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("Order Service starting on port %d", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func initTracer(ctx context.Context, endpoint string) error {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY,
			customer_id VARCHAR(255) NOT NULL,
			correlation_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			total_amount NUMERIC(15,2) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);

		CREATE TABLE IF NOT EXISTS order_items (
			id UUID PRIMARY KEY,
			order_id UUID NOT NULL REFERENCES orders(id),
			product_id VARCHAR(255) NOT NULL,
			sku VARCHAR(255) NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price NUMERIC(15,2) NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
		CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
		CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
	`)
	return err
}

func startOutboxRelay(ctx context.Context, brokers string, db *sql.DB) {
	producer, err := kafka.NewProducer(brokers, "order-events")
	if err != nil {
		log.Printf("failed to create outbox producer: %v", err)
		return
	}
	defer producer.Close()

	store := kafka.NewOutboxStore(db)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, err := store.Poll(ctx, 100)
			if err != nil {
				log.Printf("outbox poll error: %v", err)
				continue
			}

			for _, ev := range events {
				if err := producer.Produce(ctx, []byte(ev.Key), ev.Payload); err != nil {
					log.Printf("outbox produce error: %v", err)
					continue
				}
				if err := store.MarkProcessed(ctx, ev.ID); err != nil {
					log.Printf("outbox mark processed error: %v", err)
				}
			}
		}
	}
}

func startEventConsumer(ctx context.Context, brokers string, db *sql.DB, orchestrator *saga.Orchestrator) {
	handler := func(msgCtx context.Context, key, value []byte) error {
		// Parse event type from headers or payload
		// For demo, inspect payload structure
		log.Printf("received event: key=%s", string(key))
		return nil
	}

	consumer, err := kafka.NewConsumer(brokers, "order-service-saga", []string{"payment-events"}, handler)
	if err != nil {
		log.Printf("failed to create saga consumer: %v", err)
		return
	}
	defer consumer.Close()

	if err := consumer.Run(ctx); err != nil {
		log.Printf("saga consumer error: %v", err)
	}
}
