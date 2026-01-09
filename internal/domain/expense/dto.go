package expense

import "time"

type CreateExpenseRequestDTO struct {
	ExpenseCategory string    `json:"expenseCategory" validate:"required,min=3"`
	Amount          int       `json:"amount" validate:"required,min=1"`
	Description     string    `json:"description" validate:"required,min=3"`
	ExpenseDate     time.Time `json:"expenseDate" validate:"required"`
}
