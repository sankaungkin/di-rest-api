package productprice

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductPriceRepositoryInterface interface {
	Create(productPrice *models.ProductPrice) (*models.ProductPrice, error)
	GetAll() ([]ProductPriceResponseDTO, error)
	GetAllWithStock() ([]ProductPriceResponseDTO, error)
	GetInventoryTotalValue() (int64, error)
	GetById(id int) (*ProductPriceResponseDTO, error)
	UpdateProductPrice(input *models.ProductPrice) (*models.ProductPrice, error)
	DeleteProductPrice(id int) error
}

type ProductPriceRepository struct {
	db *gorm.DB
}

// ! singleton pattern
var (
	repoInstance *ProductPriceRepository
	repoOnce     sync.Once
)

// func NewProductPriceRepository(db *gorm.DB) ProductPriceRepositoryInterface {
// 	return &ProductPriceRepository{db: db}
// }

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

// constructor
func NewProductPriceRepository(db *gorm.DB) ProductPriceRepositoryInterface {
	log.Println(util.Yellow + "ProductPriceRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &ProductPriceRepository{db: db}
	})
	return repoInstance
}

func (r *ProductPriceRepository) Create(productPrice *models.ProductPrice) (*models.ProductPrice, error) {
	// err := r.db.Create(&productPrice).Error
	// return productPrice, err
	err := r.db.Create(&productPrice).Error
	if err != nil {
		return nil, err
	}

	// Insert into history
	history := models.ProductPriceHistory{
		ProductId:     productPrice.ProductId,
		UnitId:        productPrice.UnitId,
		PriceType:     productPrice.PriceType,
		UnitPrice:     productPrice.UnitPrice,
		EffectiveDate: time.Now(),
		CreatedAt:     time.Now(),
	}
	_ = r.db.Create(&history) // optional: handle error if needed

	return productPrice, nil
}

func (r *ProductPriceRepository) GetAll() ([]ProductPriceResponseDTO, error) {
	var results []ProductPriceResponseDTO

	err := r.db.
		Table("product_prices AS pp").
		Select(`pp.id, pp.product_id, p.product_name, u.unit_name, pp.unit_id, pp.unit_price, pp.price_type`).
		Joins("JOIN products AS p ON pp.product_id = p.id").
		Joins("JOIN unit_of_measures AS u ON pp.unit_id = u.id").
		Order("pp.id DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("no records found")
	}

	return results, nil
}

func (r *ProductPriceRepository) GetAllWithStock() ([]ProductPriceResponseDTO, error) {
	var results []ProductPriceResponseDTO

	err := r.db.
		Table("product_prices AS pp").
		Select(`pp.id, pp.product_id, p.product_name, u.unit_name, pp.unit_id, pp.unit_price, pp.price_type`).
		Joins("JOIN products AS p ON pp.product_id = p.id").
		Joins("JOIN unit_of_measures AS u ON pp.unit_id = u.id").
		Joins("JOIN product_stocks AS ps ON ps.product_id = pp.product_id").
		Where("ps.base_qty > ps.reorder_lvl").
		Order("pp.id DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("no records found")
	}

	return results, nil
}

func (r *ProductPriceRepository) GetInventoryTotalValue() (int64, error) {
	var result int64

	err := r.db.
		Raw(`
			SELECT 
				SUM(
					COALESCE(ps.base_qty, 0) * COALESCE(pp.unit_price, 0)
				) AS inventory_total_value
			FROM product_stocks ps
			JOIN product_prices pp ON ps.product_id = pp.product_id
			WHERE pp.price_type = 'BUY'
		`).Scan(&result).Error

	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *ProductPriceRepository) GetById(id int) (*ProductPriceResponseDTO, error) {
	var result ProductPriceResponseDTO

	err := r.db.
		Table("product_prices AS pp").
		Select(`pp.id, pp.product_id, p.product_name, pp.unit_id, u.unit_name AS unit_name, pp.unit_price`).
		Joins("JOIN products AS p ON pp.product_id = p.id").
		Joins("JOIN unit_of_measures AS u ON pp.unit_id = u.id").
		Where("pp.id = ?", id).
		First(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *ProductPriceRepository) UpdateProductPrice(input *models.ProductPrice) (*models.ProductPrice, error) {
	// Start transaction
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var existingProductPrice models.ProductPrice
	if err := tx.Where("id = ?", input.ID).First(&existingProductPrice).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Validate input
	if input.ProductId == "" || input.UnitId == 0 || input.UnitPrice == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("missing required fields")
	}

	priceChanged := existingProductPrice.UnitPrice != input.UnitPrice

	// Update the product price
	existingProductPrice.ProductId = input.ProductId
	existingProductPrice.UnitId = input.UnitId
	existingProductPrice.UnitPrice = input.UnitPrice
	remark := input.ProductId + ", UnitId: " + fmt.Sprintf("%d", input.UnitId) + ", UnitPrice: " + fmt.Sprintf("%d", input.UnitPrice) + " was changed in time : " + time.Now().Format("2006-01-02 15:04:05")
	existingProductPrice.Remark = remark
	if err := tx.Save(&existingProductPrice).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Insert into product_price_histories if price changed
	if priceChanged {
		history := models.ProductPriceHistory{
			ProductId:     existingProductPrice.ProductId,
			UnitId:        existingProductPrice.UnitId,
			PriceType:     existingProductPrice.PriceType,
			UnitPrice:     input.UnitPrice,
			Remark:        remark,
			EffectiveDate: time.Now(),
			CreatedAt:     time.Now(),
		}

		if err := tx.Create(&history).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create price history: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &existingProductPrice, nil
}

func (r *ProductPriceRepository) DeleteProductPrice(id int) error {
	// return r.db.Delete(&User{}, id).Error

	var productPrice models.ProductPrice
	result := r.db.First(&productPrice, "id = ?", id)

	if err := result.Error; err != nil {
		return err
	}

	// Optional: Record deletion in history
	history := models.ProductPriceHistory{
		ProductId:     productPrice.ProductId,
		UnitId:        productPrice.UnitId,
		PriceType:     productPrice.PriceType,
		UnitPrice:     productPrice.UnitPrice,
		EffectiveDate: time.Now(),
		CreatedAt:     time.Now(),
	}
	_ = r.db.Create(&history)

	// return r.db.Delete(&productPrice).Error
	return r.db.Unscoped().Delete(&productPrice).Error

}
