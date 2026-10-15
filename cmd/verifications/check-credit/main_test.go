package main

import (
	"context"
	"fmt"
	"testing"

	"ecommerce-order-aws-go/internal/models"
)

// mockOrderRepository implements storage.OrderRepository for testing
type mockOrderRepository struct {
	orders map[string]*models.Order
}

func (m *mockOrderRepository) SaveOrder(ctx context.Context, order models.Order) error {
	return nil
}

func (m *mockOrderRepository) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	if order, ok := m.orders[orderID]; ok {
		return order, nil
	}
	return nil, fmt.Errorf("order not found: %s", orderID)
}

func (m *mockOrderRepository) ListOrders(ctx context.Context) ([]models.Order, error) {
	return nil, nil
}

func (m *mockOrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	return nil
}

func (m *mockOrderRepository) UpdateShipmentStatus(ctx context.Context, orderID string, shipmentStatus models.ShipmentStatus, wmsShipmentID string, wmsAcceptedAt string) error {
	return nil
}

func TestHandler_CreditCheck(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name                    string
		order                   models.Order
		wantPassed              bool
		expectedReasonSubstring string
	}{
		{
			name: "Normal Order Passes Credit Check",
			order: models.Order{
				OrderID:     "ord-123-valid-credit",
				CustomerID:  "cust-456",
				TotalAmount: 150.00,
			},
			wantPassed: true,
		},
		{
			name: "Bad Credit Customer Fails",
			order: models.Order{
				OrderID:     "ord-124",
				CustomerID:  "cust-bad-credit",
				TotalAmount: 100.00,
			},
			wantPassed:              false,
			expectedReasonSubstring: "credit history check failed",
		},
		{
			name: "Excessive Amount Fails",
			order: models.Order{
				OrderID:     "ord-125",
				CustomerID:  "cust-456",
				TotalAmount: 15000.00,
			},
			wantPassed:              false,
			expectedReasonSubstring: "exceeds maximum credit limit",
		},
	}

	// Setup mock repository
	mockRepo := &mockOrderRepository{
		orders: make(map[string]*models.Order),
	}
	for _, tt := range tests {
		mockRepo.orders[tt.order.OrderID] = &tt.order
	}
	orderRepo = mockRepo

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Handler(ctx, OrderIDInput{OrderID: tt.order.OrderID})
			if err != nil {
				t.Fatalf("Handler() unexpected error = %v", err)
			}

			if resp.CreditCheckPassed != tt.wantPassed {
				t.Errorf("Handler() CreditCheckPassed = %v, want %v (Reason: %s)", resp.CreditCheckPassed, tt.wantPassed, resp.Reason)
			}

			if !tt.wantPassed && tt.expectedReasonSubstring != "" {
				if resp.Reason == "" {
					t.Errorf("Handler() expected a reason string, got empty")
				}
			}
		})
	}
}
