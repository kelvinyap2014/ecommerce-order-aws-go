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

func TestHandler_AddressCheck(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name                    string
		order                   models.Order
		wantPassed              bool
		expectedReasonSubstring string
	}{
		{
			name: "Valid Address Passes",
			order: models.Order{
				OrderID:    "ord-999-valid",
				CustomerID: "cust-456",
				ShippingAddr: models.Address{
					Street:  "123 Queen Street",
					City:    "Auckland",
					Country: "New Zealand",
				},
			},
			wantPassed: true,
		},
		{
			name: "Missing Street Fails",
			order: models.Order{
				OrderID:    "ord-124",
				CustomerID: "cust-456",
				ShippingAddr: models.Address{
					Street:  "",
					City:    "Auckland",
					Country: "New Zealand",
				},
			},
			wantPassed:              false,
			expectedReasonSubstring: "Missing required address fields",
		},
		{
			name: "Missing City Fails",
			order: models.Order{
				OrderID:    "ord-125",
				CustomerID: "cust-456",
				ShippingAddr: models.Address{
					Street:  "123 Queen Street",
					City:    "",
					Country: "New Zealand",
				},
			},
			wantPassed:              false,
			expectedReasonSubstring: "Missing required address fields",
		},
		{
			name: "Missing Country Fails",
			order: models.Order{
				OrderID:    "ord-126",
				CustomerID: "cust-456",
				ShippingAddr: models.Address{
					Street:  "123 Queen Street",
					City:    "Auckland",
					Country: "",
				},
			},
			wantPassed:              false,
			expectedReasonSubstring: "Missing required address fields",
		},
		{
			name: "Blacklisted Street Fails",
			order: models.Order{
				OrderID:    "ord-127",
				CustomerID: "cust-456",
				ShippingAddr: models.Address{
					Street:  "999 Blacklist Ave",
					City:    "Wellington",
					Country: "New Zealand",
				},
			},
			wantPassed:              false,
			expectedReasonSubstring: "blacklisted",
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

			if resp.AddressCheckPassed != tt.wantPassed {
				t.Errorf("Handler() AddressCheckPassed = %v, want %v (Reason: %s)", resp.AddressCheckPassed, tt.wantPassed, resp.Reason)
			}

			if !tt.wantPassed && tt.expectedReasonSubstring != "" {
				if resp.Reason == "" {
					t.Errorf("Handler() expected a reason string, got empty")
				}
			}
		})
	}
}
