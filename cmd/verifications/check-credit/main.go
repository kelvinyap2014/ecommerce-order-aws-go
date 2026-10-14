package main

import (
	"context"
	"strings"

	"ecommerce-order-aws-go/internal/logger"
	"ecommerce-order-aws-go/internal/models"

	"github.com/aws/aws-lambda-go/lambda"
)

// CheckCreditResponse wraps the order with the credit check results for Step Functions.
type CheckCreditResponse struct {
	models.Order
	CreditCheckPassed bool   `json:"credit_check_passed"`
	Reason            string `json:"reason,omitempty"`
}

func Handler(ctx context.Context, order models.Order) (CheckCreditResponse, error) {
	logger.Info("Starting credit check for OrderID: %s, CustomerID: %s, Amount: %.2f", order.OrderID, order.CustomerID, order.TotalAmount)

	// 1. Dummy blacklist/bad-credit customer check
	if strings.Contains(strings.ToLower(order.CustomerID), "bad-credit") {
		logger.Warn("Credit check failed for OrderID %s: customer has bad credit record", order.OrderID)
		return CheckCreditResponse{
			Order:             order,
			CreditCheckPassed: false,
			Reason:            "Customer credit history check failed",
		}, nil
	}

	// 2. Dummy high-value credit limit check (e.g., max $10,000 without manual review)
	if order.TotalAmount > 10000.0 {
		logger.Warn("Credit check failed for OrderID %s: order amount exceeds credit limit", order.OrderID)
		return CheckCreditResponse{
			Order:             order,
			CreditCheckPassed: false,
			Reason:            "Order amount exceeds maximum credit limit",
		}, nil
	}

	logger.Info("Credit check passed successfully for OrderID: %s", order.OrderID)
	return CheckCreditResponse{
		Order:             order,
		CreditCheckPassed: true,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
