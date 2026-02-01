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

func (r *PurchaseRepository) Create(input *models.Purchase) (*models.Purchase, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// --- 0. DATE & INITIALIZATION ---
		tranTime := input.PurchaseDate
		tranDateStr := tranTime.Format("2006-01-02")

		// --- 1. PAYMENT VALIDATION (Core Logic) ---
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
		fmt.Println(input)
		if err := tx.Create(input).Error; err != nil {
			return err
		}

		// --- 2. SELECT DAILY SUMMARY WITH LOCK ---
		// Locking this row prevents other transactions from calculating a wrong balance simultaneously.
		var summary models.DailySummaries
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("DATE(summary_date) = ? AND is_closed = ?", tranDateStr, false).
			First(&summary).Error

		var startingBalance int64
		if err != nil {
			// No summary for today yet, fetch previous day's closing
			var lastSum models.DailySummaries
			tx.Order("summary_date desc").First(&lastSum)
			startingBalance = lastSum.ClosingBalance
			if startingBalance < 0 {
				startingBalance = 0
			}
		} else {
			startingBalance = summary.ClosingBalance
		}

		currentRunningBalance := startingBalance

		// --- 3. CASHBOOK LOGIC (Physical Drawer Impact) ---
		if input.PaymentSource == "DEBT" {
			tx.Create(&models.Cashbook{
				TransactionDate: tranTime,
				TransactionType: "PURCHASE_DEBT",
				ReferenceID:     input.ID,
				Description:     fmt.Sprintf("Inventory Received on Credit: %s", input.ID),
				Balance:         currentRunningBalance, // No change to balance
				PaymentMethod:   "DEBT",
				CreatedAt:       time.Now(),
			})
		} else {
			// A. Record Owner Injection if they provided funds
			if input.AmountFromOwner > 0 {
				currentRunningBalance += input.AmountFromOwner
				tx.Create(&models.Cashbook{
					TransactionDate: tranTime,
					TransactionType: "OWNER_INJECTION",
					ReferenceID:     input.ID,
					Description:     fmt.Sprintf("Owner Funding for Purchase %s", input.ID),
					Debit:           input.AmountFromOwner,
					Balance:         currentRunningBalance,
					CreatedAt:       time.Now(),
				})
			}
			// B. Record the Purchase (Money leaving the business)
			if input.PaidAmount > 0 {
				currentRunningBalance -= input.PaidAmount
				tx.Create(&models.Cashbook{
					TransactionDate: tranTime,
					TransactionType: "PURCHASE",
					PaymentMethod:   input.PaymentSource,
					ReferenceID:     input.ID,
					Description:     fmt.Sprintf("Cash Out for Purchase %s", input.ID),
					Credit:          input.PaidAmount,
					Balance:         currentRunningBalance,
					CreatedAt:       time.Now(),
				})
			}
		}

		// --- 4. DAILY SUMMARY SYNC ---
		if summary.ID != 0 {
			if err := tx.Model(&summary).Updates(map[string]interface{}{
				"closing_balance": currentRunningBalance, // Set absolute result
				"cash_total":      gorm.Expr("cash_total - ?", input.AmountFromCash),
			}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&models.DailySummaries{
				SummaryDate:    tranTime,
				OpeningBalance: startingBalance,
				ClosingBalance: currentRunningBalance,
				CashTotal:      -input.AmountFromCash,
				IsClosed:       false,
			}).Error; err != nil {
				return err
			}
		}

		// --- 5. STOCK & HISTORY ---
		for i := range input.PurchaseDetails {
			pd := &input.PurchaseDetails[i]
			pd.PurchaseId = input.ID
			util.AddStockMovement(tx, pd.ProductId, pd.ProductUnitId, pd.Qty, "increase")

			tx.Create(&models.ItemTransaction{
				ProductId: pd.ProductId, TranType: "PURCHASE", InQty: pd.Qty,
				Uom: pd.UnitName, ReferenceNo: input.ID, CreatedAt: tranTime,
			})

			r.updatePurchasePrice(tx, pd.ProductId, pd.Price, pd.ProductUnitId)
		}
		return nil
	})
	return input, err
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

func (r *PurchaseRepository) PayOffPurchaseDebtOLD(paymentData PaymentRequest) error {
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
			po.PaymentStatus = "PARTIAL" // Changed from PENDING to PARTIAL for accuracy
		}

		if err := tx.Save(&po).Error; err != nil {
			return err
		}

		// --- 3. ANCHOR LOGIC (Find Current Drawer State) ---
		summaryDate := paymentData.PaymentDate.Format("2006-01-02")
		var summary models.DailySummaries

		err := tx.Where("DATE(summary_date) = ? AND is_closed = ?", summaryDate, false).First(&summary).Error

		var startingBalance int64
		if err != nil {
			// First transaction of the day - look back
			var lastSum models.DailySummaries
			tx.Order("summary_date desc").First(&lastSum)
			startingBalance = lastSum.ClosingBalance
			if startingBalance < 5000 {
				startingBalance = 5000
			}
		} else {
			startingBalance = summary.ClosingBalance
		}

		// --- 4. CASHBOOK ENTRY ---
		// Since it's OWNER_CASH, the drawer balance doesn't actually decrease.
		// We record the transaction so Ma Zin sees the history, but Balance stays at startingBalance.
		cashbookEntry := models.Cashbook{
			TransactionDate: paymentData.PaymentDate,
			TransactionType: "PURCHASE_PAYMENT",
			ReferenceID:     po.ID,
			Description:     fmt.Sprintf("Debt Payoff for %s. (Paid by Owner). Remaining Debt: %d", po.ID, po.BalanceAmount),
			Debit:           0,
			Credit:          paymentData.Amount,
			// Since PaymentMethod is OWNER_CASH, the money didn't come from the drawer.
			// Balance remains exactly what it was before this payment.
			Balance:       startingBalance,
			PaymentMethod: "OWNER_CASH",
			CreatedAt:     time.Now(),
		}

		if err := tx.Create(&cashbookEntry).Error; err != nil {
			return err
		}

		// --- 5. DAILY SUMMARY UPDATE ---
		// We update expense_total so the profit/loss is correct.
		// But we DO NOT touch closing_balance or cash_total because it was Owner Cash.
		if summary.ID != 0 {
			return tx.Model(&summary).Update("expense_total", gorm.Expr("expense_total + ?", paymentData.Amount)).Error
		} else {
			return tx.Create(&models.DailySummaries{
				SummaryDate:    paymentData.PaymentDate,
				OpeningBalance: startingBalance,
				ClosingBalance: startingBalance, // No change to drawer
				ExpenseTotal:   paymentData.Amount,
				IsClosed:       false,
			}).Error
		}
	})
}

func (r *PurchaseRepository) PayOffPurchaseDebt(paymentData PaymentRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var po models.Purchase

		// 1. GET PO AND LOCK for consistency (prevents double payments)
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, "id = ?", paymentData.PurchaseId).Error; err != nil {
			return err
		}

		// 2. UPDATE PURCHASE MATH & STATUS
		po.PaidAmount += paymentData.Amount
		po.BalanceAmount = po.GrandTotal - po.PaidAmount

		if po.BalanceAmount <= 0 {
			po.PaymentStatus = "PAID"
			po.BalanceAmount = 0
		} else {
			po.PaymentStatus = "PARTIAL"
		}

		if err := tx.Save(&po).Error; err != nil {
			return err
		}

		// --- 3. ANCHOR LOGIC WITH LOCK ---
		summaryDate := paymentData.PaymentDate.Format("2006-01-02")
		var summary models.DailySummaries

		// Lock the summary to ensure the drawer anchor is stable
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("DATE(summary_date) = ? AND is_closed = ?", summaryDate, false).
			First(&summary).Error

		var startingBalance int64
		if err != nil {
			// First transaction of the day - look back at previous closing
			var lastSum models.DailySummaries
			tx.Order("summary_date desc").First(&lastSum)
			startingBalance = lastSum.ClosingBalance
			if startingBalance < 0 {
				startingBalance = 0
			}
		} else {
			startingBalance = summary.ClosingBalance
		}

		// --- 4. CASHBOOK ENTRY ---
		// Business Logic: If Paid by OWNER_CASH, it counts as a personal injection
		// that is immediately spent. Therefore, Drawer Balance stays SAME.
		cashbookEntry := models.Cashbook{
			TransactionDate:   paymentData.PaymentDate,
			TransactionType:   "PURCHASE_PAYMENT",
			ReferenceID:       po.ID,
			Description:       fmt.Sprintf("Debt Payoff for %s (Owner Cash). Rem: %d", po.ID, po.BalanceAmount),
			Debit:             0,
			Credit:            paymentData.Amount,
			Balance:           startingBalance, // Balance does not decrease because drawer cash wasn't used
			PaymentMethod:     "OWNER_CASH",
			TransactionStatus: po.PaymentStatus,
			CreatedAt:         time.Now(),
		}

		if err := tx.Create(&cashbookEntry).Error; err != nil {
			return err
		}

		// --- 5. DAILY SUMMARY UPDATE ---
		// We track the expense, but 'closing_balance' remains 'startingBalance'.
		if summary.ID != 0 {
			return tx.Model(&summary).Updates(map[string]interface{}{
				"closing_balance": startingBalance, // Explicitly keep balance locked to starting point
				"expense_total":   gorm.Expr("expense_total + ?", paymentData.Amount),
			}).Error
		} else {
			return tx.Create(&models.DailySummaries{
				SummaryDate:    paymentData.PaymentDate,
				OpeningBalance: startingBalance,
				ClosingBalance: startingBalance,
				ExpenseTotal:   paymentData.Amount,
				IsClosed:       false,
			}).Error
		}
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
