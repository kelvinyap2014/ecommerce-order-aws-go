package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sfn"

	"ecommerce-order-aws-go/internal/logger"
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

		// 2. Extract OrderID from the stream record (minimal parsing)
		var orderData map[string]interface{}
		jsonData, err := json.Marshal(newImage)
		if err != nil {
			logger.Error("Failed to marshal DynamoDB stream NewImage to JSON: %v", err)
			continue
		}

		err = json.Unmarshal(jsonData, &orderData)
		if err != nil {
			logger.Error("Failed to unmarshal to map: %v", err)
			continue
		}

		orderID := ""
		if id, ok := orderData["OrderId"]; ok {
			// DynamoDB format: {"S": "uuid-value"}
			if idMap, ok := id.(map[string]interface{}); ok {
				if val, ok := idMap["S"]; ok {
					orderID = val.(string)
				}
			}
		}

		if orderID == "" {
			logger.Warn("Skipping record: OrderId not found or in unexpected format")
			continue
		}

		logger.Info("Triggering Step Functions workflow for Order ID: %s", orderID)

		// 3. Start Step Functions Execution with just the order_id
		// Downstream Lambdas will fetch the full order from DynamoDB
		input := map[string]interface{}{
			"order_id": orderID,
		}
		inputJSON, _ := json.Marshal(input)

		_, err = sfnClient.StartExecution(ctx, &sfn.StartExecutionInput{
			StateMachineArn: aws.String(sfnWorkflowArn),
			Name:            aws.String(fmt.Sprintf("stream-order-%s-%d", orderID, time.Now().Unix())),
			Input:           aws.String(string(inputJSON)),
		})
		if err != nil {
			logger.Error("Failed to start workflow for order %s: %v", orderID, err)
			return err // Returning error causes Lambda to retry this batch safely
		}
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
