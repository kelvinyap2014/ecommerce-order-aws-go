package models

type OrderStatus string

const (
	StatusPending       OrderStatus = "PENDING"
	StatusVerified      OrderStatus = "VERIFIED"
	StatusAddressFailed OrderStatus = "ADDRESS_FAILED"
	StatusCreditFailed  OrderStatus = "CREDIT_FAILED"
	StatusPaymentFailed OrderStatus = "PAYMENT_FAILED"
	StatusCancelled     OrderStatus = "CANCELLED"
)

type ShipmentStatus string

const (
	ShipmentNotShipped  ShipmentStatus = "NOT_SHIPPED"
	ShipmentWmsAccepted ShipmentStatus = "WMS_ACCEPTED"
	ShipmentShipped     ShipmentStatus = "SHIPPED"
	ShipmentInTransit   ShipmentStatus = "IN_TRANSIT"
	ShipmentDelivered   ShipmentStatus = "DELIVERED"
	ShipmentReturned    ShipmentStatus = "RETURNED"
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
	OrderID        string         `json:"order_id" dynamodbav:"OrderId"`
	CustomerID     string         `json:"customer_id" dynamodbav:"CustomerId"`
	Items          []OrderItem    `json:"items" dynamodbav:"Items"`
	ShippingAddr   Address        `json:"shipping_address" dynamodbav:"ShippingAddress"`
	TotalAmount    float64        `json:"total_amount" dynamodbav:"TotalAmount"`
	Status         OrderStatus    `json:"status" dynamodbav:"Status"`
	ShipmentStatus ShipmentStatus `json:"shipment_status" dynamodbav:"ShipmentStatus"`
	CreatedAt      string         `json:"created_at" dynamodbav:"CreatedAt"`
}
