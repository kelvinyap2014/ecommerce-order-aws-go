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

// WmsIntegrationResponse represents the WMS integration result for Step Functions.
type WmsIntegrationResponse struct {
	models.Order
	WmsShipmentID          string `json:"wms_shipment_id"`
	WmsSubmissionTimestamp string `json:"wms_submission_timestamp"`
	WmsAcceptedAt          string `json:"wms_accepted_at"`
}

// MockWmsPayload represents the data structure sent to the WMS system (mock).
type MockWmsPayload struct {
	ShipmentID      string             `json:"shipment_id"`
	OrderID         string             `json:"order_id"`
	CustomerID      string             `json:"customer_id"`
	Items           []models.OrderItem `json:"items"`
	ShippingAddress models.Address     `json:"shipping_address"`
	TotalAmount     float64            `json:"total_amount"`
	SubmittedAt     string             `json:"submitted_at"`
}

func Handler(ctx context.Context, order models.Order) (WmsIntegrationResponse, error) {
	logger.Info("Starting WMS integration dispatch for OrderID: %s, CustomerID: %s", order.OrderID, order.CustomerID)

	// 1. Generate unique WMS shipment ID based on order ID and timestamp
	wmsShipmentID := fmt.Sprintf("WMS-%s-%d", order.OrderID[:8], time.Now().Unix())
	submittedAt := time.Now().UTC().Format(time.RFC3339)

	// 2. Log order details for WMS integration
	logger.Info("Order Details for WMS:")
	logger.Info("  OrderID: %s", order.OrderID)
	logger.Info("  CustomerID: %s", order.CustomerID)
	logger.Info("  Total Amount: %.2f", order.TotalAmount)
	logger.Info("  Shipping Address: %s, %s, %s, %s, %s",
		order.ShippingAddr.Street,
		order.ShippingAddr.City,
		order.ShippingAddr.State,
		order.ShippingAddr.PostalCode,
		order.ShippingAddr.Country,
	)
	logger.Info("  Items:")
	for _, item := range order.Items {
		logger.Info("    - SKU: %s, Quantity: %d, Price: %.2f", item.SKU, item.Quantity, item.Price)
	}

	// 3. Create WMS payload (simulating what would be sent to actual WMS)
	wmsPayload := MockWmsPayload{
		ShipmentID:      wmsShipmentID,
		OrderID:         order.OrderID,
		CustomerID:      order.CustomerID,
		Items:           order.Items,
		ShippingAddress: order.ShippingAddr,
		TotalAmount:     order.TotalAmount,
		SubmittedAt:     submittedAt,
	}

	// 4. Simulate WMS API integration
	logger.Info("Simulating WMS API integration...")
	logger.Info("Mock WMS API Request Payload:")
	payloadJSON, _ := json.MarshalIndent(wmsPayload, "  ", "  ")
	logger.Info("  %s", string(payloadJSON))

	// Mock successful WMS API response
	logger.Info("Mock WMS API Response: 200 OK - Shipment %s created successfully", wmsShipmentID)

	// 5. Update order in DynamoDB with shipment status and WMS ID
	logger.Info("Updating order %s in DynamoDB: shipment_status -> WMS_ACCEPTED", order.OrderID)
	err := orderRepo.UpdateShipmentStatus(ctx, order.OrderID, models.ShipmentWmsAccepted, wmsShipmentID, submittedAt)
	if err != nil {
		logger.Error("Failed to update order %s in DynamoDB: %v", order.OrderID, err)
		return WmsIntegrationResponse{}, err
	}

	logger.Info("Successfully updated order %s with WMS shipment ID: %s", order.OrderID, wmsShipmentID)

	// 6. Update local order object with shipment status for response
	order.ShipmentStatus = models.ShipmentWmsAccepted

	// 7. Return response for Step Functions
	response := WmsIntegrationResponse{
		Order:                  order,
		WmsShipmentID:          wmsShipmentID,
		WmsSubmissionTimestamp: submittedAt,
		WmsAcceptedAt:          submittedAt,
	}

	logger.Info("WMS integration dispatch completed successfully for OrderID: %s", order.OrderID)
	return response, nil
}

func main() {
	lambda.Start(Handler)
}
