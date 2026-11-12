# Commission Calculator

Сервис расчёта комиссии за переводы с учётом типа перевода, суммы и типа подписки.

## Содержание

- [Обзор](#обзор)
- [Модель данных](#модель-данных)
- [Калькуляторы](#калькуляторы)
- [Ошибки](#ошибки)
- [Результат расчёта](#результат-расчёта)
- [Валидация и лимиты](#валидация-и-лимиты)
- [Использование](#использование)
- [Тесты](#тесты)
- [Структура проекта](#структура-проекта)

---

## Обзор

Сервис принимает `TransferRequest` (тип перевода, сумма, подписка), выбирает нужный калькулятор через фабрику `calculator.New(transferType)` и возвращает сумму комиссии либо ошибку.

Поддерживаемые типы переводов:

- **internal** — внутренний перевод
- **international** — международный перевод

Поддерживаемые подписки:

- **standard** — базовая
- **premium** — премиум (пониженная или нулевая комиссия)

---

## Модель данных

### `TransferRequest`

```go
type TransferRequest struct {
    TransferType TransferType
    Amount       float64
    Subscription Subscription
}
```

- `TransferType` — string-константа, определена в пакете `model`.
- `Subscription` — string-константа, определена в пакете `model`.
- Все допустимые значения заданы константами в `model` (см. ниже).

### Константы (`model`)

```go
// Типы переводов
const (
    TransferTypeInternal      TransferType = "internal"
    TransferTypeInternational TransferType = "international"
)

// Типы подписок
const (
    SubscriptionStandard Subscription = "standard"
    SubscriptionPremium  Subscription = "premium"
)
```

### `TransferRequest.Validate()`

Метод валидирует запрос. На текущем этапе проверяется **только сумма**:

- `amount <= 0` → возвращается `model.ErrAmountNotPositive`.

Проверки `TransferType` и `Subscription` выполняются соответственно в фабрике и в логике калькулятора. **Важно:** метод `Calculate` сам по себе `TransferType` не проверяет — за корректный выбор реализации отвечает вызывающая сторона (обычно через `calculator.New`).

---

## Калькуляторы

### Фабрика

```go
func New(transferType model.TransferType) (Calculator, error)
```

- `internal`      → `InternalCalculator{}`
- `international` → `InternationalCalculator{}`
- любой другой тип → `nil, fmt.Errorf("%w: %q", ErrUnknownTransferType, transferType)`

### Интерфейс

```go
type Calculator interface {
    Calculate(req model.TransferRequest) (float64, error)
}
```

Каждая реализация:

1. Вызывает `req.Validate()`.
2. Проверяет лимит по сумме.
3. Считает комиссию по формуле, зависящей от подписки.

### InternalCalculator

- Лимит: **100 000** (при `amount > limit` → `ErrLimitExceeded` + `LimitExceededError{Limit, Amount}`).
- Формулы:
  - `standard`: `commission = 10 + amount * 0.005`
  - `premium`:  `commission = 0`

### InternationalCalculator

- Лимит: **500 000** (при `amount > limit` → `ErrLimitExceeded` + `LimitExceededError{Limit, Amount}`).
- Формулы:
  - `standard`: `commission = 100 + amount * 0.015`
  - `premium`:  `commission = 100 + amount * 0.010`

---

## Ошибки

Пакет `calculator` экспортирует:

```go
var ErrLimitExceeded       = errors.New("exceeded the limit")
var ErrUnknownSubscription = errors.New("unknown subscription")
var ErrUnknownTransferType = errors.New("unknown transfer type")

type LimitExceededError struct {
    Limit  float64
    Amount float64
}

func (e LimitExceededError) Error() string
```

Плюс в пакете `model`:

```go
var ErrAmountNotPositive = errors.New("amount must be positive")
```

### Как ошибки «заворачиваются»

Ошибка лимита возвращается как `%w: %w` — обёртка над `ErrLimitExceeded` и `LimitExceededError` одновременно, поэтому работают **и** `errors.Is`, **и** `errors.As`:

```go
if errors.Is(err, calculator.ErrLimitExceeded) {
    var target calculator.LimitExceededError
    if errors.As(err, &target) {
        // target.Limit, target.Amount
    }
}
```

Ошибка неизвестной подписки возвращается как `%w: %q` — проверяется через `errors.Is(err, ErrUnknownSubscription)`.

### Сводная таблица ошибок

| Причина                              | Ошибка                                    | Проверка               |
|--------------------------------------|-------------------------------------------|------------------------|
| `amount <= 0`                        | `model.ErrAmountNotPositive`              | `errors.Is`            |
| `amount > limit`                     | `calculator.ErrLimitExceeded` + `LimitExceededError` | `errors.Is` / `errors.As` |
| Неизвестная подписка                 | `calculator.ErrUnknownSubscription`       | `errors.Is`            |
| Неизвестный `TransferType` (фабрика) | `calculator.ErrUnknownTransferType`       | `errors.Is`            |

---

## Результат расчёта

Возвращается `float64` — сумма комиссии. Сравнивать результат в тестах нужно с эпсилоном (`math.Abs(got-want) < 1e-9`), а не через `==`: `amount * 0.005` даёт хвосты в float64.

---

## Валидация и лимиты

| Тип перевода  | Лимит за операцию | standard               | premium                |
|---------------|-------------------|------------------------|------------------------|
| internal      | 100 000           | `10 + amount * 0.005`  | `0`                    |
| international | 500 000           | `100 + amount * 0.015` | `100 + amount * 0.010` |

Порядок проверок в `Calculate`:

1. `req.Validate()` — ошибка, если `amount <= 0`.
2. Проверка лимита — ошибка, если `amount > limit`.
3. Расчёт комиссии по формуле выбранной подписки (неизвестная подписка → `ErrUnknownSubscription`).

---

## Использование

`main` выполняет полный сценарий:

1. Собирает `TransferRequest`.
2. Получает калькулятор: `calculator.New(req.TransferType)`.
3. Вызывает `Calculate(req)`.
4. Печатает результат или ошибку, различая случаи через `errors.Is` / `errors.As` (switch).

Пример:

```go
req := model.TransferRequest{
    TransferType: model.TransferTypeInternational,
    Amount:       1000,
    Subscription: model.SubscriptionStandard,
}

calc, err := calculator.New(req.TransferType)
if err != nil {
    log.Fatal(err)
}

res, err := calc.Calculate(req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Commission: %.2f\n", res)
```

`main` дополнительно демонстрирует:

- валидные запросы для всех 4 комбинаций тип × подписка;
- ошибки: `amount <= 0`, превышение лимита, неизвестная подписка;
- ошибку фабрики на `TransferType("crypto")` с проверкой `errors.Is(err, ErrUnknownTransferType)`.


go run ./cmd/app - запуск тестов
---

## Тесты

Тесты лежат в `calculator/calculator_test.go` с `package calculator_test` — снаружи, чтобы проверять контракт, а не внутренности (значения `internalLimit`/`internationalLimit` не нужны: превышение проверяется через поведение).
Тесты лежат в `model/model_test.go`, `factory_test.go`, 

Покрыто:

- **`TestCalculate_HappyPath`** — табличный тест для обеих реализаций × `standard`/`premium`, поле `calc calculator.Calculator` в кейсе (проверяет, что обе реализации удовлетворяют интерфейсу). Сравнение через `math.Abs(got-want) < 1e-9`.
- **`TestCalculate_Errors`** — `amount <= 0` (в т.ч. отрицательное) → `ErrAmountNotPositive`; превышение лимита для обеих реализаций → `errors.Is(ErrLimitExceeded)` + `errors.As(LimitExceededError)` с проверкой `Limit`/`Amount`; неизвестная подписка → `ErrUnknownSubscription`.
- **`TestImplementationsDiffer`** — одинаковая сумма/подписка даёт разные комиссии у `Internal` и `International` (доказывает, что реализации не декоративные).
- **`TestNew`** — фабрика: `internal`/`international` → нужный конкретный тип; неизвестный и пустой `TransferType` → `ErrUnknownTransferType` и `nil`-калькулятор.
- **`TestNew_ReturnsUsableCalculator`** — калькулятор из фабрики реально считает и совпадает с прямым вызовом конкретной реализации.

Запуск:

```bash
go test ./...
go test -v ./calculator
```

---

## Структура проекта

```text
.
├── cmd/app/main.go
├── model/
│   ├── model.go              // типы, константы, Validate(), ErrAmountNotPositive
|   └── model_test.go         // package model_test
└── calculator/
    ├── interface.go          // Calculator
    ├── factory.go            // New
    ├── internal.go           // InternalCalculator
    ├── international.go      // InternationalCalculator
    ├── errors.go             // ErrLimitExceeded, ErrUnknownSubscription,
    │                         // ErrUnknownTransferType, LimitExceededError
    ├── calculator_test.go    // package calculator_test
    └── factory_test.go       // package calculator_test
```


### Защита

1. Почему задача?
Тезис: расчёт тарифа — близка (агенты для юрлиц), есть естественные вариации (тип перевода, подписка) → интерфейс оправдан, не «ради темы».

2. Зачем интерфейс?
Без Calculator: main знал бы обе реализации, switch в main, дублирование.

С интерфейсом: run зависит от Calculator, не от конкретных типов; добавление реализации — только фабрика.

3. Почему Calculate — один метод, а Validate — не в интерфейсе?
Validate — свойство запроса (TransferRequest), не поведение калькулятора; предусловие для любой реализации → живёт в model.

Если бы Validate был в Calculator — каждая реализация повторяла бы валидацию → дублирование. Сейчас Calculate вызывает req.Validate() первым — обе получают бесплатно.

4. Чем отличаются реализации?
Формулы (10 + 0.5% vs 100 + 1.5%), лимиты (100k vs 500k), premium (0 vs пониженный %). Не «имена файлов» — поведение разное (TestImplementationsDiffer).

5. Error flow — кто оборачивает, кто различает?
Calculate — оборачивает (%w — контекст: лимит/сумма/подписка). %w: %w — потому что нужны и errors.Is (ErrLimitExceeded), и errors.As (LimitExceededError) на одной ошибке.

run (в main) — различает через errors.Is/errors.As и реагирует по-разному (сообщение/числа). Не «добавляет контекст» — читает его.

Typed (LimitExceededError) лучше сентинела здесь, потому что нужны числа (Limit/Amount) для сообщения — сентинел даёт только факт.

6. Самый важный trade-off?
Фабрика скрывает выбор; run — интерфейс → новая реализация — только case в фабрике, run не меняется.

Value receiver — объекты маленькие, без мутации; если *T — значение не реализует Calculator, фабрика должна возвращать &... — согласованность важнее.

Typed error — сильнее сентинела (числа), но сложнее (%w: %w, errors.As).