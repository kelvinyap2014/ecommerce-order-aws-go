package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ecommerce-order-aws-go/internal/logger"
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
	logger.Info("Initializing Lambda handler with DynamoDB table: %s", tableName)

	dbClient := dynamodb.NewFromConfig(cfg)
	orderRepo = storage.NewDynamoOrderRepo(dbClient, tableName)
	logger.Info("Lambda initialization complete")
}

func Handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := request.RequestContext.HTTP.Method
	routeKey := request.RouteKey
	logger.Info("Incoming request - method: %s, route: %s", method, routeKey)

	// Route: POST /orders (Create Order)
	if method == "POST" || routeKey == "POST /orders" {
		return handleCreateOrder(ctx, request)
	}

	// Route: GET /orders (List All Orders)
	if method == "GET" || routeKey == "GET /orders" {
		return handleListOrders(ctx)
	}

	logger.Warn("Route not found for method: %s, route: %s", method, routeKey)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 404,
		Body:       `{"error": "not found"}`,
	}, nil
}

func handleCreateOrder(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// 1. Parse incoming JSON request body
	var order models.Order
	if err := json.Unmarshal([]byte(request.Body), &order); err != nil {
		logger.Error("Failed to parse request body: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "invalid request body: %v"}`, err),
		}, nil
	}
	logger.Info("Request body parsed successfully")

	// 2. Enrich order with system metadata
	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}
	order.Status = models.StatusPending
	order.ShipmentStatus = models.ShipmentNotShipped
	order.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	logger.Info("Order enriched - OrderID: %s, Status: %s, ShipmentStatus: %s", order.OrderID, order.Status, order.ShipmentStatus)

	// 3. Run business validation logic
	if err := validator.ValidateOrder(order); err != nil {
		logger.Error("Order validation failed for OrderID %s: %v", order.OrderID, err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
		}, nil
	}
	logger.Info("Order validation passed for OrderID: %s", order.OrderID)

	// 4. Persist to DynamoDB repository
	if err := orderRepo.SaveOrder(ctx, order); err != nil {
		logger.Error("Failed to save order %s to DynamoDB: %v", order.OrderID, err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       `{"error": "internal server error"}`,
		}, nil
	}
	logger.Info("Order saved successfully - OrderID: %s", order.OrderID)

	// 5. Construct successful response payload
	responseBody, _ := json.Marshal(map[string]any{
		"message":  "Order created successfully",
		"order_id": order.OrderID,
		"status":   order.Status,
	})

	logger.Info("Returning 201 Created response for OrderID: %s", order.OrderID)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func handleListOrders(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	logger.Info("Listing all orders")
	orders, err := orderRepo.ListOrders(ctx)
	if err != nil {
		logger.Error("Failed to list orders: %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       `{"error": "internal server error"}`,
		}, nil
	}
	logger.Info("Retrieved %d orders from DynamoDB", len(orders))

	responseBody, _ := json.Marshal(orders)
	logger.Info("Returning 200 OK with orders list")
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(responseBody),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
