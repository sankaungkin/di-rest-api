package productprice

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductPriceRepositoryInterface interface {
	Create(productPrice *models.ProductPrice) (*models.ProductPrice, error)
	GetAll() ([]ProductPriceResponseDTO, error)
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
	err := r.db.Create(&productPrice).Error
	return productPrice, err
}

func (r *ProductPriceRepository) GetAll() ([]ProductPriceResponseDTO, error) {
	var results []ProductPriceResponseDTO

	err := r.db.
		Table("product_prices AS pp").
		Select(`pp.id, pp.product_id, p.product_name, pp.unit_id, u.unit_name AS unit_name, pp.unit_price`).
		Joins("JOIN products AS p ON pp.product_id = p.id").
		Joins("JOIN unit_of_measures AS u ON pp.unit_id = u.id").
		Order("pp.product_id ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("no records found")
	}

	return results, nil
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
	var existingProductPrice models.ProductPrice
	err := r.db.Where("id = ?", input.ID).First(&existingProductPrice).Error
	if err != nil {
		return nil, err
	}

	log.Println("input from Repository: ", input)
	if input.ProductId == "" || input.UnitId == 0 || input.UnitPrice == 0 {
		return nil, fmt.Errorf("missing required fields")
	}

	existingProductPrice.ProductId = input.ProductId
	existingProductPrice.UnitId = input.UnitId
	existingProductPrice.UnitPrice = input.UnitPrice

	log.Println("existingProductPrice to update: ", existingProductPrice)
	err = r.db.Save(&existingProductPrice).Error
	if err != nil {
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

	// return r.db.Delete(&productPrice).Error
	return r.db.Unscoped().Delete(&productPrice).Error

}
