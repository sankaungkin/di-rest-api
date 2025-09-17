package productstock

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductStockRepositoryInterface interface {
	CreateProductStocks(productStock *models.ProductStock) (*models.ProductStock, error)
	// GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetAllProductStocks() ([]models.ProductStock, error)
	GetLowStockProducts() ([]ResponseProductStockDTO, error)
	GetOutOfStockProducts() ([]ResponseProductStockDTO, error)
	// GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	GetProductStocksById(productId string) (models.ProductStock, error)
	UpdateProductStocksById(productStock UpdateProductStockDTO) (*models.ProductStock, error)

	GetDetailsProductStockById(productId string) (*StockResponse, error)
}

type ProductStockRepository struct {
	db *gorm.DB
}

// ! singleton pattern
var (
	repoInstance *ProductStockRepository
	repoOnce     sync.Once
)

// func NewProductStockRepository(db *gorm.DB) ProductStockRepositoryInterface {
// 	return &ProductStockRepository{db: db}
// }

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

// constructor
func NewProductStockRepository(db *gorm.DB) ProductStockRepositoryInterface {
	log.Println(util.Yellow + "ProductStockRepository constructor is called " + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &ProductStockRepository{db: db}
	})
	return repoInstance
}

func (r *ProductStockRepository) GetLowStockProducts() ([]ResponseProductStockDTO, error) {
	var results []ResponseProductStockDTO

	// err := r.db.
	// 	Where("base_qty <= reorder_lvl").
	// 	Find(&results).Error
	err := r.db.
		Table("product_stocks AS p").
		Select(`
			p.product_id,
			item.product_name,
			uc.base_unit,
			p.base_unit_id,
			p.derive_unit_id,
			p.base_qty,
			uc.derive_unit,
			p.derived_qty,
			p.reorder_lvl,
			uc.factor
		`).
		Joins("JOIN unit_conversions uc ON p.product_id = uc.product_id").
		Joins("JOIN products item ON p.product_id = item.id").
		Where("p.base_qty <= p.reorder_lvl").
		Order("p.product_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProductStockRepository) GetOutOfStockProducts() ([]ResponseProductStockDTO, error) {
	// var results []models.ProductStock

	// err := r.db.
	// 	Where("base_qty = 0").
	// 	Find(&results).Error
	var results []ResponseProductStockDTO

	// err := r.db.
	// 	Where("base_qty <= reorder_lvl").
	// 	Find(&results).Error
	err := r.db.
		Table("product_stocks AS p").
		Select(`
			p.product_id,
			item.product_name,
			uc.base_unit,
			p.base_unit_id,
			p.derive_unit_id,
			p.base_qty,
			uc.derive_unit,
			p.derived_qty,
			p.reorder_lvl,
			uc.factor
		`).
		Joins("JOIN unit_conversions uc ON p.product_id = uc.product_id").
		Joins("JOIN products item ON p.product_id = item.id").
		Where("p.base_qty <= 0").
		Order("p.product_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProductStockRepository) GetAllProductStocks() ([]models.ProductStock, error) {
	var results []models.ProductStock

	err := r.db.
		Preload("Product").
		Preload("UnitOfMeasure").
		Find(&results).Error

	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no records found")
	}
	return results, nil
}

func (r *ProductStockRepository) GetProductStocksById(productId string) (models.ProductStock, error) {
	var result models.ProductStock

	err := r.db.
		Preload("Product").
		Preload("UnitOfMeasure").
		Where("product_id = ?", strings.ToUpper(productId)).
		Find(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.ProductStock{}, errors.New("no product found for this id")
		}
		return models.ProductStock{}, err
	}
	return result, nil
}

func (r *ProductStockRepository) GetAllProductStocksOld() ([]ResponseProductStockDTO, error) {
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
			p.base_unit_id,
			p.derive_unit_id,
			p.base_qty,
			uc.derive_unit,
			p.derived_qty,
			p.reorder_lvl,
			uc.factor
		`).
		Joins("JOIN unit_conversions uc ON p.product_id = uc.product_id").
		Joins("JOIN products item ON p.product_id = item.id").
		Where("p.base_qty > p.reorder_lvl").
		Order("p.id DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *ProductStockRepository) CreateProductStocks(productStock *models.ProductStock) (*models.ProductStock, error) {
	err := r.db.Create(&productStock).Error
	return productStock, err
}

func (r *ProductStockRepository) UpdateProductStocksById(productStock UpdateProductStockDTO) (*models.ProductStock, error) {
	var existingProductStock models.ProductStock
	err := r.db.Where("product_id = ?", strings.ToUpper(productStock.ProductID)).First(&existingProductStock).Error
	if err != nil {
		return nil, err
	}
	existingProductStock.DerivedQty = productStock.DerivedQty
	existingProductStock.ReorderLvl = productStock.ReorderLvl

	log.Println("existingProductStock to update: ", existingProductStock)
	err = r.db.Save(&existingProductStock).Error
	if err != nil {
		return nil, err
	}

	return &existingProductStock, nil
}

func (r *ProductStockRepository) GetDetailsProductStockByIdOld(productId string) ([]DisplayStock, error) {
	// 1. Load stock
	var stock models.ProductStock
	if err := r.db.
		Preload("Product").
		Where("product_id = ?", productId).
		First(&stock).Error; err != nil {
		return nil, err
	}

	// 2. Load product units + preload UnitOfMeasure
	var units []models.ProductUnit
	if err := r.db.
		Preload("Product").
		Preload("UnitOfMeasure").
		Where("product_id = ?", productId).
		Find(&units).Error; err != nil {
		return nil, err
	}

	// 3. Convert
	return ConvertStockToUnits(stock.BaseQty, units), nil

}

func (r *ProductStockRepository) GetDetailsProductStockById(productId string) (*StockResponse, error) {
	// 1. Load stock (always stored in smallest unit)
	var stock models.ProductStock
	if err := r.db.Where("product_id = ?", productId).
		First(&stock).Error; err != nil {
		return nil, err
	}

	// 2. Load product units with relations
	var units []models.ProductUnit
	if err := r.db.
		Preload("UnitOfMeasure").
		Preload("Product").
		Where("product_id = ?", productId).
		Find(&units).Error; err != nil {
		return nil, err
	}

	if len(units) == 0 {
		return nil, fmt.Errorf("no units found for product %s", productId)
	}

	// 3. Convert
	displayUnits := ConvertStockToUnits(stock.DerivedQty, units)

	// 4. Response
	resp := &StockResponse{
		ProductName: units[0].Product.ProductName,
		ProductId:   units[0].Product.ID,
		Quantity:    stock.DerivedQty,
		ReorderLvl:  stock.ReorderLvl,
		Units:       displayUnits,
		Message:     "Record found",
		Status:      "SUCCESS",
	}

	return resp, nil
}

func ConvertStockToUnits(stockQty int, units []models.ProductUnit) []DisplayStock {
	// Sort by ConversionToBase descending
	sort.Slice(units, func(i, j int) bool {
		return units[i].ConversionToBase > units[j].ConversionToBase
	})

	var result []DisplayStock
	remainder := stockQty

	for _, u := range units {
		if remainder >= u.ConversionToBase {
			qty := remainder / u.ConversionToBase
			remainder = remainder % u.ConversionToBase
			result = append(result, DisplayStock{
				UnitName: u.UnitOfMeasure.UnitName, // ✅ works if preloaded
				Quantity: qty,
			})
		}
	}

	// handle leftover smallest unit
	if remainder > 0 {
		for _, u := range units {
			if u.ConversionToBase == 1 {
				result = append(result, DisplayStock{
					UnitName: u.UnitOfMeasure.UnitName,
					Quantity: remainder,
				})
				break
			}
		}
	}

	return result
}
