package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ecommerce-order-aws-go/internal/models"
	"ecommerce-order-aws-go/internal/storage"
	"ecommerce-order-aws-go/internal/validator"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

var orderRepo storage.OrderRepository

// init runs once when the Lambda container spins up, caching the DynamoDB client.
func init() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load SDK config: %v", err))
	}

	tableName := os.Getenv("ORDERS_TABLE_NAME")
	if tableName == "" {
		tableName = "ecommerce-order-orders-dev" // Fallback default
	}

	dbClient := dynamodb.NewFromConfig(cfg)
	orderRepo = storage.NewDynamoOrderRepo(dbClient, tableName)
}

func Handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// 1. Parse incoming JSON request body
	var order models.Order
	if err := json.Unmarshal([]byte(request.Body), &order); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "invalid request body: %v"}`, err),
		}, nil
	}

	// 2. Enrich order with system metadata
	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}
	order.Status = models.StatusPending
	order.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	// 3. Run business validation logic
	if err := validator.ValidateOrder(order); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
		}, nil
	}

	// 4. Persist to DynamoDB repository
	if err := orderRepo.SaveOrder(ctx, order); err != nil {
		fmt.Printf("Failed to save order %s to DynamoDB: %v\n", order.OrderID, err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       `{"error": "internal server error"}`,
		}, nil
	}

	// 5. Construct successful response payload
	responseBody, _ := json.Marshal(map[string]any{
		"message":  "Order created successfully",
		"order_id": order.OrderID,
		"status":   order.Status,
	})

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
