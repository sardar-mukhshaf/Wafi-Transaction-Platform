package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/saudi-fabric/order-service/internal/models"
	"github.com/saudi-fabric/order-service/internal/outbox"
)

// HTTPHandler handles HTTP requests for the order service.
type HTTPHandler struct {
	db        *sql.DB
	publisher *outbox.Publisher
}

// NewHTTPHandler creates a new HTTPHandler.
func NewHTTPHandler(db *sql.DB, publisher *outbox.Publisher) *HTTPHandler {
	return &HTTPHandler{db: db, publisher: publisher}
}

// RegisterRoutes registers HTTP routes.
func (h *HTTPHandler) RegisterRoutes(r chi.Router) {
	r.Post("/orders", h.CreateOrder)
	r.Get("/orders/{id}", h.GetOrder)
	r.Get("/health", h.HealthCheck)
	r.Get("/ready", h.ReadinessCheck)
}

// CreateOrder handles POST /orders.
func (h *HTTPHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculate total
	var total float64
	for _, item := range req.Items {
		total += item.UnitPrice * float64(item.Quantity)
	}

	orderID := uuid.New().String()
	correlationID := r.Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = uuid.New().String()
	}

	order := models.Order{
		ID:            orderID,
		CustomerID:    req.CustomerID,
		CorrelationID: correlationID,
		Status:        models.StatusPending,
		Currency:      req.Currency,
		TotalAmount:   total,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO orders (id, customer_id, correlation_id, status, currency, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, order.ID, order.CustomerID, order.CorrelationID, order.Status, order.Currency, order.TotalAmount, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert items
	for _, item := range req.Items {
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO order_items (id, order_id, product_id, sku, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), order.ID, item.ProductID, item.SKU, item.Quantity, item.UnitPrice)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Publish OrderCreated via outbox (atomic with DB transaction)
	if err := h.publisher.PublishOrderCreated(r.Context(), tx, order, req.Items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := models.OrderResponse{
		ID:            order.ID,
		CustomerID:    order.CustomerID,
		Status:        order.Status,
		Currency:      order.Currency,
		TotalAmount:   order.TotalAmount,
		CorrelationID: order.CorrelationID,
		CreatedAt:     order.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Correlation-ID", correlationID)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetOrder handles GET /orders/{id}.
func (h *HTTPHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var order models.Order
	err := h.db.QueryRowContext(r.Context(), `
		SELECT id, customer_id, correlation_id, status, currency, total_amount, created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(&order.ID, &order.CustomerID, &order.CorrelationID, &order.Status, &order.Currency, &order.TotalAmount, &order.CreatedAt, &order.UpdatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := models.OrderResponse{
		ID:            order.ID,
		CustomerID:    order.CustomerID,
		Status:        order.Status,
		Currency:      order.Currency,
		TotalAmount:   order.TotalAmount,
		CorrelationID: order.CorrelationID,
		CreatedAt:     order.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HealthCheck handles GET /health.
func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// ReadinessCheck handles GET /ready.
func (h *HTTPHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.db.PingContext(r.Context()); err != nil {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}
