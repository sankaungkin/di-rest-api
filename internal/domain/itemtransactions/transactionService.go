package itemtransactions

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type TransactionServiceInterface interface {
	GetAll() ([]models.ItemTransaction, error)
	GetByProductId(id string) ([]models.ItemTransaction, error)
	GetByTransactionType(tranType string) ([]models.ItemTransaction, error)
	GetByProductIdAndTranType(productId string, tran_type string) ([]models.ItemTransaction, error)
}

type TransactionService struct {
	repo TransactionRepositoryInterface
}

var (
	svcInstnace *TransactionService
	svcOnce     sync.Once
)

func NewTransactionService(repo TransactionRepositoryInterface) TransactionServiceInterface {
	log.Println(util.Green + "TransactionService constructor is called" + util.Reset)
	svcOnce.Do(func() {
		svcInstnace = &TransactionService{repo: repo}
	})
	return svcInstnace
}

func (s *TransactionService) GetAll() ([]models.ItemTransaction, error) {
	return s.repo.GetAll()
}

func (s *TransactionService) GetByProductId(id string) ([]models.ItemTransaction, error) {
	return s.repo.GetByProductId(id)
}

func (s *TransactionService) GetByTransactionType(tranType string) ([]models.ItemTransaction, error) {
	return s.repo.GetByTransactionType(tranType)
}

func (s *TransactionService) GetByProductIdAndTranType(productId string, tran_type string) ([]models.ItemTransaction, error) {
	return s.repo.GetByProductIdAndTranType(productId, tran_type)
}
