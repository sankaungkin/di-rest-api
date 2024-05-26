package sale

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type SaleServiceInterface interface{
	CreateService(sale *models.Sale) (*models.Sale, error)
	GetAllService() ([]models.Sale, error)
	GetById(id string) (*models.Sale, error)
}

type SaleService struct{
	repo SaleRepositoryInterface
}

var (
	svcInstance *SaleService
	svcOnce sync.Once
)

func NewSaleService(repo SaleRepositoryInterface) SaleServiceInterface{
	log.Println(util.Blue + "SaleService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &SaleService{repo: repo}
	})
	return svcInstance
}

func (s *SaleService)CreateService(sale *models.Sale) (*models.Sale, error){
	return s.repo.Create(sale)
}

func (s *SaleService)GetAllService() ([]models.Sale, error){
	return s.repo.GetAll()
}

func (s *SaleService)GetById(id string) (*models.Sale, error){
	return s.repo.GetById(id)
}