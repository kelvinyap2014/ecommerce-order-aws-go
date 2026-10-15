package main

import (
	"context"
	"os"

	"ecommerce-order-aws-go/internal/logger"
	"ecommerce-order-aws-go/internal/models"
	"ecommerce-order-aws-go/internal/storage"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var orderRepo storage.OrderRepository

type UpdateStatusRequest struct {
	models.Order
	TargetStatus models.OrderStatus `json:"target_status"`
	Reason       string             `json:"reason,omitempty"`
}

func init() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load SDK config: %v", err)
		panic(err)
	}

	tableName := os.Getenv("ORDERS_TABLE_NAME")
	if tableName == "" {
		tableName = "ecommerce-order-orders-dev"
	}

	dbClient := dynamodb.NewFromConfig(cfg)
	orderRepo = storage.NewDynamoOrderRepo(dbClient, tableName)
}

func Handler(ctx context.Context, req UpdateStatusRequest) (models.Order, error) {
	logger.Info("Updating status for OrderID: %s to %s (Reason: %s)", req.OrderID, req.TargetStatus, req.Reason)

	err := orderRepo.UpdateOrderStatus(ctx, req.OrderID, req.TargetStatus)
	if err != nil {
		logger.Error("Failed to update status in DynamoDB for OrderID %s: %v", req.OrderID, err)
		return req.Order, err
	}

	req.Order.Status = req.TargetStatus
	logger.Info("Successfully updated order %s status to %s", req.OrderID, req.TargetStatus)
	return req.Order, nil
}

func main() {
	lambda.Start(Handler)
}
