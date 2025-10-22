package purchase

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type PurchaseServiceInterface interface {
	CreateService(purchase *models.Purchase) (*models.Purchase, error)
	GetAllService() ([]models.Purchase, error)
	GetTodayPurchases() ([]models.Purchase, error)
	GetById(id string) (*models.Purchase, error)
	GetTodayGrandTotal() (int64, error)
	GetMonthlyPurchases() ([]models.Purchase, error)
	GetMonthlyGrandTotal() (int64, error)
	UpdatePurchaseRemark(purchaseRemark UpdateRemarkPurchaseDTO) (*models.Purchase, error)
	GetPurchaseLineItems() ([]ResponsePurchaseLineItemDTO, error)
}

type PurchaseService struct {
	repo PurchaseRepositoryInterface
}

var (
	svcInstance *PurchaseService
	svcOnce     sync.Once
)

func NewSaleService(repo PurchaseRepositoryInterface) PurchaseServiceInterface {
	log.Println(util.Magenta + "SaleService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &PurchaseService{repo: repo}
	})
	return svcInstance
}

func (s *PurchaseService) CreateService(purchase *models.Purchase) (*models.Purchase, error) {
	return s.repo.Create(purchase)
}

func (s *PurchaseService) GetAllService() ([]models.Purchase, error) {
	return s.repo.GetAll()
}

func (s *PurchaseService) GetTodayPurchases() ([]models.Purchase, error) {
	return s.repo.GetTodayPurchases()
}

func (s *PurchaseService) GetById(id string) (*models.Purchase, error) {
	return s.repo.GetById(id)
}

func (s *PurchaseService) GetTodayGrandTotal() (int64, error) {
	return s.repo.GetTodayGrandTotal()
}

func (s *PurchaseService) GetMonthlyPurchases() ([]models.Purchase, error) {
	return s.repo.GetMonthlyPurchases()
}

func (s *PurchaseService) GetMonthlyGrandTotal() (int64, error) {
	return s.repo.GetMonthlyGrandTotal()
}

func (s *PurchaseService) GetPurchaseLineItems() ([]ResponsePurchaseLineItemDTO, error) {
	return s.repo.GetPurchaseLineItems()
}

func (s *PurchaseService) UpdatePurchaseRemark(purchaseRemark UpdateRemarkPurchaseDTO) (*models.Purchase, error) {
	return s.repo.UpdatePurchaseRemark(purchaseRemark)
}
