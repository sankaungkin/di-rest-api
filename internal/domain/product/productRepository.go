package product

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	Create(product *models.Product) (*models.Product, error)
	// GetAll() ([]models.Product, error)
	GetAll() ([]ResponseProductDTO, error)
	GetById(id string) (*models.Product, error)
	GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetAllProductPrices() ([]ResponseProductUnitPriceDTO, error)
	GetProductUnitPricesById(productId string) ([]ResponseProductUnitPriceDTO, error)
	GetUnitConversionsById(id string) ([]models.UnitConversion, error)
	Update(product *models.Product) (*models.Product, error)
	Delete(id string) error
}

type ProductRepository struct {
	db *gorm.DB
}

// ! singleton pattern
var (
	repoInstance *ProductRepository
	repoOnce     sync.Once
)

// func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
// 	return &ProductRepository{db: db}
// }

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

// constructor
func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	log.Println(util.Yellow + "ProductRepository constructor is called " + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &ProductRepository{db: db}
	})
	return repoInstance
}

func (r *ProductRepository) Create(product *models.Product) (*models.Product, error) {
	err := r.db.Create(&product).Error
	return product, err
}

// func (r *ProductRepository)GetAll() ([]models.Product, error){
// 	products := []models.Product{}
// 	r.db.Model(&models.Product{}).Order("ID asc").Find(&products)
// 	if len(products) == 0 {
// 		return nil, errors.New("NO records found")
// 	}
// 	return products, nil
// }

func (r *ProductRepository) GetAll() ([]ResponseProductDTO, error) {
	var products []models.Product
	err := r.db.Model(&models.Product{}).Order("id asc").Find(&products).Error
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, errors.New("no records found")
	}

	var dtos []ResponseProductDTO
	for _, p := range products {
		dto := ResponseProductDTO{
			ID:              p.ID,
			ProductName:     p.ProductName,
			CategoryId:      p.CategoryId,
			Uom:             p.Uom,
			BuyPrice:        p.BuyPrice,
			SellPriceLevel1: p.SellPriceLevel1,
			SellPriceLevel2: p.SellPriceLevel2,
			// ReorderLvl:      p.ReorderLvl,
			// QtyOnHand:       p.QtyOnHand,
			BrandName: p.BrandName,
			IsActive:  p.IsActive,
			CreatedAt: time.UnixMilli(p.CreatedAt).Format("2006-01-02 15:04:05"),
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (r *ProductRepository) GetById(id string) (*models.Product, error) {

	var product models.Product
	result := r.db.First(&product, "id = ?", strings.ToUpper(id))
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &product, nil
}

// func (r *ProductRepository) GetProductUnitPricesById(productId string) ([]ResponseProductUnitPriceDTO, error) {
// 	var results []ResponseProductUnitPriceDTO

// 	err := r.db.
// 		Table("product_prices AS pp").
// 		Select("p.id AS product_id, p.product_name, u.unit_name AS uom, pp.unit_price").
// 		Joins("JOIN products AS p ON pp.product_id = p.id").
// 		Joins("JOIN unit_of_measures AS u ON pp.unit_id = u.id").
// 		Where("pp.product_id = ?", strings.ToUpper(productId)).
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	// Add serial numbers
// 	for i := range results {
// 		results[i].Serial = i + 1
// 	}

// 	return results, nil
// }

func (r *ProductRepository) GetProductUnitPricesById(productId string) ([]ResponseProductUnitPriceDTO, error) {
	var results []ResponseProductUnitPriceDTO

	err := r.db.
		Table("product_prices AS pp").
		Select("p.id AS product_id, p.product_name, u.unit_name AS uom, pp.unit_price").
		Joins("JOIN products AS p ON pp.product_id = p.id").
		Joins("JOIN unit_of_measures AS u ON pp.unit_id = u.id").
		Where("pp.product_id = ?", strings.ToUpper(productId)).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Add serial numbers after scan
	for i := range results {
		results[i].Serial = i + 1
	}

	return results, nil
}

func (r *ProductRepository) Update(input *models.Product) (*models.Product, error) {

	var existingProduct *models.Product
	err := r.db.Where("id = ?", input.ID).First(&existingProduct).Error
	if err != nil {
		// Handle error if customer not found or other issue
		return nil, err
	}

	log.Println("input: ", input)
	if input.BrandName == "" || input.ProductName == "" || input.Uom == "" || input.BuyPrice == 0 || input.CategoryId == 0 || input.SellPriceLevel1 == 0 || input.SellPriceLevel2 == 0 {
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
	// existingProduct.ReorderLvl = input.ReorderLvl

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
func (r *ProductRepository) Delete(id string) error {
	// return r.db.Delete(&User{}, id).Error

	var product models.Product
	result := r.db.First(&product, "id = ?", id)

	if err := result.Error; err != nil {
		return err
	}

	return r.db.Delete(&product).Error

}

func (r *ProductRepository) GetAllProductStocks() ([]ResponseProductStockDTO, error) {
	var results []ResponseProductStockDTO

	// Perform the join and select necessary fields
	err := r.db.
		Table("product_stocks").
		Select("product_stocks.product_id, products.product_name, product_stocks.base_qty as base_uom_in_stock, product_stocks.derived_qty as derived_uom_in_stock, product_stocks.reorder_lvl as reorder").
		Joins("JOIN products ON products.id = product_stocks.product_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProductRepository) GetAllProductPrices() ([]ResponseProductUnitPriceDTO, error) {
	var results []ResponseProductUnitPriceDTO

	err := r.db.
		Raw(`
		SELECT 
			ROW_NUMBER() OVER (ORDER BY pp.id) AS serial,
			p.id AS product_id,
			p.product_name,
			u.unit_name AS uom,
			pp.unit_price
		FROM product_prices AS pp
		JOIN products AS p ON pp.product_id = p.id
		JOIN unit_of_measures AS u ON pp.unit_id = u.id
	`).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProductRepository) GetUnitConversionsById(id string) ([]models.UnitConversion, error) {
	var unitConversions []models.UnitConversion
	err := r.db.Where("product_id = ?", strings.ToUpper(id)).Find(&unitConversions).Error
	if err != nil {
		return nil, err
	}
	if len(unitConversions) == 0 {
		return nil, errors.New("no unit conversions found")
	}
	return unitConversions, nil
}
