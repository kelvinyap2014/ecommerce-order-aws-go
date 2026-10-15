package storage

import (
	"context"
	"ecommerce-order-aws-go/internal/models"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// OrderRepository defines the interface for order data operations.
type OrderRepository interface {
	SaveOrder(ctx context.Context, order models.Order) error
	GetOrder(ctx context.Context, orderID string) (*models.Order, error)
	ListOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error
	UpdateShipmentStatus(ctx context.Context, orderID string, shipmentStatus models.ShipmentStatus, wmsShipmentID string, wmsAcceptedAt string) error
}

// DynamoOrderRepo implements OrderRepository using Amazon DynamoDB.
type DynamoOrderRepo struct {
	Client    *dynamodb.Client
	TableName string
}

// NewDynamoOrderRepo creates a new DynamoDB order repository instance.
func NewDynamoOrderRepo(client *dynamodb.Client, tableName string) *DynamoOrderRepo {
	return &DynamoOrderRepo{
		Client:    client,
		TableName: tableName,
	}
}

// SaveOrder persists an order into DynamoDB with idempotency protection.
func (r *DynamoOrderRepo) SaveOrder(ctx context.Context, order models.Order) error {
	item, err := attributevalue.MarshalMap(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	_, err = r.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(OrderId)"),
	})
	if err != nil {
		return fmt.Errorf("failed to put item in dynamodb: %w", err)
	}

	return nil
}

// GetOrder retrieves an order by its unique OrderID.
func (r *DynamoOrderRepo) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	out, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"OrderId": &types.AttributeValueMemberS{Value: orderID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item from dynamodb: %w", err)
	}

	if out.Item == nil {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	var order models.Order
	err = attributevalue.UnmarshalMap(out.Item, &order)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}

// ListOrders retrieves all orders from the DynamoDB table.
func (r *DynamoOrderRepo) ListOrders(ctx context.Context) ([]models.Order, error) {
	out, err := r.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.TableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan orders: %w", err)
	}

	var orders []models.Order
	err = attributevalue.UnmarshalListOfMaps(out.Items, &orders)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal orders: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus modifies only the Status attribute of an order in DynamoDB.
func (r *DynamoOrderRepo) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) error {
	_, err := r.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"OrderId": &types.AttributeValueMemberS{Value: orderID},
		},
		UpdateExpression:         aws.String("SET #st = :status"),
		ExpressionAttributeNames: map[string]string{"#st": "Status"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update order status to %s: %w", status, err)
	}
	return nil
}

// UpdateShipmentStatus updates shipment-related fields (ShipmentStatus, WmsShipmentId, WmsAcceptedAt) on an existing order.
func (r *DynamoOrderRepo) UpdateShipmentStatus(ctx context.Context, orderID string, shipmentStatus models.ShipmentStatus, wmsShipmentID string, wmsAcceptedAt string) error {
	_, err := r.Client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]types.AttributeValue{
			"OrderId": &types.AttributeValueMemberS{Value: orderID},
		},
		UpdateExpression: aws.String("SET ShipmentStatus = :shipmentStatus, WmsShipmentId = :wmsShipmentID, WmsAcceptedAt = :wmsAcceptedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":shipmentStatus": &types.AttributeValueMemberS{Value: string(shipmentStatus)},
			":wmsShipmentID":  &types.AttributeValueMemberS{Value: wmsShipmentID},
			":wmsAcceptedAt":  &types.AttributeValueMemberS{Value: wmsAcceptedAt},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update shipment status for order %s: %w", orderID, err)
	}
	return nil
}
