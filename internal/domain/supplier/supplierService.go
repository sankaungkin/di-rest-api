package supplier

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/models"
)

type SupplierServiceInterface interface {
	CreateSupplier(Supplier *models.Supplier) (*models.Supplier, error)
	GetAllSuppliers() ([]models.Supplier, error)
	GetSupplierById(id uint) (*models.Supplier, error)
	UpdateSupplier(Supplier *models.Supplier) (*models.Supplier, error)
	DeleteSupplier(id uint) error
}

type SupplierService struct {
	repo SupplierRepositoryInterface
}

var (
	svcInstance *SupplierService
	svcOnce sync.Once
)
func NewSupplierService(repo SupplierRepositoryInterface) SupplierServiceInterface{

	log.Println(Green +"SupplierService constructor is called" + Reset)
	svcOnce.Do(func() {

		svcInstance = &SupplierService{repo: repo}
	})
	return svcInstance
}

func (s *SupplierService)CreateSupplier(Supplier *models.Supplier) (*models.Supplier, error){
	return s.repo.Create(Supplier)
}



func (s *SupplierService)GetAllSuppliers() ([]models.Supplier, error){
	return s.repo.GetAll()
}

func (s *SupplierService)GetSupplierById(id uint) (*models.Supplier, error){
	return s.repo.GetById(id)
}

func (s *SupplierService)UpdateSupplier(Supplier *models.Supplier) (*models.Supplier, error){
	return s.repo.Update(Supplier)
}


func (s *SupplierService)DeleteSupplier(id uint) error{
	return s.repo.Delete(id)
}
