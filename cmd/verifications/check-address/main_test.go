package main

import (
	"context"
	"testing"

	"ecommerce-order-aws-go/internal/models"
)

func TestHandler_AddressCheck(t *testing.T) {
	ctx := context.TODO()

	validOrder := models.Order{
		OrderID:    "ord-123",
		CustomerID: "cust-456",
		ShippingAddr: models.Address{
			Street:  "123 Queen Street",
			City:    "Auckland",
			Country: "New Zealand",
		},
	}

	tests := []struct {
		name                    string
		order                   models.Order
		wantPassed              bool
		expectedReasonSubstring string
	}{
		{
			name:       "Valid Address Passes",
			order:      validOrder,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := Handler(ctx, tt.order)
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
