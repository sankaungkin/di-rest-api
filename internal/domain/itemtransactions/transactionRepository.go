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

	result := r.db.Where("product_id = ?", strings.ToUpper(productId)).Find(&transactions)
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
	result := r.db.Where("product_id = ? AND tran_type = ?", strings.ToUpper(productId), strings.ToUpper(tran_type)).Find(&transactions)
	if err := result.Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
