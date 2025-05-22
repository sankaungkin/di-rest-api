package inventory

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type InventoryServiceInterface interface {
	IncreaseInventoryService(inventory *models.Inventory) (string, error)
	DecreaseInventoryService(inventory *models.Inventory) (string, error)
	GetAllService() ([]models.Product, error)
	GetInvData() ([]ResponseInventoryDTO, error)
}

type InventoryService struct {
	repo InventoryRepositoryInterface
}

var (
	svcInstance *InventoryService
	svcOnce     sync.Once
)

func NewInventoryService(repo InventoryRepositoryInterface) InventoryServiceInterface {
	log.Println(util.Cyan + "InventoryService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &InventoryService{repo: repo}
	})
	return svcInstance
}

func (s *InventoryService) IncreaseInventoryService(inventory *models.Inventory) (string, error) {
	return s.repo.Increase(inventory)
}
func (s *InventoryService) DecreaseInventoryService(inventory *models.Inventory) (string, error) {
	return s.repo.Decrease(inventory)
}
func (s *InventoryService) GetAllService() ([]models.Product, error) {
	return s.repo.Get()
}
func (s *InventoryService) GetInvData() ([]ResponseInventoryDTO, error) {
	return s.repo.GetInvData()
}
