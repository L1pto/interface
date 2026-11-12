package calculator_test

import (
	"errors"
	"math"
	"testing"

	"commission-calculator/calculator"
	"commission-calculator/model"
)

const eps = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < eps
}

// helper: собирает запрос
func req(tt model.TransferType, sub model.Subscription, amount float64) model.TransferRequest {
	return model.TransferRequest{TransferType: tt, Amount: amount, Subscription: sub}
}

func TestCalculate_HappyPath(t *testing.T) {
	tests := []struct {
		name string
		calc calculator.Calculator
		req  model.TransferRequest
		want float64
	}{
		{
			name: "internal/standard",
			calc: calculator.InternalCalculator{},
			req:  req(model.TransferTypeInternal, model.SubscriptionStandard, 3000),
			want: 10 + 3000*0.005, // 25
		},
		{
			name: "internal/premium",
			calc: calculator.InternalCalculator{},
			req:  req(model.TransferTypeInternal, model.SubscriptionPremium, 3000),
			want: 0,
		},
		{
			name: "international/standard",
			calc: calculator.InternationalCalculator{},
			req:  req(model.TransferTypeInternational, model.SubscriptionStandard, 3000),
			want: 100 + 3000*0.015, // 145
		},
		{
			name: "international/premium",
			calc: calculator.InternationalCalculator{},
			req:  req(model.TransferTypeInternational, model.SubscriptionPremium, 3000),
			want: 100 + 3000*0.010, // 130
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.calc.Calculate(tt.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !almostEqual(got, tt.want) {
				t.Errorf("commission = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculate_Errors(t *testing.T) {
	tests := []struct {
		name      string
		calc      calculator.Calculator
		req       model.TransferRequest
		wantErr   error   // проверяем через errors.Is
		wantLimit float64 // для LimitExceededError, 0 если не проверяем
	}{
		{
			name:    "internal/amount not positive",
			calc:    calculator.InternalCalculator{},
			req:     req(model.TransferTypeInternal, model.SubscriptionStandard, 0),
			wantErr: model.ErrAmountNotPositive,
		},
		{
			name:    "internal/negative amount",
			calc:    calculator.InternalCalculator{},
			req:     req(model.TransferTypeInternal, model.SubscriptionStandard, -10),
			wantErr: model.ErrAmountNotPositive,
		},
		{
			name:      "internal/limit exceeded",
			calc:      calculator.InternalCalculator{},
			req:       req(model.TransferTypeInternal, model.SubscriptionStandard, 100_001),
			wantErr:   calculator.ErrLimitExceeded,
			wantLimit: 100_000,
		},
		{
			name:      "international/limit exceeded",
			calc:      calculator.InternationalCalculator{},
			req:       req(model.TransferTypeInternational, model.SubscriptionStandard, 500_001),
			wantErr:   calculator.ErrLimitExceeded,
			wantLimit: 500_000,
		},
		{
			name:    "internal/unknown subscription",
			calc:    calculator.InternalCalculator{},
			req:     req(model.TransferTypeInternal, model.Subscription("vip"), 100),
			wantErr: calculator.ErrUnknownSubscription,
		},
		{
			name:    "international/unknown subscription",
			calc:    calculator.InternationalCalculator{},
			req:     req(model.TransferTypeInternational, model.Subscription("vip"), 100),
			wantErr: calculator.ErrUnknownSubscription,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.calc.Calculate(tt.req)
			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.wantErr)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("errors.Is(err, %v) = false; err = %v", tt.wantErr, err)
			}

			if tt.wantLimit != 0 {
				var target calculator.LimitExceededError
				if !errors.As(err, &target) {
					t.Fatalf("errors.As LimitExceededError = false; err = %v", err)
				}
				if !almostEqual(target.Limit, tt.wantLimit) {
					t.Errorf("target.Limit = %v, want %v", target.Limit, tt.wantLimit)
				}
				if !almostEqual(target.Amount, tt.req.Amount) {
					t.Errorf("target.Amount = %v, want %v", target.Amount, tt.req.Amount)
				}
			}
		})
	}
}

// TestImplementationsDiffer доказывает, что реализации не декоративные:
// на одинаковом запросе они дают разные комиссии
func TestImplementationsDiffer(t *testing.T) {
	r := req(model.TransferTypeInternal, model.SubscriptionStandard, 3000)

	internal := calculator.InternalCalculator{}
	international := calculator.InternationalCalculator{}

	gotInternal, err := internal.Calculate(r)
	if err != nil {
		t.Fatalf("internal: unexpected error: %v", err)
	}
	gotInternational, err := international.Calculate(r)
	if err != nil {
		t.Fatalf("international: unexpected error: %v", err)
	}

	if almostEqual(gotInternal, gotInternational) {
		t.Fatalf("implementations returned same commission %v; they must differ", gotInternal)
	}
}
