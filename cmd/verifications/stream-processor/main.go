package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ecommerce-order-aws-go/internal/logger"
	"ecommerce-order-aws-go/internal/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

var (
	sfnClient      *sfn.Client
	sfnWorkflowArn string
)

func init() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Error("Failed to load SDK config: %v", err)
		panic(err)
	}

	sfnWorkflowArn = os.Getenv("STATE_MACHINE_ARN")
	sfnClient = sfn.NewFromConfig(cfg)
	logger.Info("Initialized Stream Processor with Workflow ARN")
}

func Handler(ctx context.Context, event events.DynamoDBEvent) error {
	for _, record := range event.Records {
		logger.Info("Processing DynamoDB Stream Event: Action=%s", record.EventName)

		// 1. Extract NewImage from the stream record
		newImage := record.Change.NewImage
		if len(newImage) == 0 {
			logger.Warn("Skipping record: NewImage is empty")
			continue
		}

		// 2. Convert DynamoDB stream NewImage to Order model, mapping ALL attributes
		var order models.Order

		// Convert Lambda event's DynamoDB AttributeValue format to JSON, then unmarshal to Order
		// This handles all Order fields: OrderID, CustomerID, Items, ShippingAddress, TotalAmount, Status, ShipmentStatus, CreatedAt
		jsonData, err := json.Marshal(newImage)
		if err != nil {
			logger.Error("Failed to marshal DynamoDB stream NewImage to JSON: %v", err)
			continue
		}

		err = json.Unmarshal(jsonData, &order)
		if err != nil {
			logger.Error("Failed to unmarshal DynamoDB stream record to Order struct: %v", err)
			continue
		}

		// Validate critical fields are present
		if order.OrderID == "" || order.Status == "" {
			logger.Warn("Skipping record: OrderID or Status is missing from stream event")
			continue
		}

		logger.Info("Triggering Step Functions workflow for Order ID: %s", order.OrderID)

		// 3. Start Step Functions Execution
		orderJSON, _ := json.Marshal(order)
		_, err = sfnClient.StartExecution(ctx, &sfn.StartExecutionInput{
			StateMachineArn: aws.String(sfnWorkflowArn),
			Name:            aws.String(fmt.Sprintf("stream-order-%s-%d", order.OrderID, time.Now().Unix())),
			Input:           aws.String(string(orderJSON)),
		})
		if err != nil {
			logger.Error("Failed to start workflow for order %s: %v", order.OrderID, err)
			return err // Returning error causes Lambda to retry this batch safely
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
