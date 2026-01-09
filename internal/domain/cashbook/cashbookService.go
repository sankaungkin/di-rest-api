package cashbook

import (
	"log"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CashbookServiceInterface interface {
	GetBalance() (int64, error)
	GetLedger(startDate, endDate string) ([]models.Cashbook, error)
	GetAll(startDate, endDate string) ([]models.Cashbook, error)
	CloseDay(today time.Time) error
	CreateEntry(tx *gorm.DB, entry *models.Cashbook) error
	GetSettlementReport(month int, year int) ([]SettlementReport, error)
	GetTransactionHistory(startDate, endDate string) (map[string]interface{}, error)
}

type CashbookService struct {
	repo CashbookRepositoryInterface
}

var (
	svcInstance *CashbookService
	svcOnce     sync.Once
)

func NewCashbookService(repo CashbookRepositoryInterface) CashbookServiceInterface {
	log.Println(util.Cyan + "CashbookService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &CashbookService{repo: repo}
	})
	return svcInstance
}

func (s *CashbookService) GetBalance() (int64, error) {
	return s.repo.GetBalance()
}

func (s *CashbookService) GetLedger(startDate, endDate string) ([]models.Cashbook, error) {
	return s.repo.GetLedger(startDate, endDate)
}

func (s *CashbookService) CloseDay(today time.Time) error {
	return s.repo.CloseDay(today)
}

func (s *CashbookService) CreateEntry(tx *gorm.DB, entry *models.Cashbook) error {
	return s.repo.CreateEntry(tx, entry)
}

func (s *CashbookService) GetSettlementReport(month int, year int) ([]SettlementReport, error) {
	return s.repo.GetSettlementReport(month, year)
}

func (s *CashbookService) GetAll(startDate, endDate string) ([]models.Cashbook, error) {
	return s.repo.GetAll(startDate, endDate)
}

func (s *CashbookService) GetTransactionHistory(startDate, endDate string) (map[string]interface{}, error) {
	data, err := s.repo.GetAll(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalIn, totalOut int64
	for _, item := range data {
		totalIn += item.Debit
		totalOut += item.Credit
	}

	return map[string]interface{}{
		"transactions": data,
		"total_in":     totalIn,
		"total_out":    totalOut,
		"count":        len(data),
	}, nil
}
