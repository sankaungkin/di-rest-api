package purchase

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

type PurchaseRepositoryInterface interface {
	Create(sale *models.Purchase) (*models.Purchase, error)
	GetAll() ([]models.Purchase, error)
	GetById(id string) (*models.Purchase, error)
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

func (r *PurchaseRepository) Create(input *models.Purchase) (*models.Purchase, error) {

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
			TranType:    "DEBIT",
			ReferenceNo: newPurchase.ID + "-" + strconv.Itoa(int(newPurchase.PurchaseDetails[i].ID)),
			Uom:         newPurchase.PurchaseDetails[i].UnitName,
			// Remark:      "PurchaseID:" + newPurchase.ID + ", line items id:" + strconv.Itoa(int(newPurchase.PurchaseDetails[i].ID)) + ", increase quantity: " + strconv.Itoa(newPurchase.PurchaseDetails[i].Qty) + " " + newPurchase.PurchaseDetails[i].Uom,
			Remark: fmt.Sprintf(
				"PurchaseID:%s, line item id:%d, decrease %d %s ",
				newPurchase.ID, newPurchase.PurchaseDetails[i].ID, newPurchase.PurchaseDetails[i].Qty, newPurchase.PurchaseDetails[i].UnitName,
			),
		}
		tx.Save(&newItemTransaction)

	}
	tx.Commit()
	return &newPurchase, nil

}

func (r *PurchaseRepository) GetAll() ([]models.Purchase, error) {

	purchases := []models.Purchase{}
	r.db.Preload(clause.Associations).Model(&models.Purchase{}).Order("created_at DESC").Find(&purchases)
	if len(purchases) == 0 {
		return nil, errors.New("NO records found")
	}

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
