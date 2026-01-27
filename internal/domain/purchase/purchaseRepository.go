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
	PayOffPurchaseDebt(paymentData PaymentRequest) error
	GetPayables() (map[string]interface{}, error)
	GetAllPayables() ([]models.Purchase, error)
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

func (r *PurchaseRepository) Create(input *models.Purchase) (*models.Purchase, error) {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	// ✅ STEP 0: Use the Date from the frontend calendar
	// input.PurchaseDate is now the "source of truth"
	tranTime := input.PurchaseDate
	tranDateStr := tranTime.Format("2006-01-02")

	// 1. PAYMENT VALIDATION & DEBT INITIALIZATION
	switch input.PaymentSource {
	case "OWNER":
		input.AmountFromOwner = int64(input.GrandTotal)
		input.AmountFromCash = 0
		input.PaidAmount = int64(input.GrandTotal)
		input.PaymentStatus = "PAID"
	case "CASH":
		input.AmountFromCash = int64(input.GrandTotal)
		input.AmountFromOwner = 0
		input.PaidAmount = int64(input.GrandTotal)
		input.PaymentStatus = "PAID"
	case "DEBT":
		input.AmountFromCash = 0
		input.AmountFromOwner = 0
		input.PaidAmount = 0
		input.PaymentStatus = "PENDING"
	case "SPLIT":
		input.PaidAmount = input.AmountFromCash + input.AmountFromOwner
		if input.PaidAmount >= int64(input.GrandTotal) {
			input.PaymentStatus = "PAID"
		} else if input.PaidAmount > 0 {
			input.PaymentStatus = "PARTIAL"
		} else {
			input.PaymentStatus = "PENDING"
		}
	}
	input.BalanceAmount = int64(input.GrandTotal) - input.PaidAmount

	if err := tx.Create(input).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 2. CASHBOOK LOGIC
	var lastEntry models.Cashbook
	// Find the latest balance BEFORE this transaction date
	tx.Where("transaction_date <= ?", tranTime).Order("transaction_date desc, id desc").Limit(1).Find(&lastEntry)
	runningBalance := lastEntry.Balance

	if input.PaymentSource == "DEBT" {
		tx.Create(&models.Cashbook{
			TransactionDate: tranTime, // ✅ Use selected date
			TransactionType: "PURCHASE_DEBT",
			ReferenceID:     input.ID,
			Description:     fmt.Sprintf("Inventory Received on Credit: %s", input.ID),
			Balance:         runningBalance,
			PaymentMethod:   "DEBT",
		})
	} else {
		if input.AmountFromOwner > 0 {
			runningBalance += input.AmountFromOwner
			tx.Create(&models.Cashbook{
				TransactionDate: tranTime, // ✅ Use selected date
				TransactionType: "OWNER_INJECTION",
				ReferenceID:     input.ID,
				Description:     fmt.Sprintf("Owner Funding for Purchase %s", input.ID),
				Debit:           input.AmountFromOwner,
				Balance:         runningBalance,
			})
		}
		if input.PaidAmount > 0 {
			runningBalance -= input.PaidAmount
			tx.Create(&models.Cashbook{
				TransactionDate: tranTime, // ✅ Use selected date
				TransactionType: "PURCHASE",
				PaymentMethod:   input.PaymentSource,
				ReferenceID:     input.ID,
				Description:     fmt.Sprintf("Cash Out for Purchase %s", input.ID),
				Credit:          input.PaidAmount,
				Balance:         runningBalance,
			})
		}
	}

	// 3. DAILY SUMMARY SYNC (Based on tranDateStr, not time.Now)
	var summary models.DailySummaries
	err := tx.Where("DATE(summary_date) = ?", tranDateStr).First(&summary).Error

	if err == nil {
		tx.Model(&summary).Updates(map[string]interface{}{
			"closing_balance": runningBalance,
			"cash_total":      gorm.Expr("cash_total - ?", input.AmountFromCash),
		})
	} else {
		var lastSum models.DailySummaries
		tx.Order("summary_date desc").Limit(1).Find(&lastSum)
		openingVal := lastSum.ClosingBalance

		tx.Create(&models.DailySummaries{
			SummaryDate:    tranTime, // ✅ Use selected date
			OpeningBalance: openingVal,
			ClosingBalance: runningBalance,
			CashTotal:      -input.AmountFromCash,
			IsClosed:       false,
		})
	}

	// 4. STOCK & HISTORY
	for i := range input.PurchaseDetails {
		pd := &input.PurchaseDetails[i]
		pd.PurchaseId = input.ID

		util.AddStockMovement(tx, pd.ProductId, pd.ProductUnitId, pd.Qty, "increase")

		tx.Create(&models.ItemTransaction{
			ProductId:   pd.ProductId,
			TranType:    "PURCHASE",
			InQty:       pd.Qty,
			Uom:         pd.UnitName,
			ReferenceNo: input.ID,
			CreatedAt:   tranTime, // ✅ Use selected date
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

func (r *PurchaseRepository) GetAllPayables() ([]models.Purchase, error) {
	var purchases []models.Purchase

	// 1. We use balance_amount > 0 to find all debt.
	// 2. We use 'NOT IN' for PAID to ensure we ignore completed transactions.
	// 3. Preload("Supplier") is usually enough; clause.Associations preloads everything
	//    which can be slow, but I'll keep it if you need all sub-data.
	err := r.db.Preload(clause.Associations).
		Where("payment_status = ?", "PENDING").
		Order("purchase_date DESC").
		Find(&purchases).Error

	if err != nil {
		return nil, err
	}

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

func (r *PurchaseRepository) PayOffPurchaseDebt(paymentData PaymentRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var po models.Purchase
		// 1. Get PO and Lock for consistency
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, "id = ?", paymentData.PurchaseId).Error; err != nil {
			return err
		}

		// 2. Update PO Math & Status
		po.PaidAmount += paymentData.Amount
		po.BalanceAmount = po.GrandTotal - po.PaidAmount

		if po.BalanceAmount <= 0 {
			po.PaymentStatus = "PAID"
			po.BalanceAmount = 0
		} else {
			po.PaymentStatus = "PENDING"
		}

		if err := tx.Save(&po).Error; err != nil {
			return err
		}

		// 3. Prepare Cashbook Entry
		var lastEntry models.Cashbook
		// Get the latest drawer balance
		tx.Order("id desc").First(&lastEntry)

		cashbookEntry := models.Cashbook{
			TransactionDate: paymentData.PaymentDate,
			TransactionType: "PURCHASE_PAYMENT",
			ReferenceID:     po.ID,
			Credit:          paymentData.Amount,
			Balance:         lastEntry.Balance, // Physical drawer stays same (Owner Paid)

			// 💡 Refactor: Include the remaining balance in the description or a specific field
			// This makes it easy for Ma Zin to see the "Debt Left" in her history
			Description: fmt.Sprintf("Debt Payoff for %s. (Paid by Owner). Remaining Debt: %d",
				po.ID, po.BalanceAmount),

			PaymentMethod: "OWNER_CASH",
			CreatedAt:     time.Now(),
		}

		if err := tx.Create(&cashbookEntry).Error; err != nil {
			return err
		}

		// 4. Financial Summary Update
		summaryDate := paymentData.PaymentDate.Format("2006-01-02")
		return tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ? AND is_closed = ?", summaryDate, false).
			Update("expense_total", gorm.Expr("expense_total + ?", paymentData.Amount)).Error
	})
}

func (r *PurchaseRepository) GetPayables() (map[string]interface{}, error) {
	var summaries AgingSummary
	var supplierDebts []map[string]interface{}

	// 1. Summarized Aging Logic (PostgreSQL Syntax)
	summaryQuery := `
        SELECT 
            CASE 
                WHEN (CURRENT_DATE - purchase_date::date) <= 7 THEN '1-7 Days (Current)'
                WHEN (CURRENT_DATE - purchase_date::date) BETWEEN 8 AND 30 THEN '8-30 Days (Due)'
                ELSE '30+ Days (Overdue)'
            END as category,
            SUM(balance_amount) as total_balance,
            COUNT(id) as po_count
        FROM purchases
        WHERE balance_amount > 0
        GROUP BY category
        ORDER BY category ASC
    `
	if err := r.db.Raw(summaryQuery).Scan(&summaries).Error; err != nil {
		return nil, err
	}

	// 2. Debt by Supplier (So you know who to pay)
	supplierQuery := `
        SELECT 
            s.name as supplier_name,
            s.id as supplier_id,
            SUM(p.balance_amount) as total_owed,
            COUNT(p.id) as pending_invoices
        FROM purchases p
        JOIN suppliers s ON s.id = p.supplier_id
        WHERE p.balance_amount > 0
        GROUP BY s.name, s.id
        ORDER BY total_owed DESC
    `
	if err := r.db.Raw(supplierQuery).Scan(&supplierDebts).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"aging_summary":  summaries,
		"supplier_debts": supplierDebts,
	}, nil
}
