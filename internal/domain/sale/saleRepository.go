package sale

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaleRepositoryInterface interface {
	Create(sale *models.Sale) (*models.Sale, error)
	GetAll() ([]models.Sale, error)
	GetById(id string) (*models.Sale, error)
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

// Create inserts a new sale record into the database.
// It validates the sale data, handles stock updates, and manages transactions.
// If any error occurs during the process, it rolls back the transaction and returns the error.
// If successful, it commits the transaction and returns the created sale record.
func (r *SaleRepository) CreateOld(input *models.Sale) (*models.Sale, error) {
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

	// Basic validation
	err := models.ValidateStruct(newSale)
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

	// Save sale
	if err := tx.Create(&newSale).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range newSale.SaleDetails {
		sd := newSale.SaleDetails[i]

		var productStock models.ProductStock
		if err := tx.First(&productStock, "product_id = ?", sd.ProductId).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		var unitConv models.UnitConversion
		if err := tx.First(&unitConv, "product_id = ?", sd.ProductId).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("unit conversion not found for product %s", sd.ProductId)
		}

		// Handle sale by base unit
		if sd.Uom == unitConv.BaseUnit {
			if sd.Qty > productStock.BaseQty {
				tx.Rollback()
				return nil, fmt.Errorf("not enough stock: base unit of %s. requested %d, available %d", sd.ProductId, sd.Qty, productStock.BaseQty)
			}

			productStock.BaseQty -= sd.Qty

			// Save item transaction
			trx := models.ItemTransaction{
				ProductId:   sd.ProductId,
				ReferenceNo: newSale.ID + "-" + strconv.Itoa(int(sd.ID)),
				OutQty:      sd.Qty,
				Uom:         sd.Uom,
				TranType:    "CREDIT",
				Remark:      fmt.Sprintf("SaleId %s, SaleDetailId %d, ProductId %s, Sold %d %s (base unit)", sd.SaleId, sd.ID, sd.ProductId, sd.Qty, sd.Uom),
			}
			if err := tx.Create(&trx).Error; err != nil {
				tx.Rollback()
				return nil, err
			}

		} else if sd.Uom == unitConv.DeriveUnit {
			if sd.DerivedQty > productStock.DerivedQty {
				tx.Rollback()
				return nil, fmt.Errorf("not enough stock: derived unit of %s. requested %d, available %d", sd.ProductId, sd.DerivedQty, productStock.DerivedQty)
			}

			productStock.DerivedQty -= sd.DerivedQty

			// Save item transaction
			trx := models.ItemTransaction{
				ProductId:   sd.ProductId,
				ReferenceNo: newSale.ID + "-" + strconv.Itoa(int(sd.ID)),
				OutQty:      sd.DerivedQty,
				Uom:         sd.Uom,
				TranType:    "CREDIT",
				Remark:      fmt.Sprintf("SaleId %s, SaleDetailId %d, ProductId Sold %s %d %s (derived unit)", sd.SaleId, sd.ID, sd.ProductId, sd.DerivedQty, sd.Uom),
			}
			if err := tx.Create(&trx).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("invalid unit %s for product %s", sd.Uom, sd.ProductId)
		}

		// Save updated stock
		if err := tx.Save(&productStock).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &newSale, nil
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

	if err := tx.Preload("SaleDetails").First(&newSale, "id = ?", newSale.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range newSale.SaleDetails {
		sd := &newSale.SaleDetails[i]

		if err := adjustProductStock(tx, newSale.ID, sd); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &newSale, nil
}

func adjustProductStock(tx *gorm.DB, saleId string, sd *models.SaleDetail) error {
	var productStock models.ProductStock
	if err := tx.First(&productStock, "product_id = ?", sd.ProductId).Error; err != nil {
		return err
	}

	var unitConv models.UnitConversion
	if err := tx.First(&unitConv, "product_id = ?", sd.ProductId).Error; err != nil {
		return fmt.Errorf("unit conversion not found for product %s", sd.ProductId)
	}

	factor := int(unitConv.Factor)

	switch {
	case strings.EqualFold(sd.Uom, unitConv.BaseUnit):
		if sd.Qty > productStock.BaseQty {
			return fmt.Errorf("not enough stock: base unit of %s. requested %d, available %d", sd.ProductId, sd.Qty, productStock.BaseQty)
		}
		productStock.BaseQty -= sd.Qty

		// Log base unit transaction
		trx := models.ItemTransaction{
			ProductId:   sd.ProductId,
			ReferenceNo: saleId + "-" + strconv.Itoa(int(sd.ID)),
			OutQty:      sd.Qty,
			Uom:         sd.Uom,
			TranType:    "CREDIT",
			Remark:      fmt.Sprintf("SaleId %s, SaleDetailId %d, ProductId %s, Sold %d %s (base unit)", sd.SaleId, sd.ID, sd.ProductId, sd.Qty, sd.Uom),
		}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}

	case strings.EqualFold(sd.Uom, unitConv.DeriveUnit):
		totalNeeded := sd.DerivedQty

		if totalNeeded <= productStock.DerivedQty {
			productStock.DerivedQty -= totalNeeded
		} else {
			shortage := totalNeeded - productStock.DerivedQty
			productStock.DerivedQty = 0

			baseToConvert := (shortage + factor - 1) / factor // round up
			if baseToConvert > productStock.BaseQty {
				return fmt.Errorf("not enough stock for derived sale of product %s: need %d %s → convert %d base units, only %d available",
					sd.ProductId, totalNeeded, unitConv.DeriveUnit, baseToConvert, productStock.BaseQty)
			}

			productStock.BaseQty -= baseToConvert
			convertedDerived := baseToConvert * factor
			productStock.DerivedQty = convertedDerived - shortage
		}

		// Log derived unit transaction
		trx := models.ItemTransaction{
			ProductId:   sd.ProductId,
			ReferenceNo: saleId + "-" + strconv.Itoa(int(sd.ID)),
			OutQty:      sd.DerivedQty,
			Uom:         sd.Uom,
			TranType:    "CREDIT",
			Remark:      fmt.Sprintf("SaleId %s, SaleDetailId %d, ProductId %s, Sold %d %s (derived unit)", sd.SaleId, sd.ID, sd.ProductId, sd.DerivedQty, sd.Uom),
		}
		if err := tx.Create(&trx).Error; err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid unit %s for product %s (expected %s or %s)", sd.Uom, sd.ProductId, unitConv.BaseUnit, unitConv.DeriveUnit)
	}

	return tx.Save(&productStock).Error
}

func (r *SaleRepository) GetAll() ([]models.Sale, error) {

	sales := []models.Sale{}
	r.db.Preload(clause.Associations).Model(&models.Sale{}).Order("ID desc").Find(&sales)
	if len(sales) == 0 {
		return nil, errors.New("NO records found")
	}

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
			return nil, err
		}
		return nil, err
	}

	return &sale, nil
}
