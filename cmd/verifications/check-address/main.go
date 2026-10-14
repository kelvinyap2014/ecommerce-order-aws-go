package main

import (
	"context"
	"strings"

	"ecommerce-order-aws-go/internal/logger"
	"ecommerce-order-aws-go/internal/models"

	"github.com/aws/aws-lambda-go/lambda"
)

// CheckAddressResponse wraps the order with the check results for Step Functions.
type CheckAddressResponse struct {
	models.Order
	AddressCheckPassed bool   `json:"address_check_passed"`
	Reason             string `json:"reason,omitempty"`
}

func Handler(ctx context.Context, order models.Order) (CheckAddressResponse, error) {
	logger.Info("Starting address and blacklist check for OrderID: %s", order.OrderID)

	addr := order.ShippingAddr

	// 1. Check for required address fields
	if strings.TrimSpace(addr.Street) == "" || strings.TrimSpace(addr.City) == "" || strings.TrimSpace(addr.Country) == "" {
		logger.Warn("Address check failed for OrderID %s: missing street, city, or country", order.OrderID)
		return CheckAddressResponse{
			Order:              order,
			AddressCheckPassed: false,
			Reason:             "Missing required address fields (street, city, or country)",
		}, nil
	}

	// 2. Dummy blacklist check simulation (e.g. if street contains "BLACKLIST")
	if strings.Contains(strings.ToUpper(addr.Street), "BLACKLIST") {
		logger.Warn("Address check failed for OrderID %s: address matched blacklist", order.OrderID)
		return CheckAddressResponse{
			Order:              order,
			AddressCheckPassed: false,
			Reason:             "Shipping address is blacklisted",
		}, nil
	}

	logger.Info("Address check passed successfully for OrderID: %s", order.OrderID)
	return CheckAddressResponse{
		Order:              order,
		AddressCheckPassed: true,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
