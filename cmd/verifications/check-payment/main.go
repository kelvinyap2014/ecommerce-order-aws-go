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

// CheckPaymentResponse wraps the order with the payment check results for Step Functions.
type CheckPaymentResponse struct {
	models.Order
	PaymentCheckPassed bool   `json:"payment_check_passed"`
	Reason             string `json:"reason,omitempty"`
}

func Handler(ctx context.Context, input OrderIDInput) (CheckPaymentResponse, error) {
	logger.Info("Starting payment authorization check for OrderID: %s", input.OrderID)

	// 1. Fetch full order from DynamoDB
	order, err := orderRepo.GetOrder(ctx, input.OrderID)
	if err != nil {
		logger.Error("Failed to fetch order %s: %v", input.OrderID, err)
		return CheckPaymentResponse{}, err
	}

	logger.Info("Payment check for Amount: %.2f", order.TotalAmount)

	// 2. Dummy declined card simulation based on CustomerID
	if strings.Contains(strings.ToLower(order.CustomerID), "declined") {
		logger.Warn("Payment check failed for OrderID %s: card was declined", order.OrderID)
		return CheckPaymentResponse{
			Order:              *order,
			PaymentCheckPassed: false,
			Reason:             "Payment gateway declined transaction",
		}, nil
	}

	// 3. Dummy insufficient funds / high-value limit simulation (e.g., max $5,000 per single charge)
	if order.TotalAmount > 5000.0 {
		logger.Warn("Payment check failed for OrderID %s: amount exceeds single charge limit", order.OrderID)
		return CheckPaymentResponse{
			Order:              *order,
			PaymentCheckPassed: false,
			Reason:             "Insufficient funds or transaction limit exceeded",
		}, nil
	}

	logger.Info("Payment check passed successfully for OrderID: %s", order.OrderID)
	return CheckPaymentResponse{
		Order:              *order,
		PaymentCheckPassed: true,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
