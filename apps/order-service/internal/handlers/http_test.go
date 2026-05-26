package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/saudi-fabric/order-service/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE orders (
			id TEXT PRIMARY KEY,
			customer_id TEXT NOT NULL,
			correlation_id TEXT NOT NULL,
			status TEXT NOT NULL,
			currency TEXT NOT NULL,
			total_amount REAL NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
		CREATE TABLE order_items (
			id TEXT PRIMARY KEY,
			order_id TEXT NOT NULL,
			product_id TEXT NOT NULL,
			sku TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
	return db
}

func TestCreateOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := NewHTTPHandler(db, nil)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	reqBody := models.CreateOrderRequest{
		CustomerID: "cust-123",
		Currency:   "SAR",
		Items: []models.OrderItemInput{
			{ProductID: "prod-1", SKU: "SKU001", Quantity: 2, UnitPrice: 150.00},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", "test-correlation-123")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var resp models.OrderResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.CustomerID != "cust-123" {
		t.Errorf("expected customer_id cust-123, got %s", resp.CustomerID)
	}

	if resp.Status != models.StatusPending {
		t.Errorf("expected status %s, got %s", models.StatusPending, resp.Status)
	}

	if resp.TotalAmount != 300.00 {
		t.Errorf("expected total 300.00, got %.2f", resp.TotalAmount)
	}

	if resp.CorrelationID != "test-correlation-123" {
		t.Errorf("expected correlation_id test-correlation-123, got %s", resp.CorrelationID)
	}
}

func TestGetOrder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed data
	_, err := db.Exec(`
		INSERT INTO orders (id, customer_id, correlation_id, status, currency, total_amount, created_at, updated_at)
		VALUES ('order-123', 'cust-123', 'corr-123', 'PENDING', 'SAR', 150.00, datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}

	handler := NewHTTPHandler(db, nil)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/orders/order-123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp models.OrderResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "order-123" {
		t.Errorf("expected id order-123, got %s", resp.ID)
	}
}

func TestHealthCheck(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	handler := NewHTTPHandler(db, nil)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}
