package supplier

import (
	"errors"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type SupplierRepositoryInterface interface {
	Create(supplier *models.Supplier) (*models.Supplier, error)
	GetAll() ([]models.Supplier, error)
	GetById(id uint) (*models.Supplier, error)
	Update(Supplier *models.Supplier) (*models.Supplier, error)
	Delete(id uint) error
}

type SupplierRepository struct {
	db *gorm.DB
}

var (
	repoInstance *SupplierRepository
	repoOnce     sync.Once
)

func NewSupplierRepository(db *gorm.DB) SupplierRepositoryInterface {
	log.Println(util.Green + "SupplierRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &SupplierRepository{db: db}
	})

	return repoInstance
}

// Create godoc
//
//	@Summary		Create new supplier
//	@Description	Create a new supplier with name, address, and phone
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Param			supplier		body		CreateSupplierRequestDTO	true	"Supplier Input Data"
//	@Success		200			{object}	models.Supplier
//	@Failure		400			{object}	httputil.HttpError400
//	@Failure		401			{object}	httputil.HttpError401
//	@Failure		500			{object}	httputil.HttpError500
//	@Router			/api/suppliers [post]
//	@Security		Bearer
func (r *SupplierRepository) Create(supplier *models.Supplier) (*models.Supplier, error) {
	err := r.db.Create(&supplier).Error

	return supplier, err
}

// GetAll godoc
//
//	@Summary		Fetch all suppliers
//	@Description	Fetch all suppliers
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Supplier
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/suppliers	[get]
//	@Security		Bearer
func (r *SupplierRepository) GetAll() ([]models.Supplier, error) {
	suppliers := []models.Supplier{}
	r.db.Model(&models.Supplier{}).Order("ID asc").Find(&suppliers)
	if len(suppliers) == 0 {
		return nil, errors.New("NO records found")
	}
	return suppliers, nil
}

func (r *SupplierRepository) GetById(id uint) (*models.Supplier, error) {
	var supplier models.Supplier
	result := r.db.First(&supplier, "id = ?", id)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &supplier, nil
}

// Update godoc
//
//	@Summary		Update individual supplier
//	@Description	Update individual supplier
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"supplier Id"
//	@Param			supplier				body		models.Supplier	true	"Supplier Data"
//	@Success		200					{object}	models.Supplier
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/suppliers/{id}	[put]
//	@Security		Bearer
func (r *SupplierRepository) Update(input *models.Supplier) (*models.Supplier, error) {
	var existingSupplier *models.Supplier
	err := r.db.Where("id = ?", input.ID).First(&existingSupplier).Error
	if err != nil {
		// Handle error if customer not found or other issue
		return nil, err
	}

	log.Println("input: ", input)
	if input.Name == "" || input.Address == "" || input.Phone == "" {
		return nil, err
	}
	// Update relevant fields from input data
	existingSupplier.Name = input.Name
	existingSupplier.Address = input.Address
	existingSupplier.Phone = input.Phone
	// Save the updated customer data
	err = r.db.Updates(&existingSupplier).Error
	if err != nil {
		// Handle error if update fails
		return nil, err
	}

	// Return the updated customer object
	return existingSupplier, nil
}

// Delete godoc
//
//	@Summary		Delete individual supplier
//	@Description	Delete individual supplier
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"supplier Id"
//	@Success		200					{object}	models.Supplier
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/suppliers/{id}	[delete]
//	@Security		Bearer
func (r *SupplierRepository) Delete(id uint) error {

	return r.db.Delete(&models.Supplier{}, id).Error

}
