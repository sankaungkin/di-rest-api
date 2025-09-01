package util

import (
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

func ConvertQuantity(db *gorm.DB, productID string, productUnitId string, qty int) (int, error) {
	var unit models.ProductUnit
	if err := db.Where("product_id = ? AND id = ?", productID, productUnitId).First(&unit).Error; err != nil {
		return 0, err
	}
	return qty * unit.ConversionToBase, nil
}

func GetStockBalance(db *gorm.DB, productID string) (int, error) {

	var total int
	err := db.Model(&models.ProductStock{}).
		Where("product_id = ?", productID).
		Select("COALESCE(SUM(base_qty),0)").
		Scan(&total).Error
	return total, err
}

func AddStockMovement(db *gorm.DB, productID string, productUnitId string, qty int, movementType string) error {
	baseQty, err := ConvertQuantity(db, productID, productUnitId, qty)
	if err != nil {
		return err
	}

	var productStock models.ProductStock
	if err := db.Where("product_id = ?", productID).First(&productStock).Error; err != nil {
		return err
	}

	productStock.BaseQty -= baseQty

	return db.Save(&productStock).Error
}
