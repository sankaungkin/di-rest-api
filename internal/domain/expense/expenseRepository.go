package expense

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/cashbook"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExpenseRepositoryInterface interface {
	Create(expense *models.Expense) (*models.Expense, error)
	Getall() ([]models.Expense, error)
}

type ExpenseRepository struct {
	db       *gorm.DB
	cashRepo cashbook.CashbookRepositoryInterface
}

var (
	repoInstance *ExpenseRepository
	repoOnce     sync.Once
)

func NewExpenseRepository(db *gorm.DB, cashRepo cashbook.CashbookRepositoryInterface) ExpenseRepositoryInterface {
	log.Println(util.Cyan + "ExpenseRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &ExpenseRepository{db: db, cashRepo: cashRepo}
	})
	return repoInstance
}

func (r *ExpenseRepository) Create(input *models.Expense) (*models.Expense, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Save the Expense record first
		if err := tx.Create(input).Error; err != nil {
			return err
		}

		expenseAmount := int64(input.Amount)
		todayStr := input.ExpenseDate.Format("2006-01-02")

		// 2. Fetch the Current Balance with a LOCK to check for Auto-Injection
		// We use the specialized repo to ensure we get the absolute latest state
		var lastEntry models.Cashbook
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Order("id desc").First(&lastEntry).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		// ==========================================
		// 3. AUTO-INJECTION LOGIC
		// ==========================================
		if lastEntry.Balance < expenseAmount {
			injectionAmt := expenseAmount - lastEntry.Balance
			injection := &models.Cashbook{
				TransactionDate: input.ExpenseDate,
				TransactionType: "OWNER_INJECTION",
				PaymentMethod:   "CASH",
				Description:     fmt.Sprintf("Auto-Injection to cover expense: %s", input.Category),
				Debit:           injectionAmt,
			}
			// Use CreateEntry to update balance and summary
			if err := r.cashRepo.CreateEntry(tx, injection); err != nil {
				return fmt.Errorf("auto-injection failed: %v", err)
			}
		}

		// ==========================================
		// 4. RECORD EXPENSE IN CASHBOOK
		// ==========================================
		cashbookEntry := &models.Cashbook{
			TransactionDate: input.ExpenseDate,
			TransactionType: "EXPENSE",
			PaymentMethod:   "CASH",
			ReferenceID:     fmt.Sprintf("EXP-%d", input.ID),
			Description:     fmt.Sprintf("[%s] %s", input.Category, input.Description),
			Credit:          expenseAmount,
		}

		// This handles the math (Balance - expenseAmount) and updates the summary
		if err := r.cashRepo.CreateEntry(tx, cashbookEntry); err != nil {
			return fmt.Errorf("cashbook expense entry failed: %v", err)
		}

		// ==========================================
		// 5. UPDATE DAILY SUMMARY (Non-balance metrics)
		// ==========================================
		// CreateEntry handled 'closing_balance'. We just increment 'expense_total'.
		if err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ?", todayStr).
			Update("expense_total", gorm.Expr("expense_total + ?", expenseAmount)).Error; err != nil {
			return fmt.Errorf("failed to update daily expense total: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return input, nil
}

func (r *ExpenseRepository) Getall() ([]models.Expense, error) {
	var expenses []models.Expense
	result := r.db.Order("created_at desc").Find(&expenses)
	if result.Error != nil {
		return nil, result.Error
	}
	return expenses, nil
}
