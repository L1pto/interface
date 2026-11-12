package calculator

import (
	"commission-calculator/model"
	"fmt"
)

type InternalCalculator struct{}

const internalLimit = 100_000

func (c InternalCalculator) Calculate(req model.TransferRequest) (float64, error) {
	if err := req.Validate(); err != nil {
		return 0, err
	}

	if req.Amount > internalLimit {
		return 0, fmt.Errorf("%w: %w", ErrLimitExceeded, LimitExceededError{Limit: internalLimit, Amount: req.Amount})
	}
	var commission float64
	switch req.Subscription {
	case model.SubscriptionStandard:
		commission = 10 + req.Amount*0.005
	case model.SubscriptionPremium:
		commission = 0
	default:
		return 0, fmt.Errorf("%w: %q", ErrUnknownSubscription, req.Subscription)
	}
	return commission, nil
}
