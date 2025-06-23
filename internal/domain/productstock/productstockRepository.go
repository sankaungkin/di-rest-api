package productstock

import (
	"log"
	"strings"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductStockRepositoryInterface interface {
	GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	UpdateProductStocksById(productStock *models.ProductStock) (*models.ProductStock, error)
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

func (r *ProductStockRepository) GetAllProductStocks() ([]ResponseProductStockDTO, error) {
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

func (r *ProductStockRepository) GetProductStocksById(productId string) (*ResponseProductStockDTO, error) {
	var result ResponseProductStockDTO

	err := r.db.
		Table("product_stocks").
		Select("product_stocks.product_id, products.product_name, product_stocks.base_qty , product_stocks.derived_qty, product_stocks.reorder_lvl").
		Joins("JOIN products ON products.id = product_stocks.product_id").
		Where("product_stocks.product_id = ?", strings.ToUpper(productId)).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *ProductStockRepository) UpdateProductStocksById(productStock *models.ProductStock) (*models.ProductStock, error) {
	var existingProductStock models.ProductStock
	err := r.db.Where("product_id = ?", strings.ToUpper(productStock.ProductId)).First(&existingProductStock).Error
	if err != nil {
		return nil, err
	}

	// log.Println("input from Repository: ", productStock)
	// if productStock.BaseQty == 0 || productStock.DerivedQty == 0 || productStock.ReorderLvl == 0 {
	// 	return nil, fmt.Errorf("missing required fields")
	// }

	existingProductStock.BaseQty = productStock.BaseQty
	existingProductStock.DerivedQty = productStock.DerivedQty
	existingProductStock.ReorderLvl = productStock.ReorderLvl

	log.Println("existingProductStock to update: ", existingProductStock)
	err = r.db.Save(&existingProductStock).Error
	if err != nil {
		return nil, err
	}

	return &existingProductStock, nil
}
