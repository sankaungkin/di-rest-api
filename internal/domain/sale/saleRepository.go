package sale

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

type SaleRepositoryInterface interface {
	Create(sale *models.Sale) (*models.Sale, error)
	GetAll() ([]models.Sale, error)
	GetDailySales() ([]ResponseDailySalesDTO, error)
	GetTodaySales() ([]models.Sale, error)
	GetTopTenSoleProducts() ([]ResponseTopTenSoleProductsDTO, error)
	GetById(id string) (*models.Sale, error)
	GetTodayGrandTotal() (int64, error)
	GetMonthlySales() ([]models.Sale, error)
	GetMonthlyGrandTotal() (int64, error)
	TopCustomers() (*ResponseTopCustomerDTO, error)
}

type SaleRepository struct {
	db *gorm.DB
}

var (
	repoInstance *SaleRepository
	repoOnce     sync.Once
)

func NewSaleRepository(db *gorm.DB) SaleRepositoryInterface {
	log.Println(util.Blue + "SaleRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &SaleRepository{db: db}
	})
	return repoInstance
}

func (r *SaleRepository) Create(input *models.Sale) (*models.Sale, error) {
	newSale := models.Sale{
		ID:          input.ID,
		CustomerId:  input.CustomerId,
		Discount:    input.Discount,
		GrandTotal:  input.GrandTotal,
		Remark:      input.Remark,
		SaleDate:    input.SaleDate,
		SaleDetails: input.SaleDetails,
		Total:       input.Total,
	}

	if err := models.ValidateStruct(newSale); err != nil {
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

	if err := tx.Create(&newSale).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Preload all necessary relationships
	if err := tx.Preload("SaleDetails.Product").
		First(&newSale, "id = ?", newSale.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Preload("SaleDetails").First(&newSale, "id = ?", newSale.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range newSale.SaleDetails {

		sd := &newSale.SaleDetails[i]

		tx.Preload("Product").Preload("Uom").First(sd, sd.ID)

		if err := util.AddStockMovement(tx, sd.ProductId, sd.ProductUnitId, sd.Qty, "increase"); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to add stock movement: %v", err)
		}
		newItemTransactions := models.ItemTransaction{
			ProductId:   sd.ProductId,
			TranType:    "SALE",
			OutQty:      sd.Qty,
			Uom:         sd.Uom,
			ReferenceNo: input.ID + "-" + strconv.Itoa(int(sd.ID)),
			Remark: fmt.Sprintf(
				"SaleId:%s, SaleDetailId %d, ProductId %s : %s, Sold %d %s",
				input.ID, sd.ID, sd.ProductId, sd.Product.ProductName, sd.Qty, sd.Uom,
			),
			CreatedAt: time.Now().Local(),
		}
		tx.Create(&newItemTransactions)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &newSale, nil
}

// func adjustProductStock(tx *gorm.DB, saleId string, sd *models.SaleDetail) error {
// 	var productStock models.ProductStock
// 	if err := tx.First(&productStock, "product_id = ?", sd.ProductId).Error; err != nil {
// 		return err
// 	}

// 	//broadcast low stock message

// 	var unitConv models.UnitConversion
// 	if err := tx.First(&unitConv, "product_id = ?", sd.ProductId).Error; err != nil {
// 		return fmt.Errorf("unit conversion not found for product %s", sd.ProductId)
// 	}

// 	factor := int(unitConv.Factor)

// 	switch {
// 	case strings.EqualFold(sd.Uom, unitConv.BaseUnit):
// 		if sd.Qty > productStock.BaseQty {
// 			return fmt.Errorf("not enough stock: base unit of %s. requested %d, available %d", sd.ProductId, sd.Qty, productStock.BaseQty)
// 		}
// 		productStock.BaseQty -= sd.Qty

// 		// Log base unit transaction
// 		trx := models.ItemTransaction{
// 			ProductId:   sd.ProductId,
// 			ReferenceNo: saleId + "-" + strconv.Itoa(int(sd.ID)),
// 			OutQty:      sd.Qty,
// 			Uom:         sd.Uom,
// 			TranType:    "CREDIT",
// 			Remark:      fmt.Sprintf("SaleId %s, SaleDetailId %d, ProductId %s, Sold %d %s (base unit)", sd.SaleId, sd.ID, sd.ProductId, sd.Qty, sd.Uom),
// 		}
// 		if err := tx.Create(&trx).Error; err != nil {
// 			return err
// 		}

// 	case strings.EqualFold(sd.Uom, unitConv.DeriveUnit):
// 		totalNeeded := sd.DerivedQty

// 		if totalNeeded <= productStock.DerivedQty {
// 			productStock.DerivedQty -= totalNeeded
// 		} else {
// 			shortage := totalNeeded - productStock.DerivedQty
// 			productStock.DerivedQty = 0

// 			baseToConvert := (shortage + factor - 1) / factor // round up
// 			if baseToConvert > productStock.BaseQty {
// 				return fmt.Errorf("not enough stock for derived sale of product %s: need %d %s → convert %d base units, only %d available",
// 					sd.ProductId, totalNeeded, unitConv.DeriveUnit, baseToConvert, productStock.BaseQty)
// 			}

// 			productStock.BaseQty -= baseToConvert
// 			convertedDerived := baseToConvert * factor
// 			productStock.DerivedQty = convertedDerived - shortage
// 		}

// 		// Log derived unit transaction
// 		trx := models.ItemTransaction{
// 			ProductId:   sd.ProductId,
// 			ReferenceNo: saleId + "-" + strconv.Itoa(int(sd.ID)),
// 			OutQty:      sd.DerivedQty,
// 			Uom:         sd.Uom,
// 			TranType:    "CREDIT",
// 			Remark:      fmt.Sprintf("SaleId %s, SaleDetailId %d, ProductId %s, Sold %d %s (derived unit)", sd.SaleId, sd.ID, sd.ProductId, sd.DerivedQty, sd.Uom),
// 		}
// 		if err := tx.Create(&trx).Error; err != nil {
// 			return err
// 		}

// 	default:
// 		return fmt.Errorf("invalid unit %s for product %s (expected %s or %s)", sd.Uom, sd.ProductId, unitConv.BaseUnit, unitConv.DeriveUnit)
// 	}

// 	return tx.Save(&productStock).Error
// }

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

func (r *SaleRepository) GetDailySalesOLD() ([]ResponseDailySalesDTO, error) {
	var results []ResponseDailySalesDTO

	// Get the first day of the current month
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err := r.db.
		Table("sales").
		Select("sale_date::DATE , SUM(grand_total) AS total ").
		Where("sale_date >= ?", monthStart).
		Group("sale_date::DATE").
		Order("sale_date::DATE DESC").
		Scan(&results).Error

	return results, err
}

func (r *SaleRepository) GetDailySalesOld() ([]ResponseDailySalesDTO, error) {
	type rawSale struct {
		SaleDate time.Time
		Total    int64
	}

	var rawResults []rawSale
	var results []ResponseDailySalesDTO

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err := r.db.
		Table("sales").
		Select("sale_date::DATE as sale_date, SUM(grand_total) AS total").
		Where("sale_date >= ?", monthStart).
		Group("sale_date::DATE").
		Order("sale_date::DATE DESC").
		Scan(&rawResults).Error

	if err != nil {
		return nil, err
	}

	// Format date to dd-MM-yyyy
	for _, raw := range rawResults {
		formatted := ResponseDailySalesDTO{
			SaleDate: raw.SaleDate.Format("02-01-2006"), // dd-MM-yyyy
			Total:    raw.Total,
		}
		results = append(results, formatted)
	}

	return results, nil
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
