package product

import (
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface{
	Create(product *models.Product) (*models.Product, error)
	GetAll() ([]models.Product, error)
	GetById(id string)(*models.Product, error)
	Update(product *models.Product) (*models.Product, error)
	Delete(id string) error
}

type ProductRepository struct{
	db *gorm.DB
}

//! singleton pattern
var (
	repoInstance *ProductRepository
	repoOnce sync.Once
	Reset = "\033[0m" 
	Yellow = "\033[33m"
)

// func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
// 	return &ProductRepository{db: db}
// }

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

// constructor
func NewProductRepository(db *gorm.DB) ProductRepositoryInterface{
	log.Println(Yellow + "ProductRepository constructor is called " + Reset)
	repoOnce.Do(func() {
		repoInstance = &ProductRepository{db: db}
	})
	return repoInstance
}


func (r *ProductRepository)Create(product *models.Product) (*models.Product, error){
	err := r.db.Create(&product).Error
	return product, err
}

func (r *ProductRepository)GetAll() ([]models.Product, error){
	products := []models.Product{}
	r.db.Model(&models.Product{}).Order("ID asc").Find(&products)
	if len(products) == 0 {
		return nil, errors.New("NO records found")
	}
	return products, nil
}

func (r *ProductRepository)GetById(id string)(*models.Product, error){

	var product models.Product
	result := r.db.First(&product, "id = ?",strings.ToUpper(id))
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &product, nil
}

func (r *ProductRepository)Update(input *models.Product) (*models.Product, error){

	var existingProduct *models.Product
		err := r.db.Where("id = ?", input.ID).First(&existingProduct).Error
		if err != nil {
			// Handle error if customer not found or other issue
			return nil, err
		}

		log.Println("input: ", input)
		if input.BrandName == "" || input.ProductName == "" || input.Uom == "" || input.BuyPrice == 0 || input.CategoryId == 0  || input.SellPriceLevel1 == 0 || input.ReorderLvl == 0 || input.SellPriceLevel2 == 0 {
			return nil, err
		}
		// Update relevant fields from input data
		existingProduct.BrandName = input.BrandName
		existingProduct.ProductName = input.ProductName
		existingProduct.Uom = input.Uom
		existingProduct.BuyPrice = input.BuyPrice
		existingProduct.CategoryId = input.CategoryId
		existingProduct.SellPriceLevel1 = input.SellPriceLevel1
		existingProduct.SellPriceLevel2 = input.SellPriceLevel2
		existingProduct.ReorderLvl = input.ReorderLvl

		// Save the updated customer data
		log.Println("existingCustomer: ", existingProduct)
		err = r.db.Updates(&existingProduct).Error
		if err != nil {
			// Handle error if update fails
			return nil, err
		}

		// Return the updated customer object
		return existingProduct, nil
}
func(r *ProductRepository)Delete(id string) error {
	// return r.db.Delete(&User{}, id).Error

	var product models.Product
	result := r.db.First(&product, "id = ?", id)

	if err := result.Error; err != nil {
		return err
	}

	return  r.db.Delete(&product).Error
	
}