package purchase

import (
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

func (r *PurchaseRepository) Createold(input *models.Purchase) (*models.Purchase, error) {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	// 1. Payment Source Validation & Setup
	// Ensures AmountFromOwner/AmountFromCash are set correctly before any math
	switch input.PaymentSource {
	case "OWNER":
		input.AmountFromOwner = int64(input.GrandTotal)
		input.AmountFromCash = 0
	case "CASH":
		input.AmountFromCash = int64(input.GrandTotal)
		input.AmountFromOwner = 0
	case "SPLIT":
		if input.AmountFromCash+input.AmountFromOwner != int64(input.GrandTotal) {
			tx.Rollback()
			return nil, fmt.Errorf("split total mismatch: cash(%d) + owner(%d) != total(%d)",
				input.AmountFromCash, input.AmountFromOwner, input.GrandTotal)
		}
	}

	// Save Purchase Header
	if err := tx.Create(input).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var lastEntry models.Cashbook
	tx.Order("id desc").Limit(1).Find(&lastEntry)

	runningBalance := lastEntry.Balance
	now := time.Now() // The actual moment the drawer opens

	// ==========================================
	// 2. CASHBOOK LOGIC (Refactored for SPLIT/OWNER)
	// ==========================================

	// We ONLY record to cashbook if physical cash is being spent from the drawer
	if input.AmountFromCash > 0 {
		var lastEntry models.Cashbook
		// Get the latest balance from the drawer
		if err := tx.Order("id desc").Limit(1).Find(&lastEntry).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		runningBalance := lastEntry.Balance
		now := time.Now()

		// Deduct ONLY the portion that comes from the shop's cash
		runningBalance -= input.AmountFromCash

		cashOut := models.Cashbook{
			TransactionDate: now,
			TransactionType: "PURCHASE",
			PaymentMethod:   input.PaymentSource, // Stores "CASH" or "SPLIT"
			ReferenceID:     input.ID,
			Description:     fmt.Sprintf("Purchase %s (Drawer: %d, Owner: %d)", input.ID, input.AmountFromCash, input.AmountFromOwner),
			Credit:          input.AmountFromCash,
			Balance:         runningBalance,
			CreatedAt:       now,
		}

		if err := tx.Create(&cashOut).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		// ==========================================
		// 3. DAILY SUMMARY SYNC
		// ==========================================
		// Only update the daily summary balance if the cash drawer was touched
		todayStr := now.Format("2006-01-02")
		tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ?", todayStr).
			Update("closing_balance", runningBalance)

	} else {
		// If AmountFromCash is 0 (100% OWNER), we do nothing here.
		// The drawer balance remains exactly as it was.
		fmt.Printf("Purchase %s fully funded by Owner. No cashbook entry created.\n", input.ID)
	}

	// ==========================================
	// 3. DAILY SUMMARY SYNC (Physical Balance Sync)
	// ==========================================
	// We update TODAY'S summary because that's when the cash physically changed.
	todayStr := now.Format("2006-01-02")
	result := tx.Model(&models.DailySummaries{}).
		Where("DATE(summary_date) = ?", todayStr).
		Update("closing_balance", runningBalance)

	// If today's first transaction, create the summary row
	if result.RowsAffected == 0 {
		var lastSum models.DailySummaries
		tx.Order("summary_date desc").Limit(1).Find(&lastSum)

		openingVal := lastSum.ClosingBalance
		if openingVal == 0 {
			openingVal = 5000
		} // Default float

		if err := tx.Create(&models.DailySummaries{
			SummaryDate:    now,
			OpeningBalance: openingVal,
			ClosingBalance: runningBalance,
			IsClosed:       false,
		}).Error; err != nil {
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

		// Update Stock
		if err := util.AddStockMovement(tx, pd.ProductId, pd.ProductUnitId, pd.Qty, "increase"); err != nil {
			tx.Rollback()
			return nil, err
		}

		// Item Transaction Log
		itemTxn := models.ItemTransaction{
			ProductId:   pd.ProductId,
			TranType:    "PURCHASE",
			InQty:       pd.Qty,
			Uom:         pd.Uom,
			ReferenceNo: input.ID,
			Remark:      fmt.Sprintf("Purchased %d %s", pd.Qty, pd.Uom),
			CreatedAt:   now,
		}
		tx.Create(&itemTxn)

		// Price History (Uses input.PurchaseDate for logical history)
		history := models.ProductPriceHistory{
			ProductId:     pd.ProductId,
			UnitId:        pd.UnitId,
			UnitName:      pd.Uom,
			UnitPrice:     pd.Price,
			PriceType:     "BUY",
			Remark:        fmt.Sprintf("Purchase #%v", input.ID),
			EffectiveDate: input.PurchaseDate.Format("2006-01-02 15:04:05"),
		}
		tx.Create(&history)

		// Update Product Master Price
		r.updatePurchasePrice(tx, pd.ProductId, pd.Price, pd.ProductUnitId)
	}

	return input, tx.Commit().Error
}

func (r *PurchaseRepository) Create(input *models.Purchase) (*models.Purchase, error) {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	// 1. PAYMENT VALIDATION
	switch input.PaymentSource {
	case "OWNER":
		input.AmountFromOwner = int64(input.GrandTotal)
		input.AmountFromCash = 0
	case "CASH":
		input.AmountFromCash = int64(input.GrandTotal)
		input.AmountFromOwner = 0
	case "SPLIT":
		if input.AmountFromCash+input.AmountFromOwner != int64(input.GrandTotal) {
			tx.Rollback()
			return nil, fmt.Errorf("split total mismatch: cash(%d) + owner(%d) != total(%d)",
				input.AmountFromCash, input.AmountFromOwner, input.GrandTotal)
		}
	}

	if err := tx.Create(input).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2. CASHBOOK LOGIC
	var lastEntry models.Cashbook
	tx.Order("id desc").Limit(1).Find(&lastEntry)
	runningBalance := lastEntry.Balance
	now := time.Now()

	// A. Owner Injection (Money In)
	if input.AmountFromOwner > 0 {
		runningBalance += input.AmountFromOwner
		tx.Create(&models.Cashbook{
			TransactionDate: now,
			TransactionType: "OWNER_INJECTION",
			PaymentMethod:   "CASH",
			ReferenceID:     input.ID,
			Description:     fmt.Sprintf("Owner Funding for Purchase %s", input.ID),
			Debit:           input.AmountFromOwner,
			Balance:         runningBalance,
			CreatedAt:       now,
		})
	}

	// B. Purchase Payment (Money Out)
	runningBalance -= int64(input.GrandTotal)
	tx.Create(&models.Cashbook{
		TransactionDate: now,
		TransactionType: "PURCHASE",
		PaymentMethod:   input.PaymentSource,
		ReferenceID:     input.ID,
		Description:     fmt.Sprintf("Payment for Purchase %s (Cash: %d, Owner: %d)", input.ID, input.AmountFromCash, input.AmountFromOwner),
		Credit:          int64(input.GrandTotal),
		Balance:         runningBalance,
		CreatedAt:       now,
	})
	// 3. DAILY SUMMARY SYNC
	todayStr := now.Format("2006-01-02")
	var summary models.DailySummaries

	// Try to find today's summary first
	err := tx.Where("DATE(summary_date) = ?", todayStr).First(&summary).Error

	if err == nil {
		// CASE A: Summary exists - Update it
		// We use Expr to decrement the cash_total correctly
		tx.Model(&summary).Updates(map[string]interface{}{
			"closing_balance": runningBalance,
			"cash_total":      gorm.Expr("cash_total - ?", input.AmountFromCash),
		})
	} else {
		// CASE B: First transaction of the day - Create it
		var lastSum models.DailySummaries
		tx.Order("summary_date desc").Limit(1).Find(&lastSum)

		openingVal := lastSum.ClosingBalance
		if openingVal == 0 {
			openingVal = 5000
		}

		// Create the record for today
		tx.Create(&models.DailySummaries{
			SummaryDate:    now,
			OpeningBalance: openingVal,
			ClosingBalance: runningBalance,
			CashTotal:      -input.AmountFromCash, // Starting with this purchase deduction
			IsClosed:       false,
		})
	}

	// 4. STOCK & HISTORY
	for i := range input.PurchaseDetails {
		pd := &input.PurchaseDetails[i]
		pd.PurchaseId = input.ID

		util.AddStockMovement(tx, pd.ProductId, pd.ProductUnitId, pd.Qty, "increase")

		tx.Create(&models.ItemTransaction{
			ProductId: pd.ProductId, TranType: "PURCHASE", InQty: pd.Qty,
			Uom: pd.Uom, ReferenceNo: input.ID, CreatedAt: now,
		})

		r.updatePurchasePrice(tx, pd.ProductId, pd.Price, pd.ProductUnitId)
	}

	return input, tx.Commit().Error
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
