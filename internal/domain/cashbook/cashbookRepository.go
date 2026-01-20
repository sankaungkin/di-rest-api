package cashbook

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
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

func (r *CashbookRepository) CloseDayOld(today time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		todayStr := today.Format("2006-01-02")
		now := time.Now()

		// 1. Get the current status of the drawer BEFORE reset
		var lastEntry models.Cashbook
		tx.Order("id desc").Limit(1).Find(&lastEntry)
		actualEndOfDayBalance := lastEntry.Balance

		// 2. Mark Today's Summary as Closed
		// Note: We save the ACTUAL balance here so we know how much was made today.
		if err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ?", todayStr).
			Updates(map[string]interface{}{
				"closing_balance": actualEndOfDayBalance,
				"is_closed":       true,
			}).Error; err != nil {
			return err
		}

		// 3. RESET THE DRAWER (Transfer to Owner)
		// This moves money out of the "Drawer" and prepares for tomorrow.
		amountToAdjust := actualEndOfDayBalance - 5000

		if amountToAdjust != 0 {
			entry := models.Cashbook{
				TransactionDate: now,
				CreatedAt:       now,
				PaymentMethod:   "CASH",
				ReferenceID:     "RESET-" + todayStr,
			}

			if amountToAdjust > 0 {
				// Cashier gives excess profit to owner
				entry.TransactionType = "OWNER_WITHDRAWAL"
				entry.Description = fmt.Sprintf("Daily Reset: Profit withdrawal %d", amountToAdjust)
				entry.Credit = amountToAdjust
				entry.Balance = 5000
			} else {
				// Owner adds money because drawer is below 5000
				entry.TransactionType = "OWNER_INJECTION"
				entry.Description = fmt.Sprintf("Daily Reset: Owner added %d to reach float", -amountToAdjust)
				entry.Debit = -amountToAdjust
				entry.Balance = 5000
			}
			tx.Create(&entry)
		}

		// 4. PREPARE TOMORROW
		tomorrow := today.AddDate(0, 0, 1)
		tomorrowStr := tomorrow.Format("2006-01-02")

		var nextSummary models.DailySummaries
		result := tx.Where("DATE(summary_date) = ?", tomorrowStr).First(&nextSummary)

		if result.RowsAffected == 0 {
			// Create Tomorrow's Opening Record in Cashbook
			tx.Create(&models.Cashbook{
				TransactionDate: tomorrow,
				TransactionType: "OPENING",
				Description:     "Daily Opening Balance (Reset)",
				Debit:           5000,
				Balance:         5000,
				CreatedAt:       now,
				PaymentMethod:   "CASH",
			})

			// Create Tomorrow's Daily Summary row
			tx.Create(&models.DailySummaries{
				SummaryDate:    tomorrow,
				OpeningBalance: 5000,
				ClosingBalance: 5000,
				IsClosed:       false,
			})
		}

		return nil
	})
}
func (r *CashbookRepository) CloseDay(today time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		todayStr := today.Format("2006-01-02")
		now := time.Now()

		// 1. Calculate Z-Report Totals from Sales Table
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

		// 2. Get Expense Total from Cashbook
		var expenseTotal int64
		tx.Raw(`SELECT COALESCE(SUM(credit), 0) FROM cashbooks 
                WHERE DATE(transaction_date) = ? AND transaction_type = 'EXPENSE'`,
			todayStr).Scan(&expenseTotal)

		// 3. Get Final Drawer Balance BEFORE Reset
		var lastEntry models.Cashbook
		if err := tx.Order("id desc").Limit(1).Find(&lastEntry).Error; err != nil {
			return err
		}
		actualEndOfDayBalance := lastEntry.Balance

		// 4. Update Today's Summary (Using Updates to ensure we hit the existing row)
		// We use the ACTUAL cash on hand for cash_total here
		err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ? AND is_closed = ?", todayStr, false).
			Updates(map[string]interface{}{
				"cash_total":      report.CashTotal, // Total Cash Sales
				"k_pay_total":     report.KPayTotal,
				"debt_total":      report.DebtTotal,
				"expense_total":   expenseTotal,
				"total_sale":      report.TotalSale,
				"closing_balance": actualEndOfDayBalance,
				"is_closed":       true,
			}).Error
		if err != nil {
			return err
		}

		// 5. Reset Drawer to 5000 (Owner takes the rest)
		amountToAdjust := actualEndOfDayBalance - 5000
		if amountToAdjust != 0 {
			tx.Create(&models.Cashbook{
				TransactionDate: now,
				TransactionType: "OWNER_WITHDRAWAL",
				PaymentMethod:   "CASH",
				ReferenceID:     "RESET-" + todayStr,
				Description:     fmt.Sprintf("Daily Reset: Withdrawal %d", amountToAdjust),
				Credit:          amountToAdjust,
				Balance:         5000,
				CreatedAt:       now,
			})
		}

		// 6. Setup Tomorrow (FirstOrCreate prevents duplicate ID 13/15 issues)
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
	var lastEntry models.Cashbook

	// 1. Get the absolute latest balance
	// Use 'tx' to ensure this happens inside the parent transaction
	err := tx.Order("id desc").First(&lastEntry).Error

	currentBalance := int64(5000) // Default if no records exist
	if err == nil {
		currentBalance = lastEntry.Balance
	}

	// 2. Calculate new balance
	// Debit increases balance, Credit decreases it
	entry.Balance = currentBalance + entry.Debit - entry.Credit
	entry.TransactionDate = time.Now()

	// 3. Save the record
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

	// Your business rule for daily start
	summary.OpeningBalance = 5000

	err := r.db.Model(&models.Cashbook{}).
		Select(`
            -- Revenue Metrics
            COALESCE(SUM(CASE WHEN transaction_type = 'SALE' THEN debit ELSE 0 END), 0) as total_sale,
            COALESCE(SUM(CASE WHEN transaction_type = 'SALE' AND payment_method = 'DEBT' THEN debit ELSE 0 END), 0) as total_new_debt,
            COALESCE(SUM(CASE WHEN transaction_type = 'SALE' AND payment_method = 'CASH' THEN debit ELSE 0 END), 0) as total_cash_sales,
            COALESCE(SUM(CASE WHEN transaction_type = 'SALE' AND payment_method = 'KPAY' THEN debit ELSE 0 END), 0) as total_kpay_sales,
            
            -- Receivable Recovery
            COALESCE(SUM(CASE WHEN transaction_type = 'DEBT_PAYMENT' THEN debit ELSE 0 END), 0) as total_debt_collected,

            -- Outflow Metrics (Purchases)
            COALESCE(SUM(CASE WHEN transaction_type = 'PURCHASE' THEN credit ELSE 0 END), 0) as total_purchase
        `).
		Where("DATE(transaction_date) = ?", todayStr).
		Row().
		Scan(
			&summary.TotalSale,
			&summary.TotalNewDebt,
			&summary.TotalCashSales,
			&summary.TotalKPaySales,
			&summary.TotalDebtCollected,
			&summary.TotalPurchase,
		)

	if err != nil {
		return nil, err
	}

	// Cash in Hand Calculation:
	// (Start + Cash Sales + KPay + Collected) - (Stock Purchases)
	summary.ClosingBalance = (summary.OpeningBalance +
		summary.TotalCashSales +
		summary.TotalKPaySales +
		summary.TotalDebtCollected) - summary.TotalPurchase

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
