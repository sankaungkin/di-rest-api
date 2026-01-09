package purchase

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
	GetHistoricalMonthlyCOGS() ([]ResponseHistoricalCOGS, error)
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

// In purchase/repository.go (Add this method)

func (r *PurchaseRepository) GetHistoricalMonthlyCOGS() ([]ResponseHistoricalCOGS, error) {
	var results []ResponseHistoricalCOGS

	// Calculate the start date (5 months ago, starting from the 1st day of that month)
	now := time.Now()
	// Go back 4 full months + current month = 5 months total
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -4, 0)

	// SQL query to aggregate monthly purchase grand totals (COGS)
	query := `
        SELECT
            TO_CHAR(DATE_TRUNC('month', purchase_date), 'YYYY-MM') AS month,
            COALESCE(SUM(grand_total), 0) AS cogs
        FROM public.purchases
        WHERE purchase_date >= ?
        GROUP BY 1
        ORDER BY 1 ASC
    `

	if err := r.db.Raw(query, startDate).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
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
    pp.unit_price,
	ps.derived_qty as stock_qty
FROM product_units pu
JOIN product_prices pp 
    ON pu.product_id = pp.product_id 
    AND pu.unit_id = pp.unit_id
JOIN unit_of_measures uom         
    ON uom.id = pu.unit_id
JOIN products p
	ON p.id = pu.product_id
JOIN product_stocks ps
	ON ps.product_id = pu.product_id
WHERE 
    pp.price_type = 'BUY'
	`

	if err := r.db.Raw(query).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *PurchaseRepository) CreateOld(input *models.Purchase) (*models.Purchase, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Initial Validation & Creation
	if verr := models.ValidateStruct(input); verr != nil {
		tx.Rollback()
		return nil, fmt.Errorf("validation failed: %v", verr)
	}

	if err := tx.Create(input).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ==========================================
	// 2. CASHBOOK LOGIC (The New Requirement)
	// ==========================================
	var lastEntry models.Cashbook
	tx.Order("id desc").Last(&lastEntry)
	currentBalance := lastEntry.Balance
	purchaseAmount := int64(input.GrandTotal)

	// Check if owner needs to inject funds
	if currentBalance < purchaseAmount {
		injectionAmount := purchaseAmount - currentBalance
		injection := models.Cashbook{
			TransactionDate: input.PurchaseDate,
			TransactionType: "OWNER_INJECTION",
			ReferenceID:     input.ID,
			Description:     fmt.Sprintf("Auto-injection for Purchase %s", input.ID),
			Debit:           injectionAmount,
			Credit:          0,
			Balance:         currentBalance + injectionAmount, // Brings balance up to exactly purchase price
			CreatedAt:       time.Now(),
		}
		if err := tx.Create(&injection).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed cash injection: %v", err)
		}
		currentBalance = injection.Balance
	}

	// Record the Purchase payment in Cashbook
	cashOut := models.Cashbook{
		TransactionDate: input.PurchaseDate,
		TransactionType: "PURCHASE",
		ReferenceID:     input.ID,
		Description:     fmt.Sprintf("Payment for Purchase %s", input.ID),
		Debit:           0,
		Credit:          purchaseAmount,
		Balance:         currentBalance - purchaseAmount,
		CreatedAt:       time.Now(),
	}
	if err := tx.Create(&cashOut).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed cashbook payment: %v", err)
	}

	// ==========================================
	// 3. STOCK & HISTORY PROCESSING
	// ==========================================
	for i := range input.PurchaseDetails {
		pd := &input.PurchaseDetails[i]

		// Load relations for history/logging
		if err := tx.Preload("Product").Preload("ProductUnit").First(pd, pd.ID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// A. Stock Movement
		if err := util.AddStockMovement(tx, pd.ProductId, pd.ProductUnitId, pd.Qty, "increase"); err != nil {
			tx.Rollback()
			return nil, err
		}

		// B. Item Transaction Log
		itemTxn := models.ItemTransaction{
			ProductId:   pd.ProductId,
			TranType:    "PURCHASE",
			InQty:       pd.Qty,
			Uom:         pd.Uom,
			ReferenceNo: fmt.Sprintf("%s-%d", input.ID, pd.ID),
			Remark:      fmt.Sprintf("Purchased %d %s of %s", pd.Qty, pd.Uom, pd.Product.ProductName),
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&itemTxn).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// C. Price History
		history := models.ProductPriceHistory{
			ProductId:     pd.ProductId,
			ProductName:   pd.Product.ProductName,
			UnitId:        pd.UnitId,
			UnitName:      pd.Uom,
			UnitPrice:     pd.Price,
			PriceType:     "BUY",
			Remark:        fmt.Sprintf("Purchase ID: %s", input.ID),
			EffectiveDate: input.PurchaseDate.Format("2006-01-02 15:04:05"),
		}
		if err := tx.Create(&history).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// D. Update Master Product Price
		if err := r.updatePurchasePrice(tx, pd.ProductId, pd.Price, pd.ProductUnitId); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return input, nil
}

func (r *PurchaseRepository) Create(input *models.Purchase) (*models.Purchase, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Initial Validation & Creation
	if verr := models.ValidateStruct(input); verr != nil {
		tx.Rollback()
		return nil, fmt.Errorf("validation failed: %v", verr)
	}

	if err := tx.Create(input).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ==========================================
	// 2. CASHBOOK LOGIC (Balance Check & Injection)
	// ==========================================
	var lastEntry models.Cashbook
	err := tx.Order("id desc").First(&lastEntry).Error

	currentBalance := int64(0)
	if err == nil {
		currentBalance = lastEntry.Balance
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	purchaseAmount := int64(input.GrandTotal)

	// Check if owner needs to inject funds
	if currentBalance < purchaseAmount {
		injectionAmount := purchaseAmount - currentBalance
		injection := models.Cashbook{
			TransactionDate: input.PurchaseDate,
			TransactionType: "OWNER_INJECTION",
			ReferenceID:     fmt.Sprint(input.ID),
			Description:     fmt.Sprintf("Auto-injection for Purchase #%v", input.ID),
			Debit:           injectionAmount,
			Credit:          0,
			Balance:         currentBalance + injectionAmount,
			CreatedAt:       time.Now(),
		}
		if err := tx.Create(&injection).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed cash injection: %v", err)
		}
		currentBalance = injection.Balance
	}

	// Record the Purchase payment in Cashbook
	newBalance := currentBalance - purchaseAmount
	cashOut := models.Cashbook{
		TransactionDate: input.PurchaseDate,
		TransactionType: "PURCHASE",
		ReferenceID:     fmt.Sprint(input.ID),
		Description:     fmt.Sprintf("Payment for Purchase #%v", input.ID),
		Debit:           0,
		Credit:          purchaseAmount,
		Balance:         newBalance,
		CreatedAt:       time.Now(),
	}
	if err := tx.Create(&cashOut).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed cashbook payment: %v", err)
	}

	// ==========================================
	// 3. DAILY SUMMARY SYNC
	// ==========================================
	todayStr := input.PurchaseDate.Format("2006-01-02")

	// Update existing summary
	result := tx.Model(&models.DailySummaries{}).
		Where("DATE(summary_date) = ?", todayStr).
		Update("closing_balance", newBalance)

	if result.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update daily summary: %v", result.Error)
	}

	// Auto-create summary if missing (e.g., system reset or first entry of the day)
	if result.RowsAffected == 0 {
		newSummary := models.DailySummaries{
			SummaryDate:    input.PurchaseDate,
			OpeningBalance: (newBalance + purchaseAmount), // State before this specific purchase
			ClosingBalance: newBalance,
			IsClosed:       false,
		}
		if err := tx.Create(&newSummary).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// ==========================================
	// 4. STOCK & HISTORY PROCESSING
	// ==========================================
	for i := range input.PurchaseDetails {
		pd := &input.PurchaseDetails[i]
		pd.PurchaseId = input.ID

		var product models.Product
		if err := tx.Where("id = ?", pd.ProductId).First(&product).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := util.AddStockMovement(tx, pd.ProductId, pd.ProductUnitId, pd.Qty, "increase"); err != nil {
			tx.Rollback()
			return nil, err
		}

		itemTxn := models.ItemTransaction{
			ProductId:   pd.ProductId,
			TranType:    "PURCHASE",
			InQty:       pd.Qty,
			Uom:         pd.Uom,
			ReferenceNo: fmt.Sprint(input.ID),
			Remark:      fmt.Sprintf("Purchased %d %s", pd.Qty, pd.Uom),
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&itemTxn).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		history := models.ProductPriceHistory{
			ProductId:     pd.ProductId,
			ProductName:   product.ProductName,
			UnitId:        pd.UnitId,
			UnitName:      pd.Uom,
			UnitPrice:     pd.Price,
			PriceType:     "BUY",
			Remark:        fmt.Sprintf("Purchase #%v", input.ID),
			EffectiveDate: input.PurchaseDate.Format("2006-01-02 15:04:05"),
		}
		if err := tx.Create(&history).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := r.updatePurchasePrice(tx, pd.ProductId, pd.Price, pd.ProductUnitId); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
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
	if err := tx.Where("product_id = ? AND product_unit_id = ? AND price_type = 'BUY'", productId, productUnitId).First(&productPrice).Error; err != nil {
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

func (r *PurchaseRepository) getLatestBalance(tx *gorm.DB) int64 {
	var lastEntry models.Cashbook
	tx.Last(&lastEntry)
	return lastEntry.Balance
}
