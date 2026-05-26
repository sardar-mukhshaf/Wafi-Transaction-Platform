# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy shared lib first for better caching
COPY libs/kafka-clients/go /libs/kafka-clients/go
COPY apps/order-service/go.mod apps/order-service/go.sum ./
RUN go mod download

COPY apps/order-service/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o order-service .

# Runtime stage - distroless for security
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/order-service /order-service

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/order-service"]
