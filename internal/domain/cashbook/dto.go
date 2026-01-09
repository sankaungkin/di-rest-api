package cashbook

import (
	"time"
)

type CreaateCashbookEntryRequestDTO struct {
	TransactionDate time.Time `gorm:"type:timestamptz;index" json:"transactionDate"`
	TransactionType string    `json:"transactionType"` // SALE, PURCHASE, SALE_RETURN, EXPENSE, OWNER_INJECTION
	ReferenceID     string    `json:"referenceId"`     // ID from Sale, Purchase, or Expense
	Description     string    `json:"description"`
	Debit           int64     `json:"debit"`   // Cash In (+)
	Credit          int64     `json:"credit"`  // Cash Out (-)
	Balance         int64     `json:"balance"` // Running Balance
	CreatedAt       time.Time `json:"createdAt"`
}
