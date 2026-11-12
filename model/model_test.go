package model_test

import (
	"commission-calculator/model"
	"errors"
	"testing"
)

func TestTransferRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr error // nil или model.ErrAmountNotPositive
	}{
		{"positive", 100, nil},
		{"zero", 0, model.ErrAmountNotPositive},
		{"negative", -5, model.ErrAmountNotPositive},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := model.TransferRequest{Amount: tt.amount}
			err := req.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ошибка %s, ожидалась %s", err, tt.wantErr)
			}
		})
	}
}
