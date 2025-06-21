package product

import (
	"errors"
	"fmt"
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
	GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	GetAllProductPrices() ([]ResponseProductUnitPriceDTO, error)
	GetProductUnitPricesById(productId string) ([]ResponseProductUnitPriceDTO, error)
	GetUnitConversionsById(id string) (models.UnitConversion, error)
	GetAllUnitConversions() ([]models.UnitConversion, error)
	Update(product *models.Product) (*models.Product, error)
	Delete(id string) error
	GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error)
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

func (r *ProductRepository) GetAll() ([]ResponseProductDTO, error) {
	var products []models.Product
	err := r.db.Model(&models.Product{}).Order("id DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, errors.New("no records found")
	}

	var dtos []ResponseProductDTO
	for _, p := range products {
		dto := ResponseProductDTO{
			ID:          p.ID,
			ProductName: p.ProductName,
			CategoryId:  p.CategoryId,
			// Uom:             p.Uom,
			UomId:           p.UomId,
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
	var existingProduct models.Product
	err := r.db.Where("id = ?", input.ID).First(&existingProduct).Error
	if err != nil {
		return nil, err
	}

	log.Println("input from Repository: ", input)
	if input.BrandName == "" || input.ProductName == "" || input.Uom == "" || input.BuyPrice == 0 || input.CategoryId == 0 || input.SellPriceLevel1 == 0 || input.SellPriceLevel2 == 0 {
		return nil, fmt.Errorf("missing required fields")
	}

	existingProduct.BrandName = input.BrandName
	existingProduct.ProductName = input.ProductName
	// existingProduct.Uom = input.Uom
	existingProduct.UomId = input.UomId
	existingProduct.BuyPrice = input.BuyPrice
	existingProduct.CategoryId = input.CategoryId
	existingProduct.SellPriceLevel1 = input.SellPriceLevel1
	existingProduct.SellPriceLevel2 = input.SellPriceLevel2
	// existingProduct.ReorderLvl = input.ReorderLvl

	log.Println("existingProduct to update: ", existingProduct)
	err = r.db.Save(&existingProduct).Error
	if err != nil {
		return nil, err
	}

	return &existingProduct, nil
}

func (r *ProductRepository) Delete(id string) error {
	// return r.db.Delete(&User{}, id).Error

	var product models.Product
	result := r.db.First(&product, "id = ?", id)

	if err := result.Error; err != nil {
		return err
	}

	// return r.db.Delete(&product).Error
	return r.db.Unscoped().Delete(&product).Error

}

func (r *ProductRepository) GetAllProductStocks() ([]ResponseProductStockDTO, error) {
	var results []ResponseProductStockDTO

	// Perform the join and select necessary fields
	// err := r.db.
	// 	Table("product_stocks").
	// 	Select("product_stocks.product_id, products.product_name, product_stocks.base_qty as base_uom_in_stock, product_stocks.derived_qty as derived_uom_in_stock, product_stocks.reorder_lvl as reorder").
	// 	Joins("JOIN products ON products.id = product_stocks.product_id").
	// 	Scan(&results).Error

	err := r.db.
		Table("product_stocks AS p").
		Select(`
			p.product_id,
			item.product_name,
			uc.base_unit,
			p.base_qty,
			uc.derive_unit,
			p.derived_qty,
			p.reorder_lvl,
			uc.factor
		`).
		Joins("JOIN unit_conversions uc ON p.product_id = uc.product_id").
		Joins("JOIN products item ON p.product_id = item.id").
		Order("p.product_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProductRepository) GetProductStocksById(productId string) (*ResponseProductStockDTO, error) {
	var result ResponseProductStockDTO

	err := r.db.
		Table("product_stocks").
		Select("product_stocks.product_id, products.product_name, product_stocks.base_qty as base_uom_in_stock, product_stocks.derived_qty as derived_uom_in_stock, product_stocks.reorder_lvl").
		Joins("JOIN products ON products.id = product_stocks.product_id").
		Where("product_stocks.product_id = ?", strings.ToUpper(productId)).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
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

func (r *ProductRepository) GetUnitConversionsById(id string) (models.UnitConversion, error) {
	var unitConversions models.UnitConversion
	err := r.db.Where("product_id = ?", strings.ToUpper(id)).Find(&unitConversions).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.UnitConversion{}, errors.New("no unit conversion found for this product")
		}
		return models.UnitConversion{}, err
	}
	if unitConversions.ProductId == "" {
		return models.UnitConversion{}, errors.New("no unit conversion found for this product")
	}
	return unitConversions, nil
}

func (r *ProductRepository) GetAllUnitConversions() ([]models.UnitConversion, error) {
	var unitConversions []models.UnitConversion
	err := r.db.Model(&models.UnitConversion{}).Order("ID asc").Limit(100).Find(&unitConversions).Error
	if err != nil {
		return nil, err
	}
	return unitConversions, nil
}

func (r *ProductRepository) GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error) {
	var unitOfMeasures []models.UnitOfMeasure
	err := r.db.Model(&models.UnitOfMeasure{}).Order("ID asc").Limit(100).Find(&unitOfMeasures).Error
	if err != nil {
		return nil, err
	}
	return unitOfMeasures, nil
}
