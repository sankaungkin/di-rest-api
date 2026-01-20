package cashbook

import (
	"time"
)

type CreaateCashbookEntryRequestDTO struct {
	TransactionDate time.Time `gorm:"type:timestamptz;index" json:"transactionDate"`
	TransactionType string    `json:"transactionType"` // SALE, PURCHASE, SALE_RETURN, EXPENSE, OWNER_INJECTION
	ReferenceID     string    `json:"referenceId"`     // ID from Sale, Purchase, or Expense
	Description     string    `json:"description"`
	PaymentMethod   string    `json:"paymentMethod"` // CASH, CREDIT_CARD, BANK_TRANSFER, etc.
	Debit           int64     `json:"debit"`         // Cash In (+)
	Credit          int64     `json:"credit"`        // Cash Out (-)
	Balance         int64     `json:"balance"`       // Running Balance
	CreatedAt       time.Time `json:"createdAt"`
}

type DashboardSummary struct {
	OpeningBalance     int64 `json:"openingBalance"`
	TotalPurchase      int64 `json:"totalPurchase"` // Total volume
	TotalSale          int64 `json:"totalSale"`     // Total volume
	TotalNewDebt       int64 `json:"totalNewDebt"`  // Sales on credit today
	TotalCashSales     int64 `json:"totalCashSales"`
	TotalKPaySales     int64 `json:"totalKpaySales"`
	TotalDebtCollected int64 `json:"totalDebtCollected"` // Money recovered
	ClosingBalance     int64 `json:"closingBalance"`     // Actual cash position
}
