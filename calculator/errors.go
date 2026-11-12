package calculator

import (
	"errors"
	"fmt"
)

var ErrLimitExceeded = errors.New("exceeded the limit")

var ErrUnknownSubscription = errors.New("unknown subscription")

var ErrUnknownTransferType = errors.New("unknown transfer type")

type LimitExceededError struct {
	Limit  float64
	Amount float64
}

func (e LimitExceededError) Error() string {
	return fmt.Sprintf("amount %.2f exceeds limit %.2f", e.Amount, e.Limit)
}
