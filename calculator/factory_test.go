package calculator_test

import (
	"commission-calculator/calculator"
	"commission-calculator/model"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		transferType model.TransferType
		wantErr      error
		wantCalcType any // ожидаемый конкретный тип, nil если ошибка
	}{
		{
			name:         "internal",
			transferType: model.TransferTypeInternal,
			wantCalcType: calculator.InternalCalculator{},
		},
		{
			name:         "international",
			transferType: model.TransferTypeInternational,
			wantCalcType: calculator.InternationalCalculator{},
		},
		{
			name:         "unknown transfer type",
			transferType: model.TransferType("crypto"),
			wantErr:      calculator.ErrUnknownTransferType,
		},
		{
			name:         "empty transfer type",
			transferType: model.TransferType(""),
			wantErr:      calculator.ErrUnknownTransferType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc, err := calculator.New(tt.transferType)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("errors.Is(err, %v) = false; err = %v", tt.wantErr, err)
				}
				if calc != nil {
					t.Errorf("calc = %v, want nil on error", calc)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if calc == nil {
				t.Fatal("calc = nil, want non-nil")
			}
			switch want := tt.wantCalcType.(type) {
			case calculator.InternalCalculator:
				if _, ok := calc.(calculator.InternalCalculator); !ok {
					t.Errorf("calc type = %T, want %T", calc, want)
				}
			case calculator.InternationalCalculator:
				if _, ok := calc.(calculator.InternationalCalculator); !ok {
					t.Errorf("calc type = %T, want %T", calc, want)
				}
			}
		})
	}
}

// TestNew_ReturnsUsableCalculator — фабрика возвращает рабочий Calculator:
// результат совпадает с прямым вызовом конкретной реализации.
func TestNew_ReturnsUsableCalculator(t *testing.T) {
	r := req(model.TransferTypeInternal, model.SubscriptionStandard, 3000)

	fromFactory, err := calculator.New(model.TransferTypeInternal)
	if err != nil {
		t.Fatalf("New: unexpected error: %v", err)
	}

	got, err := fromFactory.Calculate(r)
	if err != nil {
		t.Fatalf("Calculate: unexpected error: %v", err)
	}

	want, err := calculator.InternalCalculator{}.Calculate(r)
	if err != nil {
		t.Fatalf("direct Calculate: unexpected error: %v", err)
	}

	if !almostEqual(got, want) {
		t.Errorf("factory commission = %v, direct = %v", got, want)
	}
}
