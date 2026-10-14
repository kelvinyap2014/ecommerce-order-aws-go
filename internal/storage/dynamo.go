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
