package sale

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/cashbook"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaleRepositoryInterface interface {
	Create(sale *models.Sale) (*models.Sale, error)
	GetAll() ([]models.Sale, error)
	GetSalesWithReceivables() ([]models.Sale, error)
	GetDailySales() ([]ResponseDailySalesDTO, error)
	GetTodaySales() ([]models.Sale, error)
	GetSalesByDate(date time.Time) ([]models.Sale, error)
	GetTodaySaleList() ([]models.Sale, error)
	GetTopTenSoleProducts() ([]ResponseTopTenSoleProductsDTO, error)
	GetById(id string) (*models.Sale, error)
	GetTodayGrandTotal() (int64, error)
	GetMonthlySales() ([]models.Sale, error)
	GetMonthlyGrandTotal() (int64, error)
	TopCustomers() (*ResponseTopCustomerDTO, error)
	UpdateSale(sale UpdateSaleRemarkDTO) (*models.Sale, error)

	GetHistoricalProfitData() ([]ResponseMonthlyProfitDataDTO, error)
	GetSaleStockItemWithPrice() ([]ResponseSaleStockItemWithPrice, error)
	ReturnSaleItems(returnItem SaleReturnDTO) (*models.SaleReturn, error)

	CollectDebt(payment *models.PaymentRecord, saleID string) error
	GetPaymentHistory(saleID string) ([]models.PaymentRecord, error)
}

type SaleRepository struct {
	db       *gorm.DB
	cashRepo cashbook.CashbookRepositoryInterface
}

var (
	repoInstance *SaleRepository
	repoOnce     sync.Once
)

func NewSaleRepository(db *gorm.DB, cashRepo cashbook.CashbookRepositoryInterface) SaleRepositoryInterface {
	log.Println(util.Blue + "SaleRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &SaleRepository{db: db, cashRepo: cashRepo}
	})
	return repoInstance
}

// In sale/repository.go (Add this method)
func (r *SaleRepository) GetHistoricalProfitData() ([]ResponseMonthlyProfitDataDTO, error) {
	var results []ResponseMonthlyProfitDataDTO

	// ... (Use the UTC-safe startDate calculation from the last response) ...
	now := time.Now().UTC()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -4, 0)

	// ... (Use the final, corrected SQL query with UNION ALL) ...
	query := `
        WITH MonthlyData AS (
            SELECT DATE_TRUNC('month', sale_date) AS month_key, SUM(grand_total) AS revenue, 0 AS cogs
            FROM public.sales WHERE sale_date >= ? GROUP BY 1
            UNION ALL
            SELECT DATE_TRUNC('month', purchase_date) AS month_key, 0 AS revenue, SUM(grand_total) AS cogs
            FROM public.purchases WHERE purchase_date >= ? GROUP BY 1
        )
        SELECT
            TO_CHAR(month_key, 'YYYY-MM') AS month,
            SUM(revenue) AS revenue,
            SUM(cogs) AS cogs
        FROM MonthlyData
        GROUP BY month_key
        ORDER BY month_key ASC
    `

	// ⬇️ THIS LINE IS THE MOST LIKELY CAUSE OF THE "syntax error at or near raw" ⬇️
	// Ensure this is exactly correct.
	if err := r.db.Raw(query, startDate, startDate).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	return results, nil
}
func (r *SaleRepository) GetSaleStockItemWithPrice() ([]ResponseSaleStockItemWithPrice, error) {
	var result []ResponseSaleStockItemWithPrice

	query := `
		SELECT 
		pp.product_unit_id, 
		p.product_name, 
		pp.product_id, 
		uom.id, 
		uom.unit_name, 
		pp.price_type, 
		pp.unit_price,
		ps.derived_qty as quantity_on_hand
FROM public.product_prices pp, public.products p, public.unit_of_measures uom, public.product_stocks ps
WHERE pp.product_id = p.id AND pp.unit_id = uom.id AND pp.price_type = 'SELL' AND ps.product_id = p.id AND ps.derived_qty > 0
ORDER BY pp.product_id ASC 
	`

	if err := r.db.Raw(query).Scan(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SaleRepository) Create(input *models.Sale) (*models.Sale, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// --- 1. CORE FINANCIAL LOGIC ---
		input.NetTotal = input.GrandTotal
		input.ReturnAmount = 0
		input.Status = "NORMAL"
		input.BalanceAmount = input.GrandTotal - input.PaidAmount

		// Determine Payment Status
		if input.PaidAmount <= 0 {
			input.PaymentStatus = "UNPAID"
			input.PaidAmount = 0
		} else if input.PaidAmount < input.GrandTotal {
			input.PaymentStatus = "PARTIAL"
		} else {
			input.PaymentStatus = "PAID"
			input.BalanceAmount = 0
		}

		// Save the Sale record first to generate the ID
		if err := tx.Create(input).Error; err != nil {
			return err
		}

		// --- 2. INTEGRATE WITH CASHBOOK ---
		// We call CreateEntry regardless of the payment method to keep an audit log.
		// Your CreateEntry function already handles the logic to only change
		// the 'Balance' if PaymentMethod == "CASH".

		statusText := input.PaymentStatus
		if input.PaymentStatus == "PAID" {
			statusText = "FULLY PAID"
		}

		cashEntry := &models.Cashbook{
			TransactionDate:   input.SaleDate,
			TransactionType:   "SALE",
			ReferenceID:       input.ID,
			Description:       fmt.Sprintf("Sale #%v (%s)", input.ID, statusText),
			Debit:             input.PaidAmount, // If KPay or Unpaid Debt, this is still the 'amount received'
			Credit:            0,
			PaymentMethod:     input.PaymentMethod,
			TransactionStatus: statusText,
		}

		// Delegate all Balance Math and DailySummary updates to the specialized function
		// This ensures the 5000 safety floor and running balance are handled in ONE place.
		if err := r.cashRepo.CreateEntry(tx, cashEntry); err != nil {
			return fmt.Errorf("cashbook integration failed: %v", err)
		}

		// --- 3. INVENTORY & CUSTOMER STATS ---
		for i := range input.SaleDetails {
			sd := &input.SaleDetails[i]
			sd.SaleId = input.ID

			// Decrease stock based on units sold
			if err := util.AddStockMovement(tx, sd.ProductId, sd.ProductUnitId, sd.Qty, "decrease"); err != nil {
				return err
			}

			// Log the item transaction for inventory history
			if err := tx.Create(&models.ItemTransaction{
				ProductId:   sd.ProductId,
				TranType:    "SALE",
				OutQty:      sd.Qty,
				Uom:         sd.Uom,
				ReferenceNo: input.ID,
				CreatedAt:   time.Now(),
			}).Error; err != nil {
				return err
			}
		}

		// Update customer total spent and debt flag
		return util.AdjustCustomerStats(tx, input.CustomerId, input.GrandTotal, input.BalanceAmount > 0)
	})

	return input, err
}

func (r *SaleRepository) CollectDebt(payment *models.PaymentRecord, saleID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		payment.ReferenceID = saleID
		payment.Type = "DEBT_COLLECTION"

		// 1. RECORD THE PAYMENT HISTORY
		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment record: %v", err)
		}

		// 2. UPDATE SALE RECORD (With Lock to prevent overpayment)
		var sale models.Sale
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sale, "id = ?", saleID).Error; err != nil {
			return err
		}

		newPaidAmount := sale.PaidAmount + payment.Amount
		newBalanceAmount := sale.GrandTotal - newPaidAmount

		newStatus := "PARTIAL"
		if newPaidAmount >= sale.GrandTotal {
			newStatus = "PAID"
			newBalanceAmount = 0
		}

		if err := tx.Model(&sale).Updates(map[string]interface{}{
			"paid_amount":    newPaidAmount,
			"balance_amount": newBalanceAmount,
			"payment_status": newStatus,
		}).Error; err != nil {
			return err
		}

		// 3. INTEGRATE WITH SPECIALIZED CASHBOOK REPO
		// Delegate all Balance Math and DailySummary updates to CreateEntry.
		pMethod := strings.ToUpper(payment.PaymentMethod)

		cashEntry := &models.Cashbook{
			TransactionDate:   payment.PaymentDate,
			TransactionType:   "DEBT_PAYMENT",
			ReferenceID:       saleID,
			Description:       fmt.Sprintf("Debt Collected (%s): %s", pMethod, saleID),
			Debit:             payment.Amount,
			Credit:            0,
			PaymentMethod:     pMethod,
			TransactionStatus: newStatus,
			CreatedAt:         time.Now(),
		}

		// This function handles the "FOR UPDATE" lock on the last balance
		// and updates the DailySummary closing balance automatically.
		if err := r.cashRepo.CreateEntry(tx, cashEntry); err != nil {
			return fmt.Errorf("cashbook integration failed: %v", err)
		}

		// 4. UPDATE DAILY SUMMARY (Non-balance metrics)
		// CreateEntry handled 'closing_balance'. We only need to update the sales-specific totals.
		tranDateStr := payment.PaymentDate.Format("2006-01-02")
		updateData := map[string]interface{}{
			"debt_collected": gorm.Expr("debt_collected + ?", payment.Amount),
		}

		if pMethod == "KPAY" {
			updateData["k_pay_total"] = gorm.Expr("k_pay_total + ?", payment.Amount)
		}

		if err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ?", tranDateStr).
			Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update daily metrics: %v", err)
		}

		// 5. UPDATE CUSTOMER STATS
		return util.AdjustCustomerStats(tx, sale.CustomerId, payment.Amount, false)
	})
}

func (r *SaleRepository) GetPaymentHistory(saleID string) ([]models.PaymentRecord, error) {
	var history []models.PaymentRecord

	// Fetch all payments linked to this sale, newest first
	err := r.db.Where("reference_id = ?", saleID).
		Order("payment_date DESC").
		Find(&history).Error

	return history, err
}

func (r *SaleRepository) UpdateSale(sale UpdateSaleRemarkDTO) (*models.Sale, error) {
	var existingSale models.Sale
	err := r.db.Where("id = ?", sale.ID).First(&existingSale).Error
	if err != nil {
		return nil, err
	}

	existingSale.Remark = sale.Remark

	log.Println("existingSale to update: ", existingSale)
	err = r.db.Save(&existingSale).Error
	if err != nil {
		return nil, err
	}

	return &existingSale, nil
}

func (r *SaleRepository) ReturnSaleItemsOLD(dto SaleReturnDTO) (*models.SaleReturn, error) {
	tx := r.db.Begin()
	if err := tx.Error; err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var sale models.Sale
	if err := tx.Preload("SaleDetails.Product").First(&sale, "id = ?", dto.SaleID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("sale not found: %v", err)
	}

	// 1️⃣ Create sale return header
	// Using SR-Date-Timestamp for unique ID
	saleReturn := models.SaleReturn{
		ID:          fmt.Sprintf("SR-%s-%d", time.Now().Format("060102"), time.Now().Unix()%10000),
		SaleID:      dto.SaleID,
		Remark:      dto.Remark,
		ReturnDate:  time.Now(),
		TotalAmount: 0,
	}
	if err := tx.Create(&saleReturn).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create sale return header: %v", err)
	}

	var totalReturnAmount int64

	// 2️⃣ Process return items
	for _, item := range dto.ReturnItems {
		var detail models.SaleDetail
		if err := tx.Preload("Product").Where("id = ? AND sale_id = ?", item.ID, dto.SaleID).
			First(&detail).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("sale detail not found (ID %d): %v", item.ID, err)
		}

		remainingQty := detail.Qty - detail.ReturnedQty
		if item.Qty <= 0 || item.Qty > remainingQty {
			tx.Rollback()
			return nil, fmt.Errorf("invalid return qty %d for %s (remaining %d)",
				item.Qty, detail.ProductId, remainingQty)
		}

		// Update Sale Detail stats
		detail.ReturnedQty += item.Qty
		// Logic: Keep original Qty, but track balance in NetQty for financial recalc
		detail.NetQty = detail.Qty - detail.ReturnedQty
		detail.Total = int64(detail.NetQty) * detail.Price
		detail.Remark = fmt.Sprintf("Returned %d items", detail.ReturnedQty)

		if err := tx.Save(&detail).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update sale detail: %v", err)
		}

		// Increase stock since items are coming back
		if err := util.AddStockMovement(tx, detail.ProductId, detail.ProductUnitId, item.Qty, "increase"); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("stock update failed for product %s: %v", detail.ProductId, err)
		}

		// Record specific return item
		returnItem := models.SaleReturnItem{
			SaleReturnID: saleReturn.ID,
			SaleDetailID: detail.ID,
			ProductID:    detail.ProductId,
			Qty:          item.Qty,
			UnitPrice:    int64(detail.Price),
			Total:        int64(item.Qty) * detail.Price,
			CreatedAt:    time.Now(),
		}
		if err := tx.Create(&returnItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to record return item: %v", err)
		}

		// Log item transaction
		itemTxn := models.ItemTransaction{
			ProductId:   detail.ProductId,
			TranType:    "SALE_RETURN",
			InQty:       item.Qty,
			Uom:         detail.Uom,
			ReferenceNo: saleReturn.ID,
			Remark:      fmt.Sprintf("Returned %d %s for Sale %v", item.Qty, detail.Uom, sale.ID),
			CreatedAt:   time.Now(),
		}
		if err := tx.Create(&itemTxn).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to log transaction: %v", err)
		}

		totalReturnAmount += int64(item.Qty) * detail.Price
	}

	// 3️⃣ Recalculate Sale Totals
	var updatedTotal int64
	tx.Model(&models.SaleDetail{}).Where("sale_id = ?", sale.ID).Select("COALESCE(SUM(total),0)").Scan(&updatedTotal)

	sale.Total = updatedTotal
	sale.ReturnAmount += totalReturnAmount
	sale.GrandTotal = sale.Total - sale.Discount

	// Status Logic
	var remaining int64
	tx.Model(&models.SaleDetail{}).Where("sale_id = ? AND net_qty > 0", sale.ID).Count(&remaining)
	if remaining == 0 {
		sale.Status = "RETURNED"
	} else {
		sale.Status = "PARTIAL_RETURN"
	}

	saleReturn.TotalAmount = totalReturnAmount

	if err := tx.Save(&sale).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Save(&saleReturn).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// ==========================================
	// 4️⃣ CASHBOOK LOGIC (Refund Cash Out)
	// ==========================================
	var lastEntry models.Cashbook
	var currentBalance int64 = 0

	// Fetch the very last balance recorded in the system
	err := tx.Order("id desc").First(&lastEntry).Error
	if err == nil {
		currentBalance = lastEntry.Balance
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	cashOutEntry := models.Cashbook{
		TransactionDate: time.Now(),
		TransactionType: "SALE_RETURN",
		ReferenceID:     saleReturn.ID,
		Description:     fmt.Sprintf("Refund for Sale Return %s (Original Sale #%v)", saleReturn.ID, sale.ID),
		Debit:           0,
		Credit:          totalReturnAmount, // Money going out to customer
		Balance:         currentBalance - totalReturnAmount,
		CreatedAt:       time.Now(),
	}

	if err := tx.Create(&cashOutEntry).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to record refund in cashbook: %v", err)
	}

	// 5️⃣ Update Customer Stats (Reduce their total spent)
	if sale.CustomerId != 0 {
		if err := util.AdjustCustomerStats(tx, sale.CustomerId, totalReturnAmount, false); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &saleReturn, nil
}

func (r *SaleRepository) ReturnSaleItems(dto SaleReturnDTO) (*models.SaleReturn, error) {
	// 1. Transaction management (Use r.db.Transaction for cleaner error handling)
	var saleReturn models.SaleReturn

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var sale models.Sale
		if err := tx.Preload("SaleDetails.Product").First(&sale, "id = ?", dto.SaleID).Error; err != nil {
			return fmt.Errorf("sale not found: %v", err)
		}

		// --- Create Header ---
		saleReturn = models.SaleReturn{
			ID:          fmt.Sprintf("SR-%s-%d", time.Now().Format("060102"), time.Now().Unix()%10000),
			SaleID:      dto.SaleID,
			Remark:      dto.Remark,
			ReturnDate:  time.Now(),
			TotalAmount: 0,
		}
		if err := tx.Create(&saleReturn).Error; err != nil {
			return fmt.Errorf("failed to create sale return header: %v", err)
		}

		var totalReturnAmount int64

		// 2. Process return items (Inventory & Totals)
		for _, item := range dto.ReturnItems {
			var detail models.SaleDetail
			if err := tx.Preload("Product").Where("id = ? AND sale_id = ?", item.ID, dto.SaleID).
				First(&detail).Error; err != nil {
				return fmt.Errorf("sale detail not found: %v", err)
			}

			remainingQty := detail.Qty - detail.ReturnedQty
			if item.Qty <= 0 || item.Qty > remainingQty {
				return fmt.Errorf("invalid return qty %d for %s", item.Qty, detail.ProductId)
			}

			// Update Detail
			detail.ReturnedQty += item.Qty
			detail.NetQty = detail.Qty - detail.ReturnedQty
			detail.Total = int64(detail.NetQty) * detail.Price

			if err := tx.Save(&detail).Error; err != nil {
				return err
			}

			// Update Stock
			if err := util.AddStockMovement(tx, detail.ProductId, detail.ProductUnitId, item.Qty, "increase"); err != nil {
				return err
			}

			// Log Transaction
			totalReturnAmount += int64(item.Qty) * detail.Price
		}

		// 3. Update Sale Master Status
		var updatedTotal int64
		tx.Model(&models.SaleDetail{}).Where("sale_id = ?", sale.ID).Select("COALESCE(SUM(total),0)").Scan(&updatedTotal)

		sale.Total = updatedTotal
		sale.ReturnAmount += totalReturnAmount
		sale.GrandTotal = sale.Total - sale.Discount

		var remaining int64
		tx.Model(&models.SaleDetail{}).Where("sale_id = ? AND net_qty > 0", sale.ID).Count(&remaining)
		sale.Status = "PARTIAL_RETURN"
		if remaining == 0 {
			sale.Status = "RETURNED"
		}

		saleReturn.TotalAmount = totalReturnAmount

		if err := tx.Save(&sale).Error; err != nil {
			return err
		}
		if err := tx.Save(&saleReturn).Error; err != nil {
			return err
		}

		// ========================================================
		// 4. CASHBOOK INTEGRATION (Refund Cash Out)
		// ========================================================
		// If the original sale was CASH, we refund CASH.
		// Even if it was DEBT, a return usually implies giving cash back
		// OR reducing debt (In this case, we'll assume Cash Refund).
		cashOutEntry := models.Cashbook{
			TransactionDate: time.Now(),
			TransactionType: "SALE_RETURN",
			ReferenceID:     saleReturn.ID,
			Description:     fmt.Sprintf("Refund for Sale Return %s (Original Sale #%v)", saleReturn.ID, sale.ID),
			Debit:           0,
			Credit:          totalReturnAmount, // Money out
			PaymentMethod:   "CASH",
			CreatedAt:       time.Now(),
		}

		// CreateEntry handles the locking of the last row and calculating the balance
		if err := r.cashRepo.CreateEntry(tx, &cashOutEntry); err != nil {
			return fmt.Errorf("failed to record refund in cashbook: %v", err)
		}

		// 5. Update Daily Summary Metrics (Expense/Refund Total)
		todayStr := time.Now().Format("2006-01-02")
		if err := tx.Model(&models.DailySummaries{}).
			Where("DATE(summary_date) = ?", todayStr).
			Update("expense_total", gorm.Expr("expense_total + ?", totalReturnAmount)).Error; err != nil {
			return fmt.Errorf("failed to update daily summary metrics: %v", err)
		}

		// 6. Update Customer Stats
		if sale.CustomerId != 0 {
			return util.AdjustCustomerStats(tx, sale.CustomerId, totalReturnAmount, false)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &saleReturn, nil
}

func (r *SaleRepository) GetAll() ([]models.Sale, error) {

	sales := []models.Sale{}
	result := r.db.Preload(clause.Associations).Model(&models.Sale{}).Order("sale_date DESC").Find(&sales)

	if result.Error != nil {
		return nil, result.Error
	}
	// if len(sales) == 0 {
	// 	return nil, errors.New("NO records found")
	// }

	return sales, nil
}

func (r *SaleRepository) GetSalesWithReceivablesOld() ([]models.Sale, error) {
	var sales []models.Sale

	result := r.db.Preload(clause.Associations).
		Model(&models.Sale{}).
		// Where("payment_method = ? and payment_status = ?", "DEBT", "PARTIAL").
		Where(" payment_status != ?", "PAID").
		Order("sale_date DESC").
		Find(&sales)

	if result.Error != nil {
		return nil, result.Error
	}
	// if len(sales) == 0 {
	// 	return nil, errors.New("NO records found")
	// }

	return sales, nil
}

func (r *SaleRepository) GetSalesWithReceivables() ([]models.Sale, error) {
	var sales []models.Sale

	// 1. Get the current time and find the start of today (Midnight)
	now := time.Now()
	// time.Local is the correct *time.Location to use here
	todayTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// 2. Convert to Milliseconds (matching your 13+ digit IDs/timestamps)
	// Use .Unix() if your DB stores 10-digit seconds instead.
	todayUnix := todayTime.UnixMilli()

	result := r.db.Preload(clause.Associations).
		Preload("PaymentRecords"). // Vital for the "Trace Back" tab
		// Use the numeric Unix value to match the int8/BigInt column
		Where("payment_status != ? OR updated_at >= ?", "PAID", todayUnix).
		Order("updated_at DESC").
		Find(&sales)

	if result.Error != nil {
		return nil, result.Error
	}

	return sales, nil
}

func (r *SaleRepository) GetSalesByDate(date time.Time) ([]models.Sale, error) {
	sales := []models.Sale{}

	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	result := r.db.Preload(clause.Associations).
		Model(&models.Sale{}).
		Where("sale_date >= ? AND sale_date < ?", startOfDay, endOfDay).
		Order("sale_date DESC").
		Find(&sales)

	if result.Error != nil {
		return nil, result.Error
	}

	return sales, nil
}

// Helper method for today's sales
func (r *SaleRepository) GetTodaySaleList() ([]models.Sale, error) {
	return r.GetSalesByDate(time.Now())
}

func (r *SaleRepository) GetDailySales() ([]ResponseDailySalesDTO, error) {
	type rawSale struct {
		SaleDate time.Time
		Total    int64
	}

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	monthStart := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)
	daysInMonth := nextMonth.Sub(monthStart).Hours() / 24

	// Generate all dates in the month
	var allDates []time.Time
	for d := 0; d < int(daysInMonth); d++ {
		allDates = append(allDates, monthStart.AddDate(0, 0, d))
	}

	// Get sales data
	var salesData []rawSale
	err := r.db.
		Table("sales").
		Select("sale_date::DATE as sale_date, COALESCE(SUM(grand_total), 0) AS total").
		Where("sale_date >= ? AND sale_date < ?", monthStart, nextMonth).
		Group("sale_date::DATE").
		Scan(&salesData).Error

	if err != nil {
		return nil, err
	}

	// Create a map of date to total for easy lookup
	salesMap := make(map[string]int64)
	for _, sale := range salesData {
		salesMap[sale.SaleDate.Format("2006-01-02")] = sale.Total
	}

	// Build results with all dates, filling 0 where no sales
	var results []ResponseDailySalesDTO
	for _, date := range allDates {
		dateStr := date.Format("2006-01-02")
		total, exists := salesMap[dateStr]
		if !exists {
			total = 0
		}
		results = append(results, ResponseDailySalesDTO{
			SaleDate: date.Format("02-01-2006"), // dd-MM-yyyy format
			Total:    total,
		})
	}

	return results, nil
}

func (r *SaleRepository) GetTopTenSoleProducts() ([]ResponseTopTenSoleProductsDTO, error) {
	var results []ResponseTopTenSoleProductsDTO

	err := r.db.
		Table("sale_details sd").
		Select("sd.product_id, sd.product_name, SUM(sd.qty) as total_qty_sold").
		Joins("JOIN sales s ON sd.sale_id = s.id").
		Group("sd.product_id, sd.product_name").
		Order("total_qty_sold DESC").
		Limit(10).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *SaleRepository) TopCustomers() (*ResponseTopCustomerDTO, error) {

	var result ResponseTopCustomerDTO

	err := r.db.Table("sales s").
		Select("cu.name, SUM(s.grand_total) AS total_spent").
		Joins("JOIN customers cu ON s.customer_id = cu.id").
		Group("cu.id, cu.name").
		Order("total_spent DESC").
		Limit(1).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *SaleRepository) GetTodaySales() ([]models.Sale, error) {
	var sales []models.Sale

	// today := time.Now().Format("2006-01-02") // e.g., "2025-07-11"

	loc, _ := time.LoadLocation("Asia/Yangon")
	today := time.Now().In(loc)
	// today := time.Now()
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	end := start.Add(24 * time.Hour)

	// Convert Yangon times to UTC for database query
	startUTC := start.UTC()
	endUTC := end.UTC()

	fmt.Println("start:", start)
	fmt.Println("end:", end)

	result := r.db.
		Preload(clause.Associations).
		Where("sale_date >= ? AND sale_date < ?", startUTC, endUTC).
		// Where("sale_date = ?", today).
		Order("sale_date DESC").
		Find(&sales)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.Error != nil {
		return nil, result.Error
	}
	// if len(sales) == 0 {
	// 	return nil, errors.New("NO records found for today")
	// }

	return sales, nil
}

func (r *SaleRepository) GetById(id string) (*models.Sale, error) {
	var sale models.Sale
	err := r.db.
		Preload("Customer").
		Preload("SaleDetails").
		First(&sale, "id = ?", strings.ToUpper(id)).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Return nil sale, nil error to indicate "not found" gracefully
		return nil, nil
	}

	return &sale, nil
}

func (r *SaleRepository) GetTodayGrandTotal() (int64, error) {
	var total int64
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.Model(&models.Sale{}).
		Select("COALESCE(SUM(grand_total), 0)").
		Where("sale_date >= ? AND sale_date < ?", startOfDay, endOfDay).
		Scan(&total).Error

	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *SaleRepository) GetMonthlySales() ([]models.Sale, error) {
	var sales []models.Sale

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
		Where("sale_date >= ? AND sale_date < ?", monthStartStr, nextMonthStr).
		Order("sale_date DESC").
		Find(&sales).Error

	if err != nil {
		return nil, err
	}
	// if len(sales) == 0 {
	// 	return nil, errors.New("NO records found for this month")
	// }
	return sales, nil
}

func (r *SaleRepository) GetMonthlyGrandTotal() (int64, error) {
	var total int64

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)

	err := r.db.Model(&models.Sale{}).
		Select("COALESCE(SUM(grand_total), 0)").
		Where("sale_date >= ? AND sale_date < ?", monthStart.Format("2006-01-02"), nextMonth.Format("2006-01-02")).
		Scan(&total).Error

	return total, err
}
