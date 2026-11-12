package calculator

import "commission-calculator/model"

type Calculator interface {
	// Calculate возвращает ошибку при невалидном запросе, превышении лимита или неизвестной подписке.
	Calculate(req model.TransferRequest) (float64, error)
}
