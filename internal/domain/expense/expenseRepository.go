package expense

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ExpenseRepositoryInterface interface {
	Create(expense *models.Expense) (*models.Expense, error)
	Getall() ([]models.Expense, error)
}

type ExpenseRepository struct {
	db *gorm.DB
}

var (
	repoInstance *ExpenseRepository
	repoOnce     sync.Once
)

func NewExpenseRepository(db *gorm.DB) ExpenseRepositoryInterface {
	log.Println(util.Cyan + "ExpenseRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &ExpenseRepository{db: db}
	})
	return repoInstance
}

func (r *ExpenseRepository) Create(input *models.Expense) (*models.Expense, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Save the Expense record
	if err := tx.Create(input).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ==========================================
	// 2. CASHBOOK LOGIC (Physical Drawer Sync)
	// ==========================================
	var lastEntry models.Cashbook
	tx.Order("id desc").Limit(1).Find(&lastEntry)

	expenseAmount := int64(input.Amount)
	currentBalance := lastEntry.Balance

	// Auto-Injection if drawer doesn't have enough cash
	if currentBalance < expenseAmount {
		injectionAmt := expenseAmount - currentBalance
		injection := models.Cashbook{
			TransactionDate: input.ExpenseDate,
			TransactionType: "OWNER_INJECTION",
			PaymentMethod:   "CASH",
			Description:     fmt.Sprintf("Injection to cover %s", input.Category),
			Debit:           injectionAmt,
			Balance:         currentBalance + injectionAmt,
			CreatedAt:       time.Now(),
		}
		tx.Create(&injection)
		currentBalance = injection.Balance
	}

	// Record the Expense in Cashbook
	newBalance := currentBalance - expenseAmount
	cashbookEntry := models.Cashbook{
		TransactionDate: input.ExpenseDate,
		TransactionType: "EXPENSE",
		PaymentMethod:   "CASH",
		ReferenceID:     fmt.Sprintf("EXP-%d", input.ID),
		Description:     fmt.Sprintf("[%s] %s", input.Category, input.Description),
		Credit:          expenseAmount,
		Balance:         newBalance,
		CreatedAt:       time.Now(),
	}

	if err := tx.Create(&cashbookEntry).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ==========================================
	// 3. DAILY SUMMARY SYNC
	// ==========================================
	todayStr := input.ExpenseDate.Format("2006-01-02")

	// Update existing summary record
	result := tx.Model(&models.DailySummaries{}).
		Where("DATE(summary_date) = ?", todayStr).
		Updates(map[string]interface{}{
			"closing_balance": gorm.Expr("closing_balance - ?", expenseAmount),
			"expense_total":   gorm.Expr("expense_total + ?", expenseAmount),
		})

	// If summary doesn't exist for today yet
	if result.RowsAffected == 0 {
		var lastSum models.DailySummaries
		tx.Order("summary_date desc").Limit(1).Find(&lastSum)

		tx.Create(&models.DailySummaries{
			SummaryDate:    input.ExpenseDate,
			OpeningBalance: lastSum.ClosingBalance,
			ClosingBalance: lastSum.ClosingBalance - expenseAmount,
			ExpenseTotal:   expenseAmount,
		})
	}

	return input, tx.Commit().Error
}

func (r *ExpenseRepository) Getall() ([]models.Expense, error) {
	var expenses []models.Expense
	result := r.db.Order("created_at desc").Find(&expenses)
	if result.Error != nil {
		return nil, result.Error
	}
	return expenses, nil
}
