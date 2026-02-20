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
	GetActiveProducts() ([]models.Product, error)
	GetInActiveProducts() ([]models.Product, error)
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
	UpdateProductUnit(input UpdateProductUnitDTO) (*models.Product, error)
	Delete(id string) error
	GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error)
	GetUniofMeasurementById(id string) (models.UnitOfMeasure, error)
	UpdateUnit(input *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	GetProductPriceHistoryByProductId(productId string) ([]ResponseProductHistoryDTO, error)
	GetAllProductPriceHistory() ([]ResponseProductHistoryDTO, error)
	GetAllProductUnits() (*[]models.ProductUnit, error)
	GetProductUnitByProductId(id string) ([]models.ProductUnit, error)
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
			ID:               u.Id,
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
		ProductId:    dto.ID,
		DeriveUnitId: dto.DeriveUnitId,
		DerivedQty:   dto.Qty,
		ReorderLvl:   dto.ReorderLvl,
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

func (r *ProductRepository) GetActiveProducts() ([]models.Product, error) {
	var products []models.Product

	err := r.db.
		Where("is_active", true).
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

func (r *ProductRepository) GetInActiveProducts() ([]models.Product, error) {
	var products []models.Product

	err := r.db.
		Where("is_active", false).
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
			BrandName:   p.BrandName,
			IsActive:    p.IsActive,
			CreatedAt:   time.UnixMilli(p.CreatedAt).Format("2006-01-02 15:04:05"),
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
			ID:          p.ID,
			ProductName: p.ProductName,
			CategoryId:  p.CategoryId,
			BrandName:   p.BrandName,
			IsActive:    p.IsActive,
			CreatedAt:   time.UnixMilli(p.CreatedAt).Format("2006-01-02 15:04:05"),
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
			ID:          p.ID,
			ProductName: p.ProductName,
			CategoryId:  p.CategoryId,
			BrandName:   p.BrandName,
			IsActive:    p.IsActive,
			CreatedAt:   time.UnixMilli(p.CreatedAt).Format("2006-01-02 15:04:05"),
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func (r *ProductRepository) GetById(id string) (*models.Product, error) {

	var product models.Product
	result := r.db.
		Preload("Category").
		Preload("UnitOfMeasure").Preload("ProductUnits").Preload("ProductUnits.UnitOfMeasure").
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

func (r *ProductRepository) GetProductUnitByProductId(id string) ([]models.ProductUnit, error) {
	var results []models.ProductUnit

	err := r.db.
		Table("product_units").
		Select("id , product_id, unit_id, conversion_to_base, is_default_unit").
		Where("product_id = ?", strings.ToUpper(id)).
		Scan(&results).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return results, nil
}

func (r *ProductRepository) Update(input UpdateProductRequstDTO) (*models.Product, error) {
	// 1. Validate required fields immediately to save DB calls
	if input.BrandName == "" || input.ProductName == "" || input.CategoryId == 0 {
		return nil, errors.New("missing required fields")
	}

	// 2. Map DTO units to Model units
	// We do this first to ensure the data is ready before starting DB operations
	// var updatedUnits []models.ProductUnit
	// for _, u := range input.ProductUnits {
	// 	if u.UnitId == 0 {
	// 		continue
	// 	}
	// 	updatedUnits = append(updatedUnits, models.ProductUnit{
	// 		ID:               u.Id,            // e.g., "P1030FEET"
	// 		ProductId:        input.ProductId, // e.g., "P1030"
	// 		UnitId:           u.UnitId,
	// 		ConversionToBase: u.ConversionToBase,
	// 		IsDefaultUnit:    u.IsDefaultUnit,
	// 	})
	// }

	// 3. Perform the update in a Transaction
	// Use a transaction to ensure that if the units fail to update,
	// the product name/brand changes are also rolled back.
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.Where("id = ?", input.ProductId).First(&product).Error; err != nil {
			return err
		}

		product.BrandName = input.BrandName
		product.ProductName = input.ProductName
		product.IsActive = input.IsActive
		product.CategoryId = input.CategoryId

		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		// 🔥 DELETE old units first
		// if err := tx.Where("product_id = ?", input.ProductId).
		// 	Delete(&models.ProductUnit{}).Error; err != nil {
		// 	return err
		// }

		// 🔥 INSERT new units
		// if len(updatedUnits) > 0 {
		// 	if err := tx.Create(&updatedUnits).Error; err != nil {
		// 		return err
		// 	}
		// }

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. Reload and Return
	var finalProduct models.Product
	r.db.Preload("ProductUnits").Where("id = ?", input.ProductId).First(&finalProduct)
	return &finalProduct, nil
}

func (r *ProductRepository) UpdateProductUnit(input UpdateProductUnitDTO) (*models.Product, error) {
	var product models.Product

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Verify product exists
		if err := tx.Where("id = ?", input.ProductId).First(&product).Error; err != nil {
			return err
		}

		// 2. Map DTO to Models
		var units []models.ProductUnit
		for _, u := range input.ProductUnits {
			units = append(units, models.ProductUnit{
				ID:               u.ID,
				UnitId:           u.UnitID,
				ConversionToBase: u.ConversionToBase,
				IsDefaultUnit:    u.IsDefaultUnit,
				ProductId:        input.ProductId, // Use parent ID
			})
		}

		// 3. Fix: Use FullSaveAssociations to force update of fields like isDefaultUnit
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
			Model(&product).
			Association("ProductUnits").
			Replace(&units); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 4. Reload product with units
	r.db.Preload("ProductUnits").First(&product, "id = ?", input.ProductId)

	return &product, nil
}

// Helper function to create new product unit
// func (r *ProductRepository) createNewProductUnit(productID string, u ProductUnit) error {
// 	// If ID is empty, generate a new UUID
// 	id := u.Id
// 	if id == "" {
// 		id = uuid.New().String()
// 	}

// 	newUnit := models.ProductUnit{
// 		ID:               id,
// 		ProductId:        productID,
// 		ConversionToBase: u.ConversionToBase,
// 		IsDefaultUnit:    u.IsDefaultUnit,
// 		UnitId:           u.UnitId,
// 	}

// 	return r.db.Create(&newUnit).Error
// }

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
