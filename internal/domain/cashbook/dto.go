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
	TotalSale            int64 `json:"totalSale"`
	TotalWithdrawals     int64 `json:"totalWithdrawals"`
	TotalNewDebt         int64 `json:"totalNewDebt"`
	TotalCashSales       int64 `json:"totalCashSales"`
	TotalKPaySales       int64 `json:"totalKPaySales"`
	TotalDebtCollected   int64 `json:"totalDebtCollected"`
	TotalKPayCollected   int64 `json:"totalKPayCollected"`
	TotalCashInflow      int64 `json:"totalCashInflow"`
	TotalPurchase        int64 `json:"totalPurchase"`
	TotalPayables        int64 `json:"totalPayables"` // Will now hold 2891100
	TotalSupplierPaid    int64 `json:"totalSupplierPaid"`
	TotalReceivables     int64 `json:"totalReceivables"`
	OpeningBalance       int64 `json:"openingBalance"`
	ClosingBalance       int64 `json:"closingBalance"`
	CurrentDrawerBalance int64 `json:"currentDrawerBalance"`
}
