package cashbook

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SettlementReport struct {
	Date           time.Time `json:"date"`
	OpeningBalance int64     `json:"openingBalance"`
	ClosingBalance int64     `json:"closingBalance"`
	CashWithdrawn  int64     `json:"cashWithdrawn"` // Excess cash moved to safe
}

type CashbookRepositoryInterface interface {
	GetBalance() (int64, error)
	GetCashBalance() (int64, error)
	GetAll(startDate, endDate string) ([]models.Cashbook, error)
	GetLedger(startDate, endDate string) ([]models.Cashbook, error)
	CloseDay(today time.Time) error
	CreateEntry(tx *gorm.DB, entry *models.Cashbook) error
	GetSettlementReport(month int, year int) ([]SettlementReport, error)
	GetDashboardSummary() (*DashboardSummary, error)
	GetPastSummaries() ([]models.DailySummaries, error)
	GetCurrentDrawerBalance() (int64, error)
	ReconcileToday() ([]map[string]interface{}, error)
	RecordOwnerWithdrawal(amount int64, description string) error
	GetEntriesByDateAndType(date time.Time, transType string) ([]models.Cashbook, error)
}

type CashbookRepository struct {
	db *gorm.DB
}

var (
	repoInstance *CashbookRepository
	repoOnce     sync.Once
)

func NewCashbookRepository(db *gorm.DB) CashbookRepositoryInterface {
	repoOnce.Do(func() {
		repoInstance = &CashbookRepository{db: db}
	})
	return repoInstance
}

// GetBalance returns the balance calculated from today's opening record
func (r *CashbookRepository) GetCashBalance() (int64, error) {
	todayStr := time.Now().Format("2006-01-02")
	fixedOpeningBalance := int64(5000)

	var totalCashIn int64  // Physical cash from Sales
	var totalCashOut int64 // Physical cash used for Purchases/Expenses

	// 1. Sum all CASH IN (Only where payment_method is CASH)
	// We include OWNER_INJECTION because that is physical cash put into the drawer
	err := r.db.Model(&models.Cashbook{}).
		Where("DATE(transaction_date) = ? AND payment_method = ? AND debit > 0 AND transaction_type = 'SALE'", todayStr, "CASH").
		// Or("DATE(transaction_date) = ? AND transaction_type = ? AND debit > 0", todayStr, "OWNER_INJECTION").
		Select("COALESCE(SUM(debit), 0)").
		Scan(&totalCashIn).Error

	if err != nil {
		return 0, err
	}

	// 2. Sum all CASH OUT (Physical cash spent)
	// We filter for CASH method to ignore purchases made via Bank/KPay if you have those
	err = r.db.Model(&models.Cashbook{}).
		Where("DATE(transaction_date) = ? AND payment_method = ? AND credit > 0", todayStr, "CASH").
		Select("COALESCE(SUM(credit), 0)").
		Scan(&totalCashOut).Error

	if err != nil {
		return 0, err
	}

	// 3. Final Calculation
	// Physical Cash = Start + Cash Sales + Injections - Cash Purchases - Expenses
	actualCashInDrawer := fixedOpeningBalance + totalCashIn - totalCashOut

	return actualCashInDrawer, nil
}
func (r *CashbookRepository) GetBalance() (int64, error) {
	var lastEntry models.Cashbook

	// Get the absolute latest record by ID.
	// In accounting, the last entry in the ledger IS the current balance.
	err := r.db.Order("id desc").First(&lastEntry).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If the shop just opened and has NO records yet
			return 0, nil
		}
		return 0, err
	}

	// Return the balance column directly from the last row
	return lastEntry.Balance, nil
}

// cashbook_repository.go

func (r *CashbookRepository) GetAll(startDate, endDate string) ([]models.Cashbook, error) {
	var entries []models.Cashbook
	query := r.db.Order("id desc")

	// Validate that dates aren't empty before adding to query
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(transaction_date) BETWEEN ? AND ?", startDate, endDate)
	} else {
		// Fallback to today if dates are missing to prevent 400 error
		today := time.Now().Format("2006-01-02")
		query = query.Where("DATE(transaction_date) = ?", today)
	}

	err := query.Find(&entries).Error
	return entries, err
}

// GetLedger returns today's transactions primarily
func (r *CashbookRepository) GetLedger(startDate, endDate string) ([]models.Cashbook, error) {
	var entries []models.Cashbook

	query := r.db.Order("id desc")

	// If dates are provided, filter by range.
	// Otherwise, default to today's transactions.
	if startDate != "" && endDate != "" {
		query = query.Where("DATE(transaction_date) BETWEEN ? AND ?", startDate, endDate)
	} else {
		today := time.Now().Format("2006-01-02")
		query = query.Where("DATE(transaction_date) = ?", today)
	}

	err := query.Find(&entries).Error
	return entries, err
}

func (r *CashbookRepository) CloseDay(today time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		todayStr := today.Format("2006-01-02")
		now := time.Now()

		// 1. Calculate Sales Totals (Cash, KPay, Debt)
		var report struct {
			CashTotal int64
			KPayTotal int64
			DebtTotal int64
			TotalSale int64
		}
		tx.Raw(`
            SELECT 
                COALESCE(SUM(CASE WHEN payment_method = 'CASH' THEN grand_total ELSE 0 END), 0),
                COALESCE(SUM(CASE WHEN payment_method = 'KPAY' THEN grand_total ELSE 0 END), 0),
                COALESCE(SUM(CASE WHEN payment_method = 'DEBT' THEN grand_total ELSE 0 END), 0),
                COALESCE(SUM(grand_total), 0)
            FROM sales 
            WHERE DATE(sale_date) = ?`, todayStr).Row().Scan(
			&report.CashTotal, &report.KPayTotal, &report.DebtTotal, &report.TotalSale,
		)

		// 2. Calculate Cashbook Metrics (Recovered Debt, Purchases, and Manual Withdrawals)
		var metrics struct {
			DebtCollected     int64
			TotalOutflow      int64
			ManualWithdrawals int64
		}
		tx.Raw(`
            SELECT 
        COALESCE(SUM(CASE WHEN transaction_type = 'DEBT_PAYMENT' AND payment_method = 'CASH' THEN debit ELSE 0 END), 0),
        -- 💡 Crucial Fix: Only count outflow if Payment Method is CASH
        COALESCE(SUM(CASE WHEN transaction_type IN ('EXPENSE', 'PURCHASE', 'PURCHASE_PAYMENT') AND payment_method = 'CASH' THEN credit ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN transaction_type = 'OWNER_WITHDRAWAL' AND payment_method = 'CASH' THEN credit ELSE 0 END), 0)
    FROM cashbooks 
    WHERE DATE(transaction_date) = ?`, todayStr).Row().Scan(
			&metrics.DebtCollected, &metrics.TotalOutflow, &metrics.ManualWithdrawals,
		)

		// 3. Get Final Drawer Balance BEFORE Reset
		var lastEntry models.Cashbook
		if err := tx.Order("id desc").Limit(1).Find(&lastEntry).Error; err != nil {
			return err
		}
		actualEndOfDayBalance := lastEntry.Balance

		// 4. Reset Drawer to 5000 (The "Take-Home" Adjustment)
		amountToAdjust := actualEndOfDayBalance - 5000
		finalWithdrawalTotal := metrics.ManualWithdrawals

		if amountToAdjust > 0 {
			tx.Create(&models.Cashbook{
				TransactionDate: now,
				TransactionType: "OWNER_WITHDRAWAL",
				PaymentMethod:   "CASH",
				ReferenceID:     "RESET-" + todayStr,
				Description:     fmt.Sprintf("Daily Reset: Withdrawal to Bank %d", amountToAdjust),
				Credit:          amountToAdjust,
				Balance:         5000,
				CreatedAt:       now,
			})
			// Add the reset amount to the day's total withdrawal figure
			finalWithdrawalTotal += amountToAdjust
		}

		// 5. Update Today's Summary Record
		err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ? AND is_closed = ?", todayStr, false).
			Updates(map[string]interface{}{
				"cash_total":       report.CashTotal,
				"k_pay_total":      report.KPayTotal,
				"debt_total":       report.DebtTotal,
				"debt_collected":   metrics.DebtCollected,
				"expense_total":    metrics.TotalOutflow,
				"total_withdrawal": finalWithdrawalTotal,
				"total_sale":       report.TotalSale,
				"closing_balance":  5000, // Drawer is now reset
				"is_closed":        true,
			}).Error
		if err != nil {
			return err
		}

		// 6. Setup Tomorrow
		tomorrow := today.AddDate(0, 0, 1)
		tomorrowStr := tomorrow.Format("2006-01-02")

		return tx.Where("DATE(summary_date) = ?", tomorrowStr).
			FirstOrCreate(&models.DailySummaries{
				SummaryDate:    tomorrow,
				OpeningBalance: 5000,
				ClosingBalance: 5000,
				IsClosed:       false,
			}).Error
	})
}

// CreateEntry handles the calculation of the running balance and saves the record
func (r *CashbookRepository) CreateEntry(tx *gorm.DB, entry *models.Cashbook) error {

	executor := tx
	if executor == nil {
		executor = r.db
	}
	var lastEntry models.Cashbook

	// 1. Get the absolute latest balance with a LOCK
	// .Clauses(clause.Locking{Strength: "UPDATE"}) prevents other transactions
	// from reading this row until this one is finished.
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id desc").First(&lastEntry).Error

	currentBalance := int64(0)
	if err == nil {
		currentBalance = lastEntry.Balance
	}

	// 2. Calculate new balance
	entry.Balance = currentBalance + entry.Debit - entry.Credit

	// Use the date from the entry if provided (for backdating), otherwise use Now
	if entry.TransactionDate.IsZero() {
		entry.TransactionDate = time.Now()
	}
	if entry.PaymentMethod == "CASH" {
		entry.Balance = currentBalance + entry.Debit - entry.Credit
	} else {
		// For OWNER_CASH or KPAY, the physical drawer balance doesn't change
		entry.Balance = currentBalance
	}

	return tx.Create(entry).Error
}

// IsDayClosed checks if a CLOSING entry exists for the given date
func (r *CashbookRepository) IsDayClosed(date time.Time) (bool, error) {
	var count int64
	// Format to YYYY-MM-DD to compare only the date part
	dateStr := date.Format("2006-01-02")

	err := r.db.Model(&models.Cashbook{}).
		Where("transaction_type = ? AND DATE(transaction_date) = ?", "CLOSING", dateStr).
		Count(&count).Error

	return count > 0, err
}

func (r *CashbookRepository) GetSettlementReport(month int, year int) ([]SettlementReport, error) {
	var reports []SettlementReport

	query := `
		SELECT 
			DATE(c1.transaction_date) as date,
			(SELECT balance FROM cashbooks WHERE type = 'OPENING' AND DATE(transaction_date) = DATE(c1.transaction_date) LIMIT 1) as opening_balance,
			c1.balance as closing_balance,
			(c1.balance - 5000) as cash_withdrawn
		FROM cashbooks c1
		WHERE c1.type = 'CLOSING' 
		AND EXTRACT(MONTH FROM c1.transaction_date) = ? 
		AND EXTRACT(YEAR FROM c1.transaction_date) = ?
		ORDER BY c1.transaction_date DESC
	`

	err := r.db.Raw(query, month, year).Scan(&reports).Error
	return reports, err
}

func (r *CashbookRepository) GetDashboardSummary() (*DashboardSummary, error) {
	todayStr := time.Now().Format("2006-01-02")
	var summary DashboardSummary

	// 1. Get Opening Balance
	var openingBalance int64
	r.db.Model(&models.DailySummaries{}).
		Select("COALESCE(closing_balance, 0)").
		Where("summary_date < ?", todayStr).
		Order("summary_date DESC").
		Limit(1).
		Scan(&openingBalance)

	if openingBalance == 0 {
		openingBalance = 5000
	}
	summary.OpeningBalance = openingBalance

	// 2. Sales Metrics (FIXED LOGIC)
	// We use grand_total and check against NORMAL/RETURNED status
	r.db.Model(&models.Sale{}).
		Select(`
            COALESCE(SUM(grand_total), 0),
            COALESCE(SUM(CASE WHEN payment_method = 'DEBT' THEN grand_total ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN payment_method = 'CASH' THEN grand_total ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN payment_method = 'KPAY' THEN grand_total ELSE 0 END), 0)
        `).
		Where("DATE(sale_date) = ? AND status IN (?, ?)", todayStr, "NORMAL", "RETURNED").
		Row().
		Scan(
			&summary.TotalSale,
			&summary.TotalNewDebt,
			&summary.TotalCashSales,
			&summary.TotalKPaySales,
		)

	// 3. Today's Returns (Unchanged, but vital)
	var totalReturn int64
	r.db.Model(&models.SaleReturn{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("DATE(return_date) = ?", todayStr).
		Scan(&totalReturn)

	// Adjusting metrics: Total Sale is now "Net"
	// summary.TotalSale -= totalReturn
	// summary.TotalCashSales -= totalReturn

	// 4. Cashbook Movements
	r.db.Model(&models.Cashbook{}).
		Select(`
            COALESCE(SUM(CASE WHEN transaction_type = 'DEBT_PAYMENT' THEN debit ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN transaction_type = 'PURCHASE' THEN credit ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN transaction_type = 'PURCHASE_PAYMENT' THEN credit ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN transaction_type = 'OWNER_WITHDRAWAL' THEN credit ELSE 0 END), 0)
        `).
		Where("DATE(transaction_date) = ?", todayStr).
		Row().
		Scan(
			&summary.TotalDebtCollected,
			&summary.TotalPurchase,
			&summary.TotalSupplierPaid,
			&summary.TotalWithdrawals,
		)

	// 5. Drawer Balance (Physical)
	var latestBalance int64
	if err := r.db.Model(&models.Cashbook{}).Select("balance").Order("id desc").Limit(1).Scan(&latestBalance).Error; err != nil {
		summary.CurrentDrawerBalance = openingBalance
	} else {
		summary.CurrentDrawerBalance = latestBalance
	}

	// 6. Global Debt Totals (FIXED: balance_amount and payment_status)
	// Payables
	r.db.Model(&models.Purchase{}).
		Select("COALESCE(SUM(grand_total - paid_amount), 0)").
		Where("payment_status IN ?", []string{"PENDING", "PARTIAL"}).
		Scan(&summary.TotalPayables)

	// Receivables (This fixes the 15,000 issue)
	r.db.Model(&models.Sale{}).
		Select("COALESCE(SUM(grand_total - paid_amount), 0)").
		Where("payment_status IN ?", []string{"UNPAID", "PARTIAL"}).
		Scan(&summary.TotalReceivables)

	// 7. Closing Balance Projection
	var drawerCashOutflows int64
	r.db.Model(&models.Cashbook{}).
		Select("COALESCE(SUM(credit), 0)").
		Where("DATE(transaction_date) = ? AND payment_method = ? AND transaction_type NOT IN (?, ?)",
			todayStr, "CASH", "SALE_RETURN", "OWNER_WITHDRAWAL").
		Scan(&drawerCashOutflows)

	summary.ClosingBalance = (summary.OpeningBalance + summary.TotalCashSales + summary.TotalDebtCollected) - drawerCashOutflows
	summary.TotalCashInflow = summary.TotalCashSales + summary.TotalDebtCollected

	return &summary, nil
}

func (r *CashbookRepository) GetPastSummaries() ([]models.DailySummaries, error) {
	var summaries []models.DailySummaries

	// Order by summary_date DESC to show the most recent days at the top
	// We limit to 5 or 10 so the response stays fast
	err := r.db.Order("summary_date DESC").
		Limit(10).
		Find(&summaries).Error

	if err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *CashbookRepository) GetCurrentDrawerBalance() (int64, error) {
	var lastEntry models.Cashbook
	err := r.db.Order("id desc").First(&lastEntry).Error
	if err != nil {
		return 5000, nil // Return opening if no transactions exist
	}
	return lastEntry.Balance, nil
}

func (r *CashbookRepository) ReconcileToday() ([]map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	var results []map[string]interface{}

	query := `
        WITH sale_totals AS (
            SELECT invoice_no, payment_method, grand_total, sale_date::DATE as d 
            FROM sales WHERE sale_date::DATE = ?
        ),
        cashbook_totals AS (
            SELECT reference_id, debit FROM cashbooks 
            WHERE transaction_type = 'SALE' AND transaction_date::DATE = ?
        )
        SELECT 
            s.invoice_no as reference,
            s.grand_total as sale_amount,
            COALESCE(c.debit, 0) as ledger_amount,
            (s.grand_total - COALESCE(c.debit, 0)) as discrepancy
        FROM sale_totals s
        LEFT JOIN cashbook_totals c ON s.invoice_no = c.reference_id
        WHERE s.grand_total != COALESCE(c.debit, 0)`

	err := r.db.Raw(query, today, today).Scan(&results).Error
	return results, err
}

func (r *CashbookRepository) RecordOwnerWithdrawal(amount int64, description string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Prepare the Cashbook Entry
		entry := &models.Cashbook{
			TransactionType: "OWNER_WITHDRAWAL",
			Description:     description,
			Debit:           0,
			Credit:          amount,
			PaymentMethod:   "CASH",
		}

		// 2. Use your existing CreateEntry logic to save and update running balance
		if err := r.CreateEntry(tx, entry); err != nil {
			return err
		}

		// 3. Update Daily Summary (Critical for Dashboard accuracy)
		today := time.Now().Format("2006-01-02")
		return tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ? AND is_closed = ?", today, false).
			Update("closing_balance", gorm.Expr("closing_balance - ?", amount)).Error
	})
}

func (r *CashbookRepository) GetEntriesByDateAndType(date time.Time, transType string) ([]models.Cashbook, error) {
	var entries []models.Cashbook
	// Use BETWEEN to cover the full day from 00:00:00 to 23:59:59
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Where("transaction_date >= ? AND transaction_date < ? AND transaction_type = ?",
		startOfDay, endOfDay, transType).
		Order("transaction_date desc").
		Find(&entries).Error

	return entries, err
}
