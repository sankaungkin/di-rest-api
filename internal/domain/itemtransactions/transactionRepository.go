package itemtransactions

import (
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type TransactionRepositoryInterface interface {
	GetAll() ([]models.ItemTransaction, error)
	GetByProductId(id string) ([]models.ItemTransaction, error)
	GetByTransactionType(tranType string) ([]models.ItemTransaction, error)
	GetByProductIdAndTranType(productId string, tran_type string) ([]models.ItemTransaction, error)
	CreateAdjustmentTransaction(transaction models.ItemTransaction) (*models.ItemTransaction, error)
}

type TransactionRepository struct {
	db *gorm.DB
}

var (
	repoInstance *TransactionRepository
	repoOnce     sync.Once
)

func NewTransactionRepository(db *gorm.DB) TransactionRepositoryInterface {
	log.Println(util.Green + "TransactionRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &TransactionRepository{db: db}
	})
	return repoInstance
}

func (r *TransactionRepository) GetAll() ([]models.ItemTransaction, error) {
	transactions := []models.ItemTransaction{}
	r.db.Model(&models.ItemTransaction{}).Order("ID asc").Find(&transactions)
	if len(transactions) == 0 {
		return nil, errors.New("NO records found")
	}
	return transactions, nil
}

func (r *TransactionRepository) GetByProductId(productId string) ([]models.ItemTransaction, error) {
	var transactions []models.ItemTransaction

	result := r.db.Where("product_id = ?",
		strings.ToUpper(productId)).
		Order("created_at DESC").
		Find(&transactions)
	if err := result.Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) GetByTransactionType(tran_type string) ([]models.ItemTransaction, error) {
	var transactions []models.ItemTransaction

	result := r.db.Where("tran_type = ?", strings.ToUpper(tran_type)).Find(&transactions)
	if err := result.Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *TransactionRepository) GetByProductIdAndTranType(productId string, tran_type string) ([]models.ItemTransaction, error) {

	var transactions []models.ItemTransaction
	result := r.db.Where("product_id = ? AND tran_type = ?", strings.ToUpper(productId), strings.ToUpper(tran_type)).Find(&transactions).Order("created_at DESC")
	if err := result.Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
func (r *TransactionRepository) CreateAdjustmentTransaction(transaction models.ItemTransaction) (*models.ItemTransaction, error) {
	// Make a copy to return after transaction
	var createdTransaction models.ItemTransaction

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Step 1: Create ItemTransaction
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		// Keep the created transaction (with auto-generated ID, etc.)
		createdTransaction = transaction

		// Step 2: Fetch ProductStock
		var stock models.ProductStock
		if err := tx.Where("product_id = ?", transaction.ProductId).First(&stock).Error; err != nil {
			return err
		}

		// Step 3: Set new quantity based on UOM
		if transaction.Uom == "EACH" {
			stock.BaseQty = transaction.InQty
		} else {
			stock.DerivedQty = transaction.InQty
		}

		// Step 4: Save updated stock
		if err := tx.Save(&stock).Error; err != nil {
			return err
		}

		return nil // commit transaction
	})

	if err != nil {
		return nil, err
	}
	return &createdTransaction, nil
}
