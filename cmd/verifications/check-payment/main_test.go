package main

import (
	"context"
	"testing"

	"ecommerce-order-aws-go/internal/models"
)

func TestHandler_PaymentCheck(t *testing.T) {
	ctx := context.TODO()

	validOrder := models.Order{
		OrderID:     "ord-123",
		CustomerID:  "cust-456",
		TotalAmount: 150.00,
	}

	tests := []struct {
		name                    string
		order                   models.Order
		wantPassed              bool
		expectedReasonSubstring string
	}{
		{
			name:       "Normal Order Passes Payment Check",
			order:      validOrder,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Handler(ctx, tt.order)
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
