module github.com/saudi-fabric/order-service

go 1.21

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.2.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/google/uuid v1.3.1
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/saudi-fabric/kafka-clients v0.0.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.45.0
	go.opentelemetry.io/otel v1.19.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0
	go.opentelemetry.io/otel/sdk v1.19.0
	go.opentelemetry.io/otel/trace v1.19.0
	google.golang.org/protobuf v1.31.0
)

replace github.com/saudi-fabric/kafka-clients => ../../libs/kafka-clients/go
