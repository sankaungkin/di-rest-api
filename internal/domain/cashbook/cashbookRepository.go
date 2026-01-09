package cashbook

import (
	"errors"
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
	GetAll(startDate, endDate string) ([]models.Cashbook, error)
	GetLedger(startDate, endDate string) ([]models.Cashbook, error)
	CloseDay(today time.Time) error
	CreateEntry(tx *gorm.DB, entry *models.Cashbook) error
	GetSettlementReport(month int, year int) ([]SettlementReport, error)
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
func (r *CashbookRepository) GetBalanceold() (int64, error) {
	var openingEntry models.Cashbook
	todayStr := time.Now().Format("2006-01-02")

	// Find the opening record for today
	err := r.db.Where("transaction_type = ? AND DATE(created_at) = ?", "OPENING", todayStr).First(&openingEntry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no opening record exists, we assume a default start
			return 5000, nil
		}
		return 0, err
	}

	// Sum all transactions occurring AFTER the opening entry today
	var totalMovement int64
	r.db.Model(&models.Cashbook{}).
		Where("id > ? AND DATE(created_at) = ?", openingEntry.ID, todayStr).
		Select("SUM(amount)").
		Scan(&totalMovement)

	return openingEntry.Balance + totalMovement, nil
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
		// 1. Get the final balance before closing
		currentBalance, err := r.GetBalance() // Returns last row balance
		if err != nil {
			return err
		}

		// 2. Record the CLOSING Entry in Cashbook
		closing := models.Cashbook{
			TransactionDate: today,
			TransactionType: "CLOSING",
			Description:     "Daily Cashbook Closed",
			Debit:           0,
			Credit:          0,
			Balance:         currentBalance,
		}
		if err := tx.Create(&closing).Error; err != nil {
			return err
		}

		// 3. UPDATE DailySummary for Today
		// This marks today as officially closed in your summary table
		todayStr := today.Format("2006-01-02")
		if err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ?", todayStr).
			Updates(map[string]interface{}{
				"closing_balance": currentBalance,
				"is_closed":       true,
			}).Error; err != nil {
			return err
		}

		// 4. Record the OPENING Entry for Tomorrow (The 5000 reset)
		tomorrow := today.AddDate(0, 0, 1)
		opening := models.Cashbook{
			TransactionDate: tomorrow,
			TransactionType: "OPENING",
			Description:     "Daily Opening Balance (Reset)",
			Debit:           5000,
			Credit:          0,
			Balance:         5000,
		}
		if err := tx.Create(&opening).Error; err != nil {
			return err
		}

		// 5. CREATE DailySummary for Tomorrow
		// This prepares the summary record for the next morning
		newSummary := models.DailySummaries{
			SummaryDate:    tomorrow,
			OpeningBalance: 5000,
			ClosingBalance: 0,
			IsClosed:       false,
		}
		if err := tx.Create(&newSummary).Error; err != nil {
			return err
		}

		return nil
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
