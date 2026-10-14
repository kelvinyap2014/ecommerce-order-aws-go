package main

import (
	"context"
	"strings"

	"ecommerce-order-aws-go/internal/logger"
	"ecommerce-order-aws-go/internal/models"

	"github.com/aws/aws-lambda-go/lambda"
)

// CheckPaymentResponse wraps the order with the payment check results for Step Functions.
type CheckPaymentResponse struct {
	models.Order
	PaymentCheckPassed bool   `json:"payment_check_passed"`
	Reason             string `json:"reason,omitempty"`
}

func Handler(ctx context.Context, order models.Order) (CheckPaymentResponse, error) {
	logger.Info("Starting payment authorization check for OrderID: %s, Amount: %.2f", order.OrderID, order.TotalAmount)

	// 1. Dummy declined card simulation based on CustomerID
	if strings.Contains(strings.ToLower(order.CustomerID), "declined") {
		logger.Warn("Payment check failed for OrderID %s: card was declined", order.OrderID)
		return CheckPaymentResponse{
			Order:              order,
			PaymentCheckPassed: false,
			Reason:             "Payment gateway declined transaction",
		}, nil
	}

	// 2. Dummy insufficient funds / high-value limit simulation (e.g., max $5,000 per single charge)
	if order.TotalAmount > 5000.0 {
		logger.Warn("Payment check failed for OrderID %s: amount exceeds single charge limit", order.OrderID)
		return CheckPaymentResponse{
			Order:              order,
			PaymentCheckPassed: false,
			Reason:             "Insufficient funds or transaction limit exceeded",
		}, nil
	}

	logger.Info("Payment check passed successfully for OrderID: %s", order.OrderID)
	return CheckPaymentResponse{
		Order:              order,
		PaymentCheckPassed: true,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
