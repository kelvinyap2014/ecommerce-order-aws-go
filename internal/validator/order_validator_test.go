package validator

import (
	"ecommerce-order-aws-go/internal/models"
	"testing"
)

func TestValidateOrder(t *testing.T) {
	validOrder := models.Order{
		CustomerID: "cust_123",
		Items: []models.OrderItem{
			{SKU: "SKU-PROD-1", Quantity: 2, Price: 49.99},
		},
		TotalAmount: 99.98,
		ShippingAddr: models.Address{
			City:    "Auckland",
			Country: "New Zealand",
		},
	}

	tests := []struct {
		name    string
		order   models.Order
		wantErr error
	}{
		{
			name:    "Valid Order Passes",
			order:   validOrder,
			wantErr: nil,
		},
		{
			name: "Missing Customer ID Fails",
			order: models.Order{
				CustomerID:   "",
				Items:        validOrder.Items,
				TotalAmount:  validOrder.TotalAmount,
				ShippingAddr: validOrder.ShippingAddr,
			},
			wantErr: ErrEmptyCustomerID,
		},
		{
			name: "Empty Items Fails",
			order: models.Order{
				CustomerID:   "cust_123",
				Items:        []models.OrderItem{},
				TotalAmount:  99.98,
				ShippingAddr: validOrder.ShippingAddr,
			},
			wantErr: ErrNoItems,
		},
		{
			name: "Zero Total Amount Fails",
			order: models.Order{
				CustomerID:   "cust_123",
				Items:        validOrder.Items,
				TotalAmount:  0,
				ShippingAddr: validOrder.ShippingAddr,
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "Missing Country in Address Fails",
			order: models.Order{
				CustomerID:   "cust_123",
				Items:        validOrder.Items,
				TotalAmount:  99.98,
				ShippingAddr: models.Address{City: "Auckland", Country: ""},
			},
			wantErr: ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrder(tt.order)
			if err != tt.wantErr {
				t.Errorf("ValidateOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
