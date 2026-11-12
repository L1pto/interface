package main

import (
	"commission-calculator/calculator"
	"commission-calculator/model"
	"errors"
	"fmt"
)

func main() {
	valid := []model.TransferRequest{
		{TransferType: model.TransferTypeInternational, Amount: 1000, Subscription: model.SubscriptionStandard},
		{TransferType: model.TransferTypeInternational, Amount: 1000, Subscription: model.SubscriptionPremium},
		{TransferType: model.TransferTypeInternal, Amount: 3000, Subscription: model.SubscriptionStandard},
		{TransferType: model.TransferTypeInternal, Amount: 3000, Subscription: model.SubscriptionPremium},
	}

	fmt.Println("=== Валидные запросы ===")
	for _, req := range valid {
		runByType(req)
	}

	fmt.Println("\n=== Ошибки ===")
	invalid := []model.TransferRequest{
		{TransferType: model.TransferTypeInternal, Amount: 0, Subscription: model.SubscriptionStandard},

		{TransferType: model.TransferTypeInternal, Amount: 150_000, Subscription: model.SubscriptionStandard},

		{TransferType: model.TransferTypeInternational, Amount: 500, Subscription: model.Subscription("gold")},
	}
	for _, req := range invalid {
		runByType(req)
	}

	fmt.Println("\n=== Ошибка фабрики ===")
	if _, err := calculator.New(model.TransferType("crypto")); err != nil {
		fmt.Printf("[crypto] factory error: %v\n", err)
		fmt.Printf("  errors.Is(ErrUnknownTransferType) = %t\n",
			errors.Is(err, calculator.ErrUnknownTransferType))
	}
}

func runByType(req model.TransferRequest) {
	calc, err := calculator.New(req.TransferType)
	if err != nil {
		fmt.Printf("[%s/%s %.2f] factory error: %v\n",
			req.TransferType, req.Subscription, req.Amount, err)
		return
	}
	run(calc, req)
}

func run(calc calculator.Calculator, req model.TransferRequest) {
	res, err := calc.Calculate(req)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrAmountNotPositive):
			fmt.Printf("%v\n", err)
		case errors.Is(err, calculator.ErrLimitExceeded):
			var target calculator.LimitExceededError
			if errors.As(err, &target) {
				fmt.Printf("exceeded the limit: amount %.2f, limit %.2f\n", target.Amount, target.Limit)
			}
		case errors.Is(err, calculator.ErrUnknownSubscription):
			fmt.Printf("%v\n", err)
		default:
			fmt.Printf("[%s/%s %.2f] error: %v\n",
				req.TransferType, req.Subscription, req.Amount, err)

		}
		return
	}
	fmt.Printf("[%s/%s %.2f] commission: %.2f\n",
		req.TransferType, req.Subscription, req.Amount, res)
}
