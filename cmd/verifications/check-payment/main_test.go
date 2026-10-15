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

func TestHandler_PaymentCheck(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name                    string
		order                   models.Order
		wantPassed              bool
		expectedReasonSubstring string
	}{
		{
			name: "Normal Order Passes Payment Check",
			order: models.Order{
				OrderID:     "ord-123-valid-payment",
				CustomerID:  "cust-456",
				TotalAmount: 150.00,
			},
			wantPassed: true,
		},
		{
			name: "Declined Card Customer Fails",
			order: models.Order{
				OrderID:     "ord-124",
				CustomerID:  "cust-declined",
				TotalAmount: 100.00,
			},
			wantPassed:              false,
			expectedReasonSubstring: "declined transaction",
		},
		{
			name: "Amount Exceeds Limit Fails",
			order: models.Order{
				OrderID:     "ord-125",
				CustomerID:  "cust-456",
				TotalAmount: 6000.00,
			},
			wantPassed:              false,
			expectedReasonSubstring: "Insufficient funds",
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

			if resp.PaymentCheckPassed != tt.wantPassed {
				t.Errorf("Handler() PaymentCheckPassed = %v, want %v (Reason: %s)", resp.PaymentCheckPassed, tt.wantPassed, resp.Reason)
			}

			if !tt.wantPassed && tt.expectedReasonSubstring != "" {
				if resp.Reason == "" {
					t.Errorf("Handler() expected a reason string, got empty")
				}
			}
		})
	}
}
