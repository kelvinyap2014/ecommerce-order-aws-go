package main

import (
	"context"
	"os"
	"strings"

	"ecommerce-order-aws-go/internal/logger"
	"ecommerce-order-aws-go/internal/models"
	"ecommerce-order-aws-go/internal/storage"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	ddbClient *dynamodb.Client
	orderRepo storage.OrderRepository
)

func init() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load SDK config: %v", err)
		panic(err)
	}

	ddbClient = dynamodb.NewFromConfig(cfg)
	tableName := os.Getenv("ORDERS_TABLE_NAME")
	orderRepo = storage.NewDynamoOrderRepo(ddbClient, tableName)
}

// OrderIDInput represents the input from Step Functions containing just the order_id.
type OrderIDInput struct {
	OrderID string `json:"order_id"`
}

// CheckCreditResponse wraps the order with the credit check results for Step Functions.
type CheckCreditResponse struct {
	models.Order
	CreditCheckPassed bool   `json:"credit_check_passed"`
	Reason            string `json:"reason,omitempty"`
}

func Handler(ctx context.Context, input OrderIDInput) (CheckCreditResponse, error) {
	logger.Info("Starting credit check for OrderID: %s", input.OrderID)

	// 1. Fetch full order from DynamoDB
	order, err := orderRepo.GetOrder(ctx, input.OrderID)
	if err != nil {
		logger.Error("Failed to fetch order %s: %v", input.OrderID, err)
		return CheckCreditResponse{}, err
	}

	logger.Info("Credit check for CustomerID: %s, Amount: %.2f", order.CustomerID, order.TotalAmount)

	// 2. Dummy blacklist/bad-credit customer check
	if strings.Contains(strings.ToLower(order.CustomerID), "bad-credit") {
		logger.Warn("Credit check failed for OrderID %s: customer has bad credit record", order.OrderID)
		return CheckCreditResponse{
			Order:             *order,
			CreditCheckPassed: false,
			Reason:            "Customer credit history check failed",
		}, nil
	}

	// 3. Dummy high-value credit limit check (e.g., max $10,000 without manual review)
	if order.TotalAmount > 10000.0 {
		logger.Warn("Credit check failed for OrderID %s: order amount exceeds credit limit", order.OrderID)
		return CheckCreditResponse{
			Order:             *order,
			CreditCheckPassed: false,
			Reason:            "Order amount exceeds maximum credit limit",
		}, nil
	}

	logger.Info("Credit check passed successfully for OrderID: %s", order.OrderID)
	return CheckCreditResponse{
		Order:             *order,
		CreditCheckPassed: true,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
