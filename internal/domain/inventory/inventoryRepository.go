package inventory

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type InventoryRepositoryInterface interface {
	Increase(inventory *models.Inventory) (string, error)
	Decrease(inventory *models.Inventory) (string, error)
	Get() ([]models.Product, error)
	GetInvData() ([]ResponseInventoryDTO, error)
}

type InventoryRepository struct {
	db *gorm.DB
}

var (
	repoInstance *InventoryRepository
	repoOnce     sync.Once
)

func NewInventoryRepository(db *gorm.DB) InventoryRepositoryInterface {
	log.Println(util.Cyan + "InventoryRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &InventoryRepository{db: db}
	})
	return repoInstance
}

func (r *InventoryRepository) Increase(input *models.Inventory) (string, error) {

	newInventory := models.Inventory{
		InQty:     input.InQty,
		OutQty:    input.OutQty,
		ProductId: input.ProductId,
		Remark:    input.Remark,
	}

	newItemTransaction := models.ItemTransaction{
		InQty:       int(input.InQty),
		OutQty:      int(input.OutQty),
		ProductId:   input.ProductId,
		TranType:    "DEBIT",
		ReferenceNo: strconv.Itoa(int(input.ID)),
		Remark:      input.Remark,
	}

	tx := r.db.Begin()

	defer func() {
		if recover := recover(); recover != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return "", err
	}
	if err := tx.Create(&newInventory).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	if err := tx.Create(&newItemTransaction).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	var product models.Product
	result := tx.First(&product, "id = ?", input.ProductId)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return "", err
		}
		return "", err
	}
	product.QtyOnHand += int(input.InQty)
	tx.Save(&product)
	tx.Commit()
	message := input.ProductId + " is increased by " + strconv.Itoa(int(input.InQty)) + " EACH"

	return message, nil
}

func (r *InventoryRepository) Decrease(input *models.Inventory) (string, error) {
	newInventory := models.Inventory{
		InQty:     input.InQty,
		OutQty:    input.OutQty,
		ProductId: input.ProductId,
		Remark:    input.Remark,
	}

	newItemTransaction := models.ItemTransaction{
		InQty:       int(input.InQty),
		OutQty:      int(input.OutQty),
		ProductId:   input.ProductId,
		TranType:    "CREDIT",
		ReferenceNo: strconv.Itoa(int(input.ID)),
		Remark:      input.Remark,
	}

	tx := r.db.Begin()

	defer func() {
		if recover := recover(); recover != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return "", err
	}
	if err := tx.Create(&newInventory).Error; err != nil {
		tx.Rollback()
		return "", err
	}
	if err := tx.Create(&newItemTransaction).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	var product models.Product
	result := tx.First(&product, "id = ?", input.ProductId)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return "", err
		}
		return "", err
	}
	product.QtyOnHand -= int(input.OutQty)
	tx.Save(&product)
	tx.Commit()
	message := input.ProductId + " is decrease by " + strconv.Itoa(int(input.OutQty)) + " EACH"

	return message, nil
}

func (r *InventoryRepository) Get() ([]models.Product, error) {
	inventories := []models.Product{}
	r.db.Preload("Product").Model(&models.Inventory{}).Order("ID desc").Find(&inventories)
	if len(inventories) == 0 {
		return nil, errors.New("NO records found")
	}

	return inventories, nil
}

func (r *InventoryRepository) GetInvData() ([]ResponseInventoryDTO, error) {
	var products []models.Product
	var result []ResponseInventoryDTO

	err := r.db.
		Preload("ItemTransactions").
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	for _, p := range products {
		for _, it := range p.ItemTransactions {
			dto := ResponseInventoryDTO{
				ProductName: p.ProductName,
				OutQty:      uint(it.OutQty), // Fixed: using it instead of inv
				InQty:       uint(it.InQty),  // Fixed: using it instead of inv
				ProductId:   p.ID,
				Remark:      it.Remark,   // Fixed: using it instead of inv
				TranType:    it.TranType, // Fixed: using it instead of inv
				QtyOnHand:   p.QtyOnHand,
				CreatedAt:   time.Unix(int64(it.CreatedAt), 0), // Fixed: using it instead of inv
			}
			result = append(result, dto)
		}
	}

	return result, nil
}
