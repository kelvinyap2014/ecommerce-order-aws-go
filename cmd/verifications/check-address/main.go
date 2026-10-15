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

// CheckAddressResponse wraps the order with the check results for Step Functions.
type CheckAddressResponse struct {
	models.Order
	AddressCheckPassed bool   `json:"address_check_passed"`
	Reason             string `json:"reason,omitempty"`
}

func Handler(ctx context.Context, input OrderIDInput) (CheckAddressResponse, error) {
	logger.Info("Starting address and blacklist check for OrderID: %s", input.OrderID)

	// 1. Fetch full order from DynamoDB
	order, err := orderRepo.GetOrder(ctx, input.OrderID)
	if err != nil {
		logger.Error("Failed to fetch order %s: %v", input.OrderID, err)
		return CheckAddressResponse{}, err
	}

	addr := order.ShippingAddr

	// 2. Check for required address fields
	if strings.TrimSpace(addr.Street) == "" || strings.TrimSpace(addr.City) == "" || strings.TrimSpace(addr.Country) == "" {
		logger.Warn("Address check failed for OrderID %s: missing street, city, or country", order.OrderID)
		return CheckAddressResponse{
			Order:              *order,
			AddressCheckPassed: false,
			Reason:             "Missing required address fields (street, city, or country)",
		}, nil
	}

	// 3. Dummy blacklist check simulation (e.g. if street contains "BLACKLIST")
	if strings.Contains(strings.ToUpper(addr.Street), "BLACKLIST") {
		logger.Warn("Address check failed for OrderID %s: address matched blacklist", order.OrderID)
		return CheckAddressResponse{
			Order:              *order,
			AddressCheckPassed: false,
			Reason:             "Shipping address is blacklisted",
		}, nil
	}

	logger.Info("Address check passed successfully for OrderID: %s", order.OrderID)
	return CheckAddressResponse{
		Order:              *order,
		AddressCheckPassed: true,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
