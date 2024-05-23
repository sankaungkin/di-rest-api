package supplier

import "github.com/sankangkin/di-rest-api/internal/models"

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

func NewSupplierService(repo SupplierRepositoryInterface) SupplierServiceInterface{
	return &SupplierService{repo:repo}
}

func (s *SupplierService)CreateSupplier(Supplier *models.Supplier) (*models.Supplier, error){
	return s.repo.CreateSupplier(Supplier)
}



func (s *SupplierService)GetAllSuppliers() ([]models.Supplier, error){
	return s.repo.GetAllSuppliers()
}

func (s *SupplierService)GetSupplierById(id uint) (*models.Supplier, error){
	return s.repo.GetSupplierById(id)
}

func (s *SupplierService)UpdateSupplier(Supplier *models.Supplier) (*models.Supplier, error){
	return s.repo.UpdateSupplier(Supplier)
}


func (s *SupplierService)DeleteSupplier(id uint) error{
	return s.repo.DeleteSupplier(id)
}
