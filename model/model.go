package model

import "errors"

type TransferType string

type Subscription string

var ErrAmountNotPositive = errors.New("amount must be positive")

const (
	TransferTypeInternal      TransferType = "internal"
	TransferTypeInternational TransferType = "international"
)

const (
	SubscriptionStandard Subscription = "standard"
	SubscriptionPremium  Subscription = "premium"
)

type TransferRequest struct {
	TransferType TransferType
	Amount       float64
	Subscription Subscription
}

// Validate проверяет, что Amount > 0.
func (req TransferRequest) Validate() error {
	if req.Amount <= 0 {
		return ErrAmountNotPositive
	}
	return nil
}
