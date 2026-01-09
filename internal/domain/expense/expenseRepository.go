package expense

import (
	"errors"
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

func (r *ExpenseRepository) CreateOld(input *models.Expense) (*models.Expense, error) {
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
	// 2. CASHBOOK LOGIC (Cash Out)
	// ==========================================
	var lastEntry models.Cashbook
	var currentBalance int64 = 0

	// Safety Check: Get the most recent balance.
	// If no records exist, currentBalance stays 0.
	result := tx.Order("id desc").Limit(1).Find(&lastEntry)
	if result.Error == nil && result.RowsAffected > 0 {
		currentBalance = lastEntry.Balance
	}

	cashOutEntry := models.Cashbook{
		TransactionDate: input.ExpenseDate,
		TransactionType: "EXPENSE",
		// Using fmt.Sprint to safely handle uint ID to string
		ReferenceID: fmt.Sprintf("%d", input.ID),
		Description: fmt.Sprintf("[%s] %s", input.Category, input.Description),
		Debit:       0,
		Credit:      int64(input.Amount), // Ensure matching types
		Balance:     currentBalance - int64(input.Amount),
		CreatedAt:   time.Now(),
	}

	if err := tx.Create(&cashOutEntry).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to record expense in cashbook: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return input, nil
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
	// 2. CASHBOOK LOGIC
	// ==========================================
	var lastEntry models.Cashbook
	var currentBalance int64 = 0

	err := tx.Order("id desc").First(&lastEntry).Error
	if err == nil {
		currentBalance = lastEntry.Balance
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	expenseAmount := int64(input.Amount)

	// AUTO-INJECTION logic
	if currentBalance < expenseAmount {
		injectionAmount := expenseAmount - currentBalance
		injection := models.Cashbook{
			TransactionDate: input.ExpenseDate,
			TransactionType: "OWNER_INJECTION",
			ReferenceID:     fmt.Sprint(input.ID),
			Description:     fmt.Sprintf("Auto-injection to cover %s expense", input.Category),
			Debit:           injectionAmount,
			Credit:          0,
			Balance:         currentBalance + injectionAmount,
			CreatedAt:       time.Now(),
		}

		if err := tx.Create(&injection).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed auto-injection: %v", err)
		}
		currentBalance = injection.Balance
	}

	// Record Expense Cash Out
	newBalance := currentBalance - expenseAmount
	cashOutEntry := models.Cashbook{
		TransactionDate: input.ExpenseDate,
		TransactionType: "EXPENSE",
		ReferenceID:     fmt.Sprint(input.ID),
		Description:     fmt.Sprintf("[%s] %s", input.Category, input.Description),
		Debit:           0,
		Credit:          expenseAmount,
		Balance:         newBalance,
		CreatedAt:       time.Now(),
	}

	if err := tx.Create(&cashOutEntry).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to record expense in cashbook: %v", err)
	}

	// ==========================================
	// 3. DAILY SUMMARY SYNC
	// ==========================================
	todayStr := input.ExpenseDate.Format("2006-01-02")

	// Attempt to update the closing balance for today
	result := tx.Model(&models.DailySummaries{}).
		Where("DATE(summary_date) = ?", todayStr).
		Update("closing_balance", newBalance)

	if result.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update daily summary: %v", result.Error)
	}

	// If no summary exists yet for today, create one
	if result.RowsAffected == 0 {
		newSummary := models.DailySummaries{
			SummaryDate:    input.ExpenseDate,
			OpeningBalance: (newBalance + expenseAmount), // The balance before the expense (but after injection)
			ClosingBalance: newBalance,
			IsClosed:       false,
		}
		if err := tx.Create(&newSummary).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
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
