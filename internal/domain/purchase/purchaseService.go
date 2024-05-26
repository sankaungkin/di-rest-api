package purchase

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)


type PurchaseServiceInterface interface{
	CreateService(purchase *models.Purchase) (*models.Purchase, error)
	GetAllService() ([]models.Purchase, error)
	GetById(id string) (*models.Purchase, error)
}

type PurchaseService struct{
	repo PurchaseRepositoryInterface
}

var (
	svcInstance *PurchaseService
	svcOnce sync.Once
)

func NewSaleService(repo PurchaseRepositoryInterface) PurchaseServiceInterface{
	log.Println(util.Magenta + "SaleService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &PurchaseService{repo: repo}
	})
	return svcInstance
}

func (s *PurchaseService)CreateService(purchase *models.Purchase) (*models.Purchase, error){
	return s.repo.Create(purchase)
}

func (s *PurchaseService)GetAllService() ([]models.Purchase, error){
	return s.repo.GetAll()
}

func (s *PurchaseService)GetById(id string) (*models.Purchase, error){
	return s.repo.GetById(id)
}