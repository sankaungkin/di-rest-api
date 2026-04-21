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
	// CreateEntry(tx *gorm.DB, entry *models.Cashbook) error
	CreateEntry(entry *models.Cashbook) error
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
func (r *CashbookRepository) GetLedgerOLD(startDate, endDate string) ([]models.Cashbook, error) {
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

func (r *CashbookRepository) GetLedger(startDate, endDate string) ([]models.Cashbook, error) {
	var entries []models.Cashbook

	// Postgres handles the "if empty" logic now
	err := r.db.Raw("SELECT * FROM get_cashbook_ledger(?, ?)", startDate, endDate).Scan(&entries).Error

	return entries, err
}

func (r *CashbookRepository) CloseDayOld(today time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		loc, _ := time.LoadLocation("Asia/Yangon")
		// Use the passed 'today' but ensure we are looking at the YMD part correctly
		todayStr := today.In(loc).Format("2006-01-02")
		tomorrow := today.In(loc).AddDate(0, 0, 1)
		// tomorrowStr := tomorrow.Format("2006-01-02")

		// 1. Get the LATEST balance snapshot in the ledger (FOR UPDATE)
		var lastEntry models.Cashbook
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Order("transaction_date DESC, id DESC"). // Better sorting for consistency
			First(&lastEntry).Error; err != nil {
			return fmt.Errorf("ledger empty: %v", err)
		}

		// 2. Calculate the "Take Home"
		currentBalance := lastEntry.Balance
		amountToWithdraw := currentBalance - 5000

		// 3. Record the Withdrawal ONLY if there is surplus cash
		if amountToWithdraw > 0 {
			sweep := models.Cashbook{
				// Crucial: Use the END of the day for the timestamp to keep it as the last entry
				TransactionDate:   today.In(loc),
				TransactionType:   "OWNER_WITHDRAWAL",
				PaymentMethod:     "CASH",
				ReferenceID:       "EOD-" + todayStr,
				Description:       "End of Day Reset (Float: 5000)",
				Debit:             0,
				Credit:            amountToWithdraw,
				Balance:           5000,
				TransactionStatus: "COMPLETED",
			}
			if err := tx.Create(&sweep).Error; err != nil {
				return err
			}
		}

		// 4. Update Today's Summary
		if err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ?", todayStr).
			Updates(map[string]interface{}{
				"closing_balance": 5000,
				"is_closed":       true,
			}).Error; err != nil {
			return err
		}

		// 5. Upsert Tomorrow's Summary (Using Upsert to avoid duplicate errors)
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "summary_date"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"opening_balance": 5000}),
		}).Create(&models.DailySummaries{
			SummaryDate:    tomorrow,
			OpeningBalance: 5000,
			ClosingBalance: 5000,
			IsClosed:       false,
		}).Error
	})
}

func (r *CashbookRepository) CloseDay(today time.Time) error {
	// Just pass the time; Postgres handles the Yangon conversion and the logic
	return r.db.Exec("SELECT close_day(?)", today).Error
}

func (r *CashbookRepository) CreateEntryOLD(tx *gorm.DB, entry *models.Cashbook) error {
	return tx.Transaction(func(subTx *gorm.DB) error {
		// 1. Find the balance as of the MOMENT BEFORE this transaction date
		// Use a row-level lock to prevent race conditions during simultaneous sales
		var previousEntry models.Cashbook
		err := subTx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("transaction_date <= ?", entry.TransactionDate).
			Order("transaction_date DESC, id DESC").
			First(&previousEntry).Error

		var startingBalance int64
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				startingBalance = 5000 // Default base float
			} else {
				return err
			}
		} else {
			startingBalance = previousEntry.Balance
		}

		// 2. SAFETY CHECK: Prevent the drawer from going into impossible negatives
		// If it's a Cash Withdrawal or Expense, ensure we actually have the cash.
		if entry.PaymentMethod == "CASH" && entry.Credit > 0 {
			if entry.TransactionType == "OWNER_WITHDRAWAL" || entry.TransactionType == "EXPENSE" {
				if startingBalance < entry.Credit {
					return fmt.Errorf("insufficient cash in drawer (Available: %d, Requested: %d)", startingBalance, entry.Credit)
				}
			}
		}

		// 3. Calculate New Balance (Only for CASH)
		if entry.PaymentMethod == "CASH" {
			entry.Balance = startingBalance + entry.Debit - entry.Credit
		} else {
			// Non-cash transactions (KPAY) don't affect the physical drawer balance
			entry.Balance = startingBalance
		}

		// 4. Save the record
		if err := subTx.Create(entry).Error; err != nil {
			return err
		}

		// 5. THE RIPPLE EFFECT (Crucial for Backdating)
		if entry.PaymentMethod == "CASH" {
			changeAmount := entry.Debit - entry.Credit
			if changeAmount != 0 {
				// Update every subsequent record's balance snapshot
				err := subTx.Model(&models.Cashbook{}).
					Where("transaction_date > ? OR (transaction_date = ? AND id > ?)",
						entry.TransactionDate, entry.TransactionDate, entry.ID).
					Update("balance", gorm.Expr("balance + ?", changeAmount)).Error
				if err != nil {
					return err
				}
			}
		}

		// 6. Update Daily Summaries
		// This ensures the Dashboard JSON reflects the changes immediately
		return r.syncDailySummaries(subTx, entry.TransactionDate)
	})
}

func (r *CashbookRepository) CreateEntry(entry *models.Cashbook) error {
	return r.db.Exec("SELECT create_cashbook_entry(?, ?, ?, ?, ?, ?)",
		entry.TransactionDate,
		entry.TransactionType,
		entry.PaymentMethod,
		entry.Debit,
		entry.Credit,
		entry.Description,
	).Error
}

func (r *CashbookRepository) syncDailySummaries(tx *gorm.DB, transactionDate time.Time) error {
	dateStr := transactionDate.Format("2006-01-02")

	// 1. Get totals from CASHBOOK (Money actually received/spent)
	var cb struct {
		CashIn        int64 // cash_total
		KPayIn        int64 // k_pay_total
		Expenses      int64 // expense_total
		Withdrawals   int64 // total_withdrawal
		DebtCollected int64 // debt_collected_total
	}
	tx.Model(&models.Cashbook{}).
		Select(`
            SUM(CASE WHEN transaction_type = 'SALE' AND payment_method = 'CASH' THEN debit ELSE 0 END) as cash_in,
            SUM(CASE WHEN transaction_type = 'SALE' AND payment_method = 'KPAY' THEN debit ELSE 0 END) as k_pay_in,
            SUM(CASE WHEN transaction_type = 'EXPENSE' THEN credit ELSE 0 END) as expenses,
            SUM(CASE WHEN transaction_type = 'OWNER_WITHDRAWAL' THEN credit ELSE 0 END) as withdrawals,
            SUM(CASE WHEN transaction_type = 'DEBT_PAYMENT' THEN debit ELSE 0 END) as debt_collected
        `).
		Where("DATE(transaction_date) = ?", dateStr).
		Scan(&cb)

	// 2. Get totals from SALES (Total revenue and New Debt created)
	var s struct {
		GrandTotal int64 // total_sale
		NewDebt    int64 // debt_total
	}
	tx.Model(&models.Sale{}).
		Select(`
            SUM(grand_total) as grand_total,
            SUM(balance_amount) as new_debt
        `).
		Where("DATE(sale_date) = ?", dateStr).
		Scan(&s)

	// 3. Get Final Balance
	var finalBalance int64
	tx.Model(&models.Cashbook{}).
		Where("DATE(transaction_date) = ?", dateStr).
		Order("transaction_date DESC, id DESC").
		Limit(1).Pluck("balance", &finalBalance)

	// 4. MAPPING TO DATABASE (Ensure keys match your DB column names exactly)
	return tx.Model(&models.DailySummaries{}).
		Where("DATE(summary_date) = ?", dateStr).
		Updates(map[string]interface{}{
			"total_sale":           s.GrandTotal, // All Sales (Paid + Unpaid)
			"cash_total":           cb.CashIn,    // Actual Cash from Sales
			"total_cash_sales":     cb.CashIn,
			"k_pay_total":          cb.KPayIn,
			"debt_total":           s.NewDebt,        // New Debt created today
			"debt_collected_total": cb.DebtCollected, // Debt paid back today
			"expense_total":        cb.Expenses,
			"total_withdrawal":     cb.Withdrawals,
			"closing_balance":      finalBalance,
		}).Error
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

func (r *CashbookRepository) GetDashboardSummaryOLD() (*DashboardSummary, error) {
	loc, _ := time.LoadLocation("Asia/Yangon")
	today := time.Now().In(loc)
	todayStr := today.Format("2006-01-02")
	var summary DashboardSummary

	// 1. OPENING BALANCE (Previous day's last balance)
	// We look for the last entry BEFORE today
	r.db.Model(&models.Cashbook{}).
		Select("COALESCE(balance, 5000)").
		Where("DATE(transaction_date AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Yangon') < ?", todayStr).
		Order("transaction_date DESC, id DESC").
		Limit(1).Scan(&summary.OpeningBalance)

	// 2. SALES METRICS (Excluding Debt Payments to prevent double-counting)
	// We only want money that came in via the SALE itself, not later collection.
	r.db.Model(&models.Sale{}).
		Select(`
            COALESCE(SUM(grand_total), 0),
            COALESCE(SUM(balance_amount), 0),
            COALESCE(SUM(CASE WHEN payment_method = 'CASH' THEN (grand_total - balance_amount) ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN payment_method = 'KPAY' THEN (grand_total - balance_amount) ELSE 0 END), 0)
        `).
		Where("DATE(sale_date AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Yangon') = ?", todayStr).
		Where("status IN ('NORMAL', 'PARTIAL_RETURN', 'RETURNED')").
		Row().Scan(
		&summary.TotalSale,
		&summary.TotalNewDebt,
		&summary.TotalCashSales, // This is now 'Cash collected AT TIME OF SALE'
		&summary.TotalKPaySales,
	)

	// 3. CASHBOOK MOVEMENTS & CURRENT BALANCE
	// We fetch the very last balance recorded in the ledger for today.
	var cashRefunds int64
	r.db.Model(&models.Cashbook{}).
		Select(`
            COALESCE(SUM(CASE WHEN transaction_type = 'DEBT_PAYMENT' AND payment_method = 'CASH' THEN debit ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN transaction_type = 'DEBT_PAYMENT' AND payment_method = 'KPAY' THEN debit ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN transaction_type = 'OWNER_WITHDRAWAL' THEN credit ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN transaction_type = 'EXPENSE' THEN credit ELSE 0 END), 0),
            COALESCE(SUM(CASE WHEN transaction_type = 'SALE_RETURN' THEN credit ELSE 0 END), 0),
            (SELECT balance FROM cashbooks 
             WHERE DATE(transaction_date AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Yangon') <= ? 
             ORDER BY transaction_date DESC, id DESC LIMIT 1)
        `, todayStr).
		Where("DATE(transaction_date AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Yangon') = ?", todayStr).
		Row().Scan(
		&summary.TotalDebtCollected,
		&summary.TotalKPayCollected,
		&summary.TotalWithdrawals,
		&summary.TotalExpenses,
		&cashRefunds,
		&summary.CurrentDrawerBalance, // TRUST THE LEDGER
	)

	// // 4. PURCHASES (Daily)
	// r.db.Model(&models.Purchase{}).
	// 	Select("COALESCE(SUM(grand_total), 0)").
	// 	Where("DATE(purchase_date AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Yangon') = ?", todayStr).
	// 	Scan(&summary.TotalPurchase)

	// // 5. FINAL TOTALS & RECEIVABLES
	// summary.TotalCashInflow = summary.TotalCashSales + summary.TotalDebtCollected

	// Replace your Section 4 logic with this:
	r.db.Model(&models.Cashbook{}).
		Select(`
        COALESCE(SUM(CASE WHEN transaction_type = 'SALE' AND payment_method = 'CASH' THEN debit ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN transaction_type = 'DEBT_PAYMENT' AND payment_method = 'CASH' THEN debit ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN transaction_type = 'OWNER_WITHDRAWAL' THEN credit ELSE 0 END), 0)
    `).
		Where("DATE(transaction_date AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Yangon') = ?", todayStr).
		Row().Scan(&summary.TotalCashSales, &summary.TotalDebtCollected, &summary.TotalWithdrawals)

	// Total Cash Inflow should now be a clean sum of these two distinct Cashbook categories
	summary.TotalCashInflow = summary.TotalCashSales + summary.TotalDebtCollected
	summary.ClosingBalance = summary.CurrentDrawerBalance

	r.db.Model(&models.Sale{}).
		Select("COALESCE(SUM(balance_amount), 0)").
		Where("payment_status NOT IN ?", []string{"PAID", "FULLY PAID"}).
		Scan(&summary.TotalReceivables)

	r.db.Model(&models.Purchase{}).
		Select("COALESCE(SUM(balance_amount), 0)").
		Where("payment_status != ?", "PAID").
		Scan(&summary.TotalPayables)

	return &summary, nil
}

func (r *CashbookRepository) GetDashboardSummary() (*DashboardSummary, error) {
	var summary DashboardSummary
	// GORM can scan the custom type directly into your struct
	// as long as the struct fields match the type property names.
	err := r.db.Raw("SELECT * FROM get_dashboard_summary()").Scan(&summary).Error
	if err != nil {
		return nil, err
	}
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

func (r *CashbookRepository) RecordOwnerWithdrawalOLD(amount int64, description string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Prepare the Entry
		// This will trigger your CreateEntry logic which handles the
		// clause.Locking{Strength: "UPDATE"} to prevent math errors
		entry := &models.Cashbook{
			TransactionDate: time.Now(),
			TransactionType: "OWNER_WITHDRAWAL",
			Description:     description, // e.g., "Owner took for lunch" or "Random withdrawal"
			Debit:           0,
			Credit:          amount,
			PaymentMethod:   "CASH",
			ReferenceID:     "DW-" + time.Now().Format("150405"), // Timestamp based ID
		}

		// 2. CreateEntry handles the "Running Balance" calculation
		if err := r.CreateEntry(entry); err != nil {
			return err
		}

		return nil
	})
}

func (r *CashbookRepository) RecordOwnerWithdrawal(amount int64, description string) error {
	// We no longer need the Go-level transaction or the complex struct building.
	// The DB handles the timestamp and the business rules.
	return r.db.Exec("SELECT record_owner_withdrawal(?, ?)", amount, description).Error
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
