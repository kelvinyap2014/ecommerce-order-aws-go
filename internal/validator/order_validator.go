package validator

import (
	"ecommerce-order-aws-go/internal/models"
	"errors"
)

var (
	ErrEmptyCustomerID = errors.New("customer ID is required")
	ErrNoItems         = errors.New("order must contain at least one item")
	ErrInvalidAmount   = errors.New("total amount must be greater than zero")
	ErrInvalidAddress  = errors.New("shipping address country and city are required")
)

func ValidateOrder(order models.Order) error {
	if order.CustomerID == "" {
		return ErrEmptyCustomerID
	}
	if len(order.Items) == 0 {
		return ErrNoItems
	}
	if order.TotalAmount <= 0 {
		return ErrInvalidAmount
	}
	if order.ShippingAddr.Country == "" || order.ShippingAddr.City == "" {
		return ErrInvalidAddress
	}
	return nil
}
