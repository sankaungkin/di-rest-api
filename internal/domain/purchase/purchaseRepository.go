package purchase

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PurchaseRepositoryInterface interface {
	Create(sale *models.Purchase) (*models.Purchase, error)
	GetAll() ([]models.Purchase, error)
	GetTodayPurchaseList() ([]models.Purchase, error)
	GetPurchasesByDate(date time.Time) ([]models.Purchase, error)
	GetTodayPurchases() ([]models.Purchase, error)
	GetById(id string) (*models.Purchase, error)
	GetTodayGrandTotal() (int64, error)
	GetMonthlyPurchases() ([]models.Purchase, error)
	GetMonthlyGrandTotal() (int64, error)
	UpdatePurchaseRemark(purchaseRemark UpdateRemarkPurchaseDTO) (*models.Purchase, error)
	GetPurchaseLineItems() ([]ResponsePurchaseLineItemDTO, error)
}

type PurchaseRepository struct {
	db *gorm.DB
}

var (
	repoInstance *PurchaseRepository
	repoOnce     sync.Once
)

func NewSaleRepository(db *gorm.DB) PurchaseRepositoryInterface {
	log.Println(util.Magenta + "SaleRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &PurchaseRepository{db: db}
	})
	return repoInstance
}

func (r *PurchaseRepository) GetPurchaseLineItems() ([]ResponsePurchaseLineItemDTO, error) {
	var result []ResponsePurchaseLineItemDTO

	query := `
		SELECT 
    pu.id as product_unit_id,
	p.product_name,
    pu.product_id,
    pu.unit_id as unit_id,
    uom.unit_name as unit_name,           
    pp.unit_id as price_unit_id,
    pp.price_type,
    pp.unit_price
FROM product_units pu
JOIN product_prices pp 
    ON pu.product_id = pp.product_id 
    AND pu.unit_id = pp.unit_id
JOIN unit_of_measures uom         
    ON uom.id = pu.unit_id
JOIN products p
	ON p.id = pu.product_id
WHERE 
    pp.price_type = 'BUY'
	`

	if err := r.db.Raw(query).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PurchaseRepository) CreateOld(input *models.Purchase) (*models.Purchase, error) {

	newPurchase := models.Purchase{
		ID:              input.ID,
		SupplierId:      input.SupplierId,
		Discount:        input.Discount,
		GrandTotal:      input.GrandTotal,
		Remark:          input.Remark,
		PurchaseDate:    input.PurchaseDate,
		PurchaseDetails: input.PurchaseDetails,
		Total:           input.Total,
	}
	err := models.ValidateStruct(newPurchase)
	if err != nil {
		return nil, gorm.ErrCheckConstraintViolated
	}

	tx := r.db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	if err := tx.Create(&newPurchase).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Preload("PurchaseDetails.Product").First(&newPurchase, "id = ?", newPurchase.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range newPurchase.PurchaseDetails {

		// increase productStock qtyonhand
		var productStock models.ProductStock
		result := tx.First(&productStock, "product_id = ?", newPurchase.PurchaseDetails[i].ProductId)
		if err := result.Error; err != nil {
			return nil, err
		}
		productStock.BaseQty += int(newPurchase.PurchaseDetails[i].Qty)
		tx.Save(&productStock)

		newItemTransaction := models.ItemTransaction{
			InQty:       newPurchase.PurchaseDetails[i].Qty,
			OutQty:      0,
			ProductId:   newPurchase.PurchaseDetails[i].ProductId,
			TranType:    "PURCHASE",
			ReferenceNo: newPurchase.ID + "-" + strconv.Itoa(int(newPurchase.PurchaseDetails[i].ID)),
			Uom:         newPurchase.PurchaseDetails[i].UnitName,
			Remark: fmt.Sprintf(
				"PurchaseID:%s, purchaseDetail id:%d, buy %s %s %d %s ",
				newPurchase.ID, newPurchase.PurchaseDetails[i].ID, newPurchase.PurchaseDetails[i].ProductId, newPurchase.PurchaseDetails[i].Product.ProductName, newPurchase.PurchaseDetails[i].Qty, newPurchase.PurchaseDetails[i].UnitName,
			),
		}
		tx.Save(&newItemTransaction)

		newProductPrice := models.ProductPrice{
			ProductId: newPurchase.PurchaseDetails[i].ProductId,
			UnitId:    newPurchase.PurchaseDetails[i].UnitId,
			UnitPrice: newPurchase.PurchaseDetails[i].Price,
			PriceType: "BUY",
			Remark:    "PurchaseID:" + newPurchase.ID + ", line item id:" + strconv.Itoa(int(newPurchase.PurchaseDetails[i].ID)) + ", increase quantity: " + strconv.Itoa(newPurchase.PurchaseDetails[i].Qty) + " " + newPurchase.PurchaseDetails[i].UnitName,
		}

		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "product_id"}, {Name: "unit_id"}, {Name: "price_type"}},
			DoUpdates: clause.AssignmentColumns([]string{"unit_price", "updated_at"}),
		}).Create(&newProductPrice).Error

		if err != nil {
			return nil, err
		}

		//save new record to ProductPriceHistory
		newProductPriceHistory := models.ProductPriceHistory{
			ProductId:     newPurchase.PurchaseDetails[i].ProductId,
			ProductName:   newPurchase.PurchaseDetails[i].ProductName,
			UnitName:      newPurchase.PurchaseDetails[i].UnitName,
			UnitId:        newPurchase.PurchaseDetails[i].UnitId,
			UnitPrice:     newPurchase.PurchaseDetails[i].Price,
			PriceType:     "BUY",
			EffectiveDate: newPurchase.PurchaseDate.Format("2006-01-02 15:04:05"),
		}
		tx.Save(&newProductPriceHistory)

	}
	tx.Commit()
	return &newPurchase, nil

}

func (r *PurchaseRepository) Create(input *models.Purchase) (*models.Purchase, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Precompute BaseQty for each detail before saving
		for i := range input.PurchaseDetails {
			detail := &input.PurchaseDetails[i]

			// Find ProductUnit
			var pu models.ProductUnit
			if err := tx.Where("id = ?", detail.ProductUnitId).First(&pu).Error; err != nil {
				return fmt.Errorf("product unit not found: %w", err)
			}

			detail.BaseQty = detail.Qty * pu.ConversionToBase
		}

		// 2. Save Purchase + PurchaseDetails (so IDs are available)
		if err := tx.Create(&input).Error; err != nil {
			return err
		}
		// 3. Create ItemTransactions for each detail
		for _, detail := range input.PurchaseDetails {
			// Load product to get ProductName
			var product models.Product
			if err := tx.Where("id = ?", detail.ProductId).First(&product).Error; err != nil {
				return fmt.Errorf("product not found: %w", err)
			}

			// 5. Create ItemTransactions
			newItemTransaction := models.ItemTransaction{
				ProductId:   detail.ProductId,
				InQty:       detail.Qty,
				TranType:    "PURCHASE",
				ReferenceNo: fmt.Sprintf("%s-%d", input.ID, detail.ID),
				Uom:         detail.UnitName,
				Remark: fmt.Sprintf("PurchaseID:%s, PurchaseDetail id:%d, Buy %s : %s : %d %s",
					input.ID, detail.ID, detail.ProductId, product.ProductName, detail.Qty, detail.UnitName),
				CreatedAt: time.Now().Local(),
			}
			if err := tx.Create(&newItemTransaction).Error; err != nil {
				return err
			}

			// 6. Create ProductPriceHistory
			newProductPriceHistory := models.ProductPriceHistory{
				ProductId:   detail.ProductId,
				ProductName: product.ProductName,
				UnitName:    detail.UnitName,
				UnitId:      detail.UnitId,
				UnitPrice:   detail.Price,
				PriceType:   "BUY",
				Remark: fmt.Sprintf("PurchaseID:%s, PurchaseDetail id:%d, Buy %s : %s : %d %s",
					input.ID, detail.ID, detail.ProductId, product.ProductName, detail.Qty, detail.UnitName),
				EffectiveDate: input.PurchaseDate.Format("2006-01-02 15:04:05"),
			}

			if err := tx.Create(&newProductPriceHistory).Error; err != nil {
				return err
			}

		}

		// 4. Update stock
		for _, detail := range input.PurchaseDetails {
			if err := r.updateStock(tx, detail.ProductId, detail.BaseQty); err != nil {
				return err
			}
		}

		// 5. Update purchase price
		for _, detail := range input.PurchaseDetails {
			if err := r.updatePurchasePrice(tx, detail.ProductId, detail.Price, detail.ProductUnitId); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return input, nil
}

// Update stock always in base unit
func (r *PurchaseRepository) updateStock(tx *gorm.DB, productId string, qty int) error {
	var stock models.ProductStock
	if err := tx.Where("product_id = ?", productId).First(&stock).Error; err != nil {
		// If not exists, create new
		if errors.Is(err, gorm.ErrRecordNotFound) {
			stock = models.ProductStock{
				ProductId:  productId,
				DerivedQty: qty,
			}
			return tx.Create(&stock).Error
		}
		return err
	}

	// Increase stock
	stock.DerivedQty += qty
	return tx.Save(&stock).Error
}

func (r *PurchaseRepository) updatePurchasePrice(tx *gorm.DB, productId string, unitPrice int, productUnitId string) error {
	var newPrice models.ProductPrice
	if err := tx.Where("product_id = ? AND price_type = 'BUY'", productId).First(&newPrice).Error; err != nil {
		return err
	}

	// Find the productPrice with the same productId and unitId
	var productPrice models.ProductPrice
	if err := tx.Where("product_id = ? AND product_unit_id = ?", productId, productUnitId).First(&productPrice).Error; err != nil {
		return err
	}

	// Update the productPrice
	productPrice.UnitPrice = unitPrice
	return tx.Save(&productPrice).Error

}
func (r *PurchaseRepository) GetAll() ([]models.Purchase, error) {

	purchases := []models.Purchase{}
	results := r.db.Preload(clause.Associations).Model(&models.Purchase{}).Order("created_at DESC").Find(&purchases)

	if results.Error != nil {
		return nil, results.Error
	}

	// if len(purchases) == 0 {
	// 	return nil, errors.New("NO records found")
	// }

	return purchases, nil
}

func (r *PurchaseRepository) UpdatePurchaseRemark(purchaseRemark UpdateRemarkPurchaseDTO) (*models.Purchase, error) {
	var existingPurchase models.Purchase
	err := r.db.Where("id = ?", purchaseRemark.ID).First(&existingPurchase).Error
	if err != nil {
		return nil, err
	}

	existingPurchase.Remark = purchaseRemark.Remark

	log.Println("existingPurchase to update: ", existingPurchase)
	err = r.db.Save(&existingPurchase).Error
	if err != nil {
		return nil, err
	}

	return &existingPurchase, nil
}

func (r *PurchaseRepository) GetTodayPurchases() ([]models.Purchase, error) {
	var purchases []models.Purchase

	// today := time.Now().Format("2006-01-02") // e.g., "2025-07-11"

	today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start.Add(24 * time.Hour)

	result := r.db.
		Preload(clause.Associations).
		Where("purchase_date >= ? AND purchase_date < ?", start, end).
		// Where("sale_date = ?", today).
		Order("purchase_date DESC").
		Find(&purchases)

	if result.Error != nil {
		return nil, result.Error
	}

	// if len(purchases) == 0 {
	// 	return nil, errors.New("NO records found for today")
	// }

	return purchases, nil
}

func (r *PurchaseRepository) GetById(id string) (*models.Purchase, error) {

	var purchase models.Purchase
	err := r.db.
		Preload("Supplier").
		Preload("PurchaseDetails").
		First(&purchase, "id = ?", strings.ToUpper(id)).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	return &purchase, nil
}

func (r *PurchaseRepository) GetTodayGrandTotal() (int64, error) {

	var total int64

	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Model(&models.Purchase{}).
		Select("COALESCE(SUM(grand_total), 0)").
		Where("purchase_date >= ? AND purchase_date < ?", startOfDay, endOfDay).
		Scan(&total).Error

	if err != nil {
		return 0, err
	}
	return total, nil

}

func (r *PurchaseRepository) GetMonthlyPurchases() ([]models.Purchase, error) {
	var purchases []models.Purchase

	// Get first day of current month
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthStartStr := monthStart.Format("2006-01-02")

	// Get first day of next month
	nextMonth := monthStart.AddDate(0, 1, 0)
	nextMonthStr := nextMonth.Format("2006-01-02")

	// Query sales within this month range
	err := r.db.
		Preload(clause.Associations).
		Where("purchase_date >= ? AND purchase_date < ?", monthStartStr, nextMonthStr).
		Order("purchase_date DESC").
		Find(&purchases).Error

	if err != nil {
		return nil, err
	}
	// if len(purchases) == 0 {
	// 	return nil, errors.New("NO records found for this month")
	// }

	return purchases, nil
}

func (r *PurchaseRepository) GetMonthlyGrandTotal() (int64, error) {
	var total int64

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)

	err := r.db.Model(&models.Purchase{}).
		Select("COALESCE(SUM(grand_total), 0)").
		Where("purchase_date >= ? AND purchase_date < ?", monthStart.Format("2006-01-02"), nextMonth.Format("2006-01-02")).
		Scan(&total).Error

	return total, err
}

func (r *PurchaseRepository) GetTodayPurchaseList() ([]models.Purchase, error) {
	var purchases []models.Purchase

	// today := time.Now().Format("2006-01-02") // e.g., "2025-07-11"

	loc, _ := time.LoadLocation("Asia/Yangon")
	today := time.Now().In(loc)
	// today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	end := start.Add(24 * time.Hour)

	// Convert Yangon times to UTC for database query
	startUTC := start.UTC()
	endUTC := end.UTC()

	fmt.Println("start:", start)
	fmt.Println("end:", end)

	result := r.db.
		Preload(clause.Associations).
		Where("purchase_date >= ? AND purchase_date < ?", startUTC, endUTC).
		// Where("sale_date = ?", today).
		Order("purchase_date DESC").
		Find(&purchases)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.Error != nil {
		return nil, result.Error
	}
	// if len(purchases) == 0 {
	// 	return nil, errors.New("NO records found for today")
	// }

	return purchases, nil
}

func (r *PurchaseRepository) GetPurchasesByDate(date time.Time) ([]models.Purchase, error) {
	purchases := []models.Purchase{}

	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	result := r.db.
		Preload(clause.Associations).
		Where("purchase_date >= ? AND purchase_date < ?", startOfDay, endOfDay).
		Order("purchase_date DESC").
		Find(&purchases)

	if result.Error != nil {
		return nil, result.Error
	}

	return purchases, nil
}
