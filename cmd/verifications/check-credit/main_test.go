package main

import (
	"context"
	"testing"

	"ecommerce-order-aws-go/internal/models"
)

func TestHandler_CreditCheck(t *testing.T) {
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
			name:       "Normal Order Passes Credit Check",
			order:      validOrder,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Handler(ctx, tt.order)
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
