package calculator

import (
	"commission-calculator/model"
	"fmt"
)

type InternationalCalculator struct{}

const internationalLimit = 500_000

func (c InternationalCalculator) Calculate(req model.TransferRequest) (float64, error) {
	if err := req.Validate(); err != nil {
		return 0, err
	}

	if req.Amount > internationalLimit {
		return 0, fmt.Errorf("%w: %w", ErrLimitExceeded, LimitExceededError{Limit: internationalLimit, Amount: req.Amount})
	}

	var commission float64
	switch req.Subscription {
	case model.SubscriptionStandard:
		commission = 100 + req.Amount*0.015
	case model.SubscriptionPremium:
		commission = 100 + req.Amount*0.010
	default:
		return 0, fmt.Errorf("%w: %q", ErrUnknownSubscription, req.Subscription)
	}
	return commission, nil
}
