package supplier

import (
	"errors"

	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type SupplierRepositoryInterface interface{
	CreateSupplier(supplier *models.Supplier) (*models.Supplier, error)
	GetAllSuppliers() ([]models.Supplier, error)
	GetSupplierById(id uint) (*models.Supplier, error)
	UpdateSupplier(Supplier *models.Supplier) (*models.Supplier, error)
	DeleteSupplier(id uint) error
}

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) SupplierRepositoryInterface{
	return &SupplierRepository{db: db}
}

func(r *SupplierRepository)CreateSupplier(Supplier *models.Supplier) (*models.Supplier, error){
	input := new(CreateSupplierRequestDTO)
	newSupplier := &models.Supplier{
		Name: input.Name,
		Address: input.Address,
		Phone: input.Phone,
	}

	err := r.db.Create(newSupplier)
	if err != nil {
		return nil, err.Error
	}

	return newSupplier, nil
}

	func(r *SupplierRepository)GetAllSuppliers() ([]models.Supplier, error){
		Suppliers := []models.Supplier{}
		r.db.Model(&models.Supplier{}).Order("ID asc").Find(&Suppliers)
		if len(Suppliers) == 0 {
			return nil, errors.New("NO records found")
		}
		return Suppliers, nil
	}

	func(r *SupplierRepository)GetSupplierById(id uint) (*models.Supplier, error){
		var Supplier models.Supplier
		result := r.db.First(&Supplier, "id = ?",id)
		if err := result.Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, err
			}
		}
		return &Supplier, nil
	}


	func(r *SupplierRepository)UpdateSupplier(Supplier *models.Supplier) (*models.Supplier, error){
		var updateSupplier *models.Supplier
		err := r.db.First(&updateSupplier,"id = ?",Supplier.ID)
		if err != nil {
			return nil, err.Error
		}
		r.db.Save(&updateSupplier)
		return updateSupplier, nil
	}

	func(r *SupplierRepository)DeleteSupplier(id uint) error {
		var deleteSupplier *models.Supplier
		err := r.db.First(&deleteSupplier, "id = ?", id)
		if err != nil {
			return err.Error
		}
		r.db.Delete(&deleteSupplier)
		return nil
	}