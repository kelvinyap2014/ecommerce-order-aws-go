package models

type OrderStatus string

const (
	StatusPending       OrderStatus = "PENDING"
	StatusVerified      OrderStatus = "VERIFIED"
	StatusPaymentFailed OrderStatus = "PAYMENT_FAILED"
	StatusConfirmed     OrderStatus = "CONFIRMED"
	StatusCancelled     OrderStatus = "CANCELLED"
)

type Address struct {
	Street     string `json:"street" dynamodbav:"Street"`
	City       string `json:"city" dynamodbav:"City"`
	State      string `json:"state" dynamodbav:"State"`
	PostalCode string `json:"postal_code" dynamodbav:"PostalCode"`
	Country    string `json:"country" dynamodbav:"Country"`
}

type OrderItem struct {
	SKU      string  `json:"sku" dynamodbav:"SKU"`
	Quantity int     `json:"quantity" dynamodbav:"Quantity"`
	Price    float64 `json:"price" dynamodbav:"Price"`
}

type Order struct {
	OrderID      string      `json:"order_id" dynamodbav:"OrderId"`
	CustomerID   string      `json:"customer_id" dynamodbav:"CustomerId"`
	Items        []OrderItem `json:"items" dynamodbav:"Items"`
	ShippingAddr Address     `json:"shipping_address" dynamodbav:"ShippingAddress"`
	TotalAmount  float64     `json:"total_amount" dynamodbav:"TotalAmount"`
	Status       OrderStatus `json:"status" dynamodbav:"Status"`
	CreatedAt    string      `json:"created_at" dynamodbav:"CreatedAt"`
}
