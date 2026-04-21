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
	GetCashBalance() (int64, error)
	GetLedger(startDate, endDate string) ([]models.Cashbook, error)
	GetAll(startDate, endDate string) ([]models.Cashbook, error)
	CloseDay(today time.Time) error
	// CreateEntry(tx *gorm.DB, entry *models.Cashbook) error
	CreateEntry(entry *models.Cashbook) error
	GetSettlementReport(month int, year int) ([]SettlementReport, error)
	GetTransactionHistory(startDate, endDate string) (map[string]interface{}, error)
	GetDashboardSummary() (*DashboardSummary, error)
	GetCurrentDrawerBalance() (int64, error)
	GetPastSummaries() ([]models.DailySummaries, error)
	ReconcileToday() ([]map[string]interface{}, error)
	RecordOwnerWithdrawal(amount int64, description string) error
	GetEntriesByDateAndType(date time.Time, transType string) ([]models.Cashbook, error)
}

type CashbookService struct {
	repo CashbookRepositoryInterface
	db   *gorm.DB
}

var (
	svcInstance *CashbookService
	svcOnce     sync.Once
)

func NewCashbookService(repo CashbookRepositoryInterface, db *gorm.DB) CashbookServiceInterface {
	log.Println(util.Cyan + "CashbookService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &CashbookService{repo: repo, db: db}
	})
	return svcInstance
}

func (s *CashbookService) GetBalance() (int64, error) {
	return s.repo.GetBalance()
}

func (s *CashbookService) GetCashBalance() (int64, error) {
	return s.repo.GetCashBalance()
}

func (s *CashbookService) GetLedger(startDate, endDate string) ([]models.Cashbook, error) {
	return s.repo.GetLedger(startDate, endDate)
}

func (s *CashbookService) CloseDay(today time.Time) error {
	return s.repo.CloseDay(today)
}

//	func (s *CashbookService) CreateEntry(tx *gorm.DB, entry *models.Cashbook) error {
//		return s.repo.CreateEntry(tx, entry)
//	}
func (s *CashbookService) CreateEntry(entry *models.Cashbook) error {
	// Start a transaction here
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.repo.CreateEntry(entry)
	})
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

func (s *CashbookService) GetDashboardSummary() (*DashboardSummary, error) {
	return s.repo.GetDashboardSummary()
}

func (s *CashbookService) GetCurrentDrawerBalance() (int64, error) {
	return s.repo.GetCurrentDrawerBalance()
}

func (s *CashbookService) GetPastSummaries() ([]models.DailySummaries, error) {
	return s.repo.GetPastSummaries()
}

func (s *CashbookService) ReconcileToday() ([]map[string]interface{}, error) {
	return s.repo.ReconcileToday()
}

func (s *CashbookService) RecordOwnerWithdrawal(amount int64, description string) error {
	return s.repo.RecordOwnerWithdrawal(amount, description)
}

func (s *CashbookService) GetEntriesByDateAndType(date time.Time, transType string) ([]models.Cashbook, error) {
	return s.repo.GetEntriesByDateAndType(date, transType)
}
