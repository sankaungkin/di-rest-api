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
	// Create(product *models.Product) (*models.Product, error)
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
	Update(input UpdateProductRequstDTO) (*models.Product, error)
	Delete(id string) error
	GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error)
	GetUniofMeasurementById(id string) (models.UnitOfMeasure, error)
	UpdateUnit(input *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	GetProductPriceHistoryByProductId(productId string) ([]ResponseProductHistoryDTO, error)
	GetAllProductPriceHistory() ([]ResponseProductHistoryDTO, error)
	GetAllProductUnits() (*[]models.ProductUnit, error)
	GetProductUnitById(id string) (*models.ProductUnit, error)
	GetAllProducts() ([]models.Product, error)
}

type ProductRepository struct {
	db *gorm.DB
}

// ! singleton pattern
var (
	repoInstance *ProductRepository
	repoOnce     sync.Once
)

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

// constructor
func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	log.Println(util.Yellow + "ProductRepository constructor is called " + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &ProductRepository{db: db}
	})
	return repoInstance
}

func (s *ProductRepository) CreateProductWithDetails(dto *Create_Product_UnitConversion_Stock_Price_DTO) (*Create_Product_UnitConversion_Stock_Price_DTO, error) {
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
		BrandName:   dto.BrandName,
		IsActive:    dto.IsActive,
	}

	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// 2. Bulk Insert Product Units
	var units []models.ProductUnit
	for _, u := range dto.ProductUnits {
		units = append(units, models.ProductUnit{
			ID:               u.ProductUnitId,
			ProductId:        dto.ID,
			UnitId:           u.UnitId,
			ConversionToBase: u.ConversionToBase,
			IsDefaultUnit:    u.IsDefaultUnit,
		})
	}
	if len(units) > 0 {
		if err := tx.Create(&units).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create product units: %w", err)
		}
	}

	// 3. Create Product Stock
	productStock := &models.ProductStock{
		ProductId:  dto.ID,
		BaseUnitId: dto.BaseUnitId,
		BaseQty:    dto.BaseQty,
		ReorderLvl: dto.ReorderLvl,
	}
	if err := tx.Create(productStock).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create product stock: %w", err)
	}

	// 4. Bulk Insert Product Prices
	var prices []models.ProductPrice
	for _, p := range dto.ProductPrices {
		prices = append(prices, models.ProductPrice{
			ProductId:     dto.ID,
			ProductUnitId: p.ProductUnitId,
			UnitId:        p.UnitId,
			PriceType:     p.PriceType,
			UnitPrice:     p.UnitPrice,
			Remark:        fmt.Sprintf("%s : %s : type = %s changed to %d", p.ProductId, p.ProductUnitId, p.PriceType, p.UnitPrice),
		})
	}
	if len(prices) > 0 {
		if err := tx.Create(&prices).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create product prices: %w", err)
		}
	}

	// 5. Bulk Insert Price Histories
	var histories []models.ProductPriceHistory
	for _, p := range prices {
		histories = append(histories, models.ProductPriceHistory{
			ProductId:     dto.ID,
			ProductName:   product.ProductName,
			UnitId:        p.UnitId,
			ProductUnitId: p.ProductUnitId,
			PriceType:     p.PriceType,
			UnitPrice:     p.UnitPrice,
			Remark:        fmt.Sprintf("%s : %s : %s was changed to %d in %s", p.ProductId, p.ProductUnitId, p.PriceType, p.UnitPrice, time.Now().Local().Format("2006-01-02")),
			EffectiveDate: time.Now().Local().Format("2006-01-02"),
		})
	}
	if len(histories) > 0 {
		if err := tx.Create(&histories).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create price histories: %w", err)
		}
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return dto, nil
}

func (r *ProductRepository) GetAllProducts() ([]models.Product, error) {
	var products []models.Product

	err := r.db.
		Preload("Category").
		Preload("ProductUnits"). // 👈 also preload nested UnitOfMeasure
		Find(&products).Error

	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return nil, errors.New("no records found")
	}
	return products, nil
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
	result := r.db.
		Preload("Category").
		Preload("UnitOfMeasure").Preload("ProductUnits").
		First(&product, "id = ?", strings.ToUpper(id))
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

func (r *ProductRepository) GetAllProductUnits() (*[]models.ProductUnit, error) {
	var results []models.ProductUnit

	err := r.db.
		Preload("UnitOfMeasure").
		Preload("Product").
		Model(&models.ProductUnit{}).
		Order("ID asc").
		Find(&results).Error

	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no records found")
	}

	return &results, nil
}

func (r *ProductRepository) GetProductUnitById(id string) (*models.ProductUnit, error) {
	var result models.ProductUnit

	err := r.db.
		Preload("UnitOfMeasure").
		Preload("Product").
		First(&result, "id = ?", strings.ToUpper(id)).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &result, nil
}

func (r *ProductRepository) UpdateOld(input *models.Product) (*models.Product, error) {
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
	existingProduct.IsActive = input.IsActive
	existingProduct.CategoryId = input.CategoryId

	log.Println("existingProduct to update: ", existingProduct)
	err = r.db.Save(&existingProduct).Error
	if err != nil {
		return nil, err
	}

	return &existingProduct, nil
}

func (r *ProductRepository) Update(input UpdateProductRequstDTO) (*models.Product, error) {
	var existingProduct models.Product

	// 1. Load product with its units
	if err := r.db.Preload("ProductUnits").
		Where("id = ?", input.ProductId).
		First(&existingProduct).Error; err != nil {
		return nil, err
	}

	// 2. Validate required fields
	if input.BrandName == "" || input.ProductName == "" || input.CategoryId == 0 {
		return nil, fmt.Errorf("missing required fields")
	}

	// 3. Update main product fields
	existingProduct.BrandName = input.BrandName
	existingProduct.ProductName = input.ProductName
	existingProduct.IsActive = input.IsActive
	existingProduct.CategoryId = input.CategoryId

	// 4. Update only existing productUnits
	for _, u := range input.ProductUnits {
		var unit models.ProductUnit
		// Find the existing row
		err := r.db.Where("product_id = ? AND id = ?", existingProduct.ID, u.ProductUnitId).
			First(&unit).Error
		if err != nil {
			// if unit doesn’t exist, skip (ignore new ones)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}

		// Update fields
		unit.ConversionToBase = u.ConversionToBase
		unit.IsDefaultUnit = u.IsDefaultUnit
		unit.UnitId = u.UnitId

		if err := r.db.Save(&unit).Error; err != nil {
			return nil, err
		}
	}

	// 5. Save main product
	if err := r.db.Save(&existingProduct).Error; err != nil {
		return nil, err
	}

	// 6. Reload with units for response
	if err := r.db.Preload("ProductUnits").
		First(&existingProduct, "id = ?", input.ProductId).Error; err != nil {
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
