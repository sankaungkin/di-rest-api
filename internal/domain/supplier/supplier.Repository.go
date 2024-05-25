package supplier

import (
	"errors"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type SupplierRepositoryInterface interface{
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
	repoOnce sync.Once
	Reset = "\033[0m" 
	Green = "\033[32m" 
)

func NewSupplierRepository(db *gorm.DB) SupplierRepositoryInterface{
	log.Println(Green + "SupplierRepository constructor is called" + Reset)
	repoOnce.Do(func ()  {
		repoInstance = &SupplierRepository{db: db}
	})

	return repoInstance
}

func(r *SupplierRepository)Create(supplier *models.Supplier) (*models.Supplier, error){
	// input := new(CreateSupplierRequestDTO)
	// newSupplier := &models.Supplier{
	// 	Name: input.Name,
	// 	Address: input.Address,
	// 	Phone: input.Phone,
	// }

	// err := r.db.Create(newSupplier)
	// if err != nil {
	// 	return nil, err.Error
	// }

	// return newSupplier, nil

	err := r.db.Create(&supplier).Error

	return supplier, err
}

func(r *SupplierRepository)GetAll() ([]models.Supplier, error){
	suppliers := []models.Supplier{}
	r.db.Model(&models.Supplier{}).Order("ID asc").Find(&suppliers)
	if len(suppliers) == 0 {
		return nil, errors.New("NO records found")
	}
	return suppliers, nil
}

func(r *SupplierRepository)GetById(id uint) (*models.Supplier, error){
	var supplier models.Supplier
	result := r.db.First(&supplier, "id = ?",id)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &supplier, nil
}


func(r *SupplierRepository)Update(input *models.Supplier) (*models.Supplier, error){
	// var updateSupplier *models.Supplier
	// err := r.db.First(&updateSupplier,"id = ?",Supplier.ID)
	// if err != nil {
	// 	return nil, err.Error
	// }
	// r.db.Save(&updateSupplier)
	// return updateSupplier, nil
	var existingSupplier *models.Supplier
		err := r.db.Where("id = ?", input.ID).First(&existingSupplier).Error
		if err != nil {
			// Handle error if customer not found or other issue
			return nil, err
		}

		log.Println("input: ", input)
		if input.Name == "" || input.Address == "" || input.Phone == ""  {
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

func(r *SupplierRepository)Delete(id uint) error {

	return r.db.Delete(&models.Supplier{}, id).Error
	
}