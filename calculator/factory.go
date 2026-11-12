package calculator

import (
	"commission-calculator/model"
	"fmt"
)

func New(transferType model.TransferType) (Calculator, error) {
	switch transferType {
	case model.TransferTypeInternal:
		return InternalCalculator{}, nil
	case model.TransferTypeInternational:
		return InternationalCalculator{}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownTransferType, transferType)
	}
}
