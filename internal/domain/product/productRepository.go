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
	CreateProductWithDetails(dto *Create_Product_UnitConversion_Stock_Price_DTO) (*Create_Product_UnitConversion_Stock_Price_DTO, error)
	// GetAll() ([]models.Product, error)
	GetAll() ([]ResponseProductDTO, error)
	GetAllWithoutStock() ([]ResponseProductDTO, error)
	GetProductsWithoutPrices() ([]ResponseProductDTO, error)
	GetById(id string) (*models.Product, error)
	GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	GetAllProductPrices() ([]ResponseProductUnitPriceDTO, error)
	GetProductUnitPricesById(productId string) ([]ResponseProductUnitPriceDTO, error)
	GetUnitConversionsById(id string) (models.UnitConversion, error)
	GetAllUnitConversionsWithProductName() ([]UnitConversionWithProductDTO, error)
	GetAllUnitConversions() ([]models.UnitConversion, error)
	UpdateUnitConversion(input *models.UnitConversion) (*models.UnitConversion, error)
	Update(product *models.Product) (*models.Product, error)
	Delete(id string) error
	GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error)
	GetUniofMeasurementById(id string) (models.UnitOfMeasure, error)
	UpdateUnit(input *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	GetProductPriceHistoryByProductId(productId string) ([]ResponseProductHistoryDTO, error)
	GetAllProductPriceHistory() ([]ResponseProductHistoryDTO, error)
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

func (s *ProductRepository) CreateProductWithDetails(dto *Create_Product_UnitConversion_Stock_Price_DTO) (*Create_Product_UnitConversion_Stock_Price_DTO, error) {
	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Create Product
	product := &models.Product{
		ID:          dto.ID,
		ProductName: dto.ProductName,
		CategoryId:  dto.CategoryId,
		Uom:         dto.Uom,
		UomId:       dto.UomId,
		DeriveUom:   dto.DeriveUom,
		DeriveUomId: dto.DeriveUomId,
		BrandName:   dto.BrandName,
		IsActive:    dto.IsActive,
	}

	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// 2. Create Unit Conversion
	unitConversion := &models.UnitConversion{
		Description:  dto.Description,
		ProductId:    dto.ID,
		BaseUnit:     dto.BaseUnitName,
		DeriveUnit:   dto.DeriveUom,
		BaseUnitId:   int(dto.BaseUnitId),
		DeriveUnitId: int(dto.DeriveUomId),
		Factor:       dto.Factor,
	}

	if err := tx.Create(unitConversion).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create unit conversion: %w", err)
	}

	// Update DTO with generated IDs from unit conversion
	dto.Description = unitConversion.Description
	dto.BaseUnitId = uint(unitConversion.BaseUnitId)
	dto.DeriveUnitId = uint(unitConversion.DeriveUnitId)
	dto.Factor = unitConversion.Factor

	// 3. Create Product Stock
	productStock := &models.ProductStock{
		ProductId:    dto.ID,
		BaseUnitId:   int(dto.BaseUnitId),
		DeriveUnitId: int(dto.DeriveUnitId),
		BaseQty:      dto.BaseQty,
		DerivedQty:   dto.DerivedQty,
		ReorderLvl:   dto.ReorderLvl,
	}

	if err := tx.Create(productStock).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create product stock: %w", err)
	}

	// Update DTO with stock information
	dto.BaseQty = productStock.BaseQty
	dto.DerivedQty = productStock.DerivedQty
	dto.ReorderLvl = productStock.ReorderLvl

	// 4. Create Product Price
	// productPrice := &models.ProductPrice{
	// 	ProductId: dto.ID,
	// 	UnitId:    dto.BaseUnitId,
	// 	PriceType: dto.PriceType,
	// 	UnitPrice: dto.Price,
	// }

	// if err := tx.Create(productPrice).Error; err != nil {
	// 	tx.Rollback()
	// 	return nil, fmt.Errorf("failed to create product price: %w", err)
	// }

	for _, price := range dto.Prices {
		productPrice := &models.ProductPrice{
			ProductId: dto.ID,
			PriceType: price.PriceType,
			UnitId:    price.UnitId,
			UnitPrice: price.Price,
		}

		if err := tx.Create(productPrice).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create %s price: %w", price.PriceType, err)
		}

		// Create price history
		priceHistory := &models.ProductPriceHistory{
			ProductId:     dto.ID,
			UnitId:        price.UnitId,
			PriceType:     price.PriceType,
			UnitPrice:     price.Price,
			EffectiveDate: time.Now().Local().Format("2006-01-02"),
		}
		if err := tx.Create(priceHistory).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create price history: %w", err)
		}
	}

	// // Update DTO with price information
	// dto.PriceType = productPrice.PriceType
	// dto.Price = productPrice.UnitPrice

	// // 5. Create Price History (as part of the same transaction)
	// priceHistory := &models.ProductPriceHistory{
	// 	ProductId:     productPrice.ProductId,
	// 	UnitId:        productPrice.UnitId,
	// 	PriceType:     productPrice.PriceType,
	// 	UnitPrice:     productPrice.UnitPrice,
	// 	EffectiveDate: time.Now(),
	// 	CreatedAt:     time.Now(),
	// }

	// if err := tx.Create(priceHistory).Error; err != nil {
	// 	tx.Rollback()
	// 	return nil, fmt.Errorf("failed to create price history: %w", err)
	// }

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the updated DTO with all generated fields
	return dto, nil
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
			UomId:       p.UomId,
			BaseUnit:    p.Uom,
			DeriveUnit:  p.DeriveUom,
			DeriveUomId: p.DeriveUomId,

			BuyPrice:        p.BuyPrice,
			SellPriceLevel1: p.SellPriceLevel1,
			DeriveUnitPrice: p.DeriveUnitPrice,
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

func (r *ProductRepository) GetAllWithoutStock() ([]ResponseProductDTO, error) {
	var products []models.Product

	err := r.db.
		Model(&models.Product{}).
		Select("products.*").
		Joins("LEFT JOIN product_stocks ON products.id = product_stocks.product_id").
		Where("product_stocks.product_id IS NULL").
		Order("products.id DESC").
		Find(&products).Error

	if err != nil {
		return nil, err
	}
	// if len(products) == 0 {
	// 	return nil, errors.New("no records found")
	// }

	var dtos []ResponseProductDTO
	for _, p := range products {
		dto := ResponseProductDTO{
			ID:              p.ID,
			ProductName:     p.ProductName,
			CategoryId:      p.CategoryId,
			UomId:           p.UomId,
			BaseUnit:        p.Uom,
			DeriveUnit:      p.DeriveUom,
			DeriveUomId:     p.DeriveUomId,
			BuyPrice:        p.BuyPrice,
			SellPriceLevel1: p.SellPriceLevel1,
			DeriveUnitPrice: p.DeriveUnitPrice,
			BrandName:       p.BrandName,
			IsActive:        p.IsActive,
			CreatedAt:       time.UnixMilli(p.CreatedAt).Format("2006-01-02 15:04:05"),
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (r *ProductRepository) GetProductsWithoutPrices() ([]ResponseProductDTO, error) {
	var products []models.Product

	err := r.db.
		Model(&models.Product{}).
		Select("products.*").
		Joins("LEFT JOIN product_prices ON products.id = product_prices.product_id").
		Where("product_prices.product_id IS NULL").
		Order("products.id DESC").
		Find(&products).Error

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
			UomId:           p.UomId,
			BaseUnit:        p.Uom,
			DeriveUnit:      p.DeriveUom,
			DeriveUomId:     p.DeriveUomId,
			BuyPrice:        p.BuyPrice,
			SellPriceLevel1: p.SellPriceLevel1,
			DeriveUnitPrice: p.DeriveUnitPrice,
			BrandName:       p.BrandName,
			IsActive:        p.IsActive,
			CreatedAt:       time.UnixMilli(p.CreatedAt).Format("2006-01-02 15:04:05"),
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
	if input.BrandName == "" || input.ProductName == "" || input.CategoryId == 0 {
		return nil, fmt.Errorf("missing required fields")
	}

	existingProduct.BrandName = input.BrandName
	existingProduct.ProductName = input.ProductName
	existingProduct.Uom = input.Uom
	existingProduct.UomId = input.UomId
	existingProduct.DeriveUom = input.DeriveUom
	existingProduct.DeriveUomId = input.DeriveUomId
	existingProduct.IsActive = input.IsActive
	// existingProduct.BuyPrice = input.BuyPrice
	existingProduct.CategoryId = input.CategoryId
	// existingProduct.SellPriceLevel1 = input.SellPriceLevel1
	// existingProduct.DeriveUnitPrice = input.DeriveUnitPrice
	// existingProduct.ReorderLvl = input.ReorderLvl

	log.Println("existingProduct to update: ", existingProduct)
	err = r.db.Save(&existingProduct).Error
	if err != nil {
		return nil, err
	}

	return &existingProduct, nil
}

func (r *ProductRepository) UpdateUnit(input *models.UnitOfMeasure) (*models.UnitOfMeasure, error) {
	var existingUnit models.UnitOfMeasure
	err := r.db.Where("id = ?", input.ID).First(&existingUnit).Error
	if err != nil {
		return nil, err
	}

	existingUnit.UnitName = input.UnitName

	log.Println("existingUnit to update: ", existingUnit)
	err = r.db.Save(&existingUnit).Error
	if err != nil {
		return nil, err
	}

	return &existingUnit, nil
}

func (r *ProductRepository) UpdateUnitConversion(input *models.UnitConversion) (*models.UnitConversion, error) {
	var existingUnit models.UnitConversion
	err := r.db.Where("id = ?", input.ID).First(&existingUnit).Error
	if err != nil {
		return nil, err
	}

	existingUnit.BaseUnit = input.BaseUnit
	existingUnit.DeriveUnit = input.DeriveUnit
	existingUnit.BaseUnitId = input.BaseUnitId
	existingUnit.DeriveUnitId = input.DeriveUnitId
	existingUnit.Factor = input.Factor

	log.Println("existingUnit to update: ", existingUnit)
	err = r.db.Save(&existingUnit).Error
	if err != nil {
		return nil, err
	}

	return &existingUnit, nil
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

func (r *ProductRepository) GetAllUnitConversionsWithProductName() ([]UnitConversionWithProductDTO, error) {
	var result []UnitConversionWithProductDTO

	err := r.db.
		Table("unit_conversions as uc").
		Select("uc.id, uc.product_id, p.product_name, uc.description, uc.base_unit, uc.derive_unit, uc.factor").
		Joins("join products p on uc.product_id = p.id").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *ProductRepository) GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error) {
	var unitOfMeasures []models.UnitOfMeasure
	err := r.db.Model(&models.UnitOfMeasure{}).Order("ID asc").Limit(100).Find(&unitOfMeasures).Error
	if err != nil {
		return nil, err
	}
	return unitOfMeasures, nil
}

func (r *ProductRepository) GetUniofMeasurementById(id string) (models.UnitOfMeasure, error) {
	var unitOfMeasure models.UnitOfMeasure
	err := r.db.Where("id = ?", strings.ToUpper(id)).Find(&unitOfMeasure).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.UnitOfMeasure{}, errors.New("no unit of measurement found for this id")
		}
		return models.UnitOfMeasure{}, err
	}

	return unitOfMeasure, nil
}

func (r *ProductRepository) GetAllProductPriceHistory() ([]ResponseProductHistoryDTO, error) {
	var productPriceHistory []ResponseProductHistoryDTO

	err := r.db.
		Table("product_price_histories").
		Select(`
			product_price_histories.product_id,
			p.product_name,
			uom.unit_name,
			product_price_histories.unit_id,
			product_price_histories.unit_price,
			product_price_histories.effective_date,
			product_price_histories.price_type
		`).
		Joins("JOIN products p ON product_price_histories.product_id = p.id").
		Joins("JOIN unit_of_measures uom ON product_price_histories.unit_id = uom.id").
		Order("product_price_histories.product_id, product_price_histories.effective_date DESC").
		Scan(&productPriceHistory).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []ResponseProductHistoryDTO{}, nil
		}
		return []ResponseProductHistoryDTO{}, err
	}
	return productPriceHistory, nil
}

func (r *ProductRepository) GetProductPriceHistoryByProductId(productId string) ([]ResponseProductHistoryDTO, error) {
	var productPriceHistory []ResponseProductHistoryDTO

	err := r.db.
		Table("product_price_histories").
		Select(`
			product_price_histories.product_id,
			p.product_name,
			uom.unit_name,
			product_price_histories.unit_id,
			product_price_histories.unit_price,
			product_price_histories.effective_date,
			product_price_histories.price_type
		`).
		Joins("JOIN products p ON product_price_histories.product_id = p.id").
		Joins("JOIN unit_of_measures uom ON product_price_histories.unit_id = uom.id").
		Where("product_price_histories.product_id = ?", productId).
		Order("product_price_histories.product_id, product_price_histories.effective_date DESC").
		Scan(&productPriceHistory).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []ResponseProductHistoryDTO{}, nil
		}
		return []ResponseProductHistoryDTO{}, err
	}
	return productPriceHistory, nil
}
