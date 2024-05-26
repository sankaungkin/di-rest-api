package sale

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SaleRepositoryInterface interface{
	Create(sale *models.Sale) (*models.Sale, error)
	GetAll() ([]models.Sale, error)
	GetById(id string) (*models.Sale, error)
}

type SaleRepository struct{
	db *gorm.DB
}

var (
	repoInstance *SaleRepository
	repoOnce sync.Once
)

func NewSaleRepository(db *gorm.DB) SaleRepositoryInterface {
	log.Println(util.Blue + "SaleRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &SaleRepository{db: db}
	})
	return repoInstance
}

func (r *SaleRepository)Create(input *models.Sale) (*models.Sale, error ){
	
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

	if err := tx.Create(&newSale).Error; err != nil {
		tx.Rollback()
		return nil, err
	}	
	for i := range newSale.SaleDetails {

		// decrease product qtyonhand
		var product models.Product
		result := tx.First(&product, "id = ?", newSale.SaleDetails[i].ProductId)
		if err := result.Error; err != nil {
			return nil, err
		}
		product.QtyOnHand -= int(newSale.SaleDetails[i].Qty)
		tx.Save(&product)

		// create inventory record
		newInventory := models.Inventory{
			InQty:     0,
			OutQty:    uint(newSale.SaleDetails[i].Qty),
			ProductId: newSale.SaleDetails[i].ProductId,
			Remark:    "SaleID:" + newSale.ID + ", line items id:" + strconv.Itoa(int(newSale.SaleDetails[i].ID)) + ", decrease quantity: " + strconv.Itoa(newSale.SaleDetails[i].Qty),
		}
		tx.Save(&newInventory)

		newItemTransaction := models.ItemTransaction{
			InQty:       0,
			OutQty:      newSale.SaleDetails[i].Qty,
			ProductId:   newSale.SaleDetails[i].ProductId,
			TranType:    "CREDIT",
			ReferenceNo: newSale.ID + "-" + strconv.Itoa(int(newSale.SaleDetails[i].ID)),
			Remark:      "SaleID:" + newSale.ID + ", line items id:" + strconv.Itoa(int(newSale.SaleDetails[i].ID)) + ", decrease quantity: " + strconv.Itoa(newSale.SaleDetails[i].Qty),
		}
		tx.Save(&newItemTransaction)

	}
	tx.Commit()
	return &newSale, nil

}

func (r *SaleRepository)GetAll() ([]models.Sale, error) {

	sales := []models.Sale{}
	r.db.Preload(clause.Associations).Model(&models.Sale{}).Order("ID desc").Find(&sales)
	if len(sales) == 0 {
		return nil, errors.New("NO records found")
	}

	return sales, nil
}

func (r *SaleRepository)GetById(id string) (*models.Sale, error) {

	var sale models.Sale
	result := r.db.First(&sale, "id = ?", strings.ToUpper(id))
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return &sale, nil
}