package product

import (
	"errors"
	"strings"

	"github.com/sankangkin/di-rest-api/internal/dto"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface{
	CreateProduct(product *models.Product) (*models.Product, error)
	GetAllProducts() ([]models.Product, error)
	GetProductById(id string)(*models.Product, error)
	UpdateProduct(product *models.Product) (*models.Product, error)
	DeleteProduct(id string) error
}

type ProductRepository struct{
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &ProductRepository{db: db}
}

func (r *ProductRepository)CreateProduct(product *models.Product) (*models.Product, error){

	input := new(dto.CreateProductRequstDTO)
	newProduct := &models.Product{
		ID: input.ID ,
		ProductName: input.ProductName,
		CategoryId: input.CategoryId,
		Uom: input.Uom,
		BuyPrice: input.BuyPrice,
		SellPriceLevel1: input.SellPriceLevel1,
		SellPriceLevel2: input.SellPriceLevel2,
		ReorderLvl: input.ReorderLvl,
		QtyOnHand: input.QtyOnHand,
		BrandName: input.BrandName,
		IsActive: input.IsActive,
	}

	err := r.db.Create(newProduct)
	if err != nil {
		return nil, err.Error
	}

	return newProduct, nil
}

func (r *ProductRepository)GetAllProducts() ([]models.Product, error){
	products := []models.Product{}
	r.db.Model(&models.Product{}).Order("ID asc").Find(&products)
	if len(products) == 0 {
		return nil, errors.New("NO records found")
	}
	return products, nil
}

func (r *ProductRepository)GetProductById(id string)(*models.Product, error){

	var product models.Product
	result := r.db.First(&product, "id = ?",strings.ToUpper(id))
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &product, nil
}

func (r *ProductRepository)UpdateProduct(product *models.Product) (*models.Product, error){

	var updateProduct *models.Product
	err := r.db.Find(&updateProduct, "id = ?", strings.ToUpper(product.ID))
	if err != nil {
		return nil, err.Error
	}
	r.db.Save(&updateProduct)
	return updateProduct, nil
}

func (r *ProductRepository)DeleteProduct(id string) error {
	
	var deleteProduct *models.Product
	err := r.db.First(&deleteProduct, "id = ?", strings.ToUpper(id))
	if err != nil {
		return err.Error
	}
	r.db.Delete(&deleteProduct)
	return nil
}