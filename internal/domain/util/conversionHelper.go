package util

import (
	"fmt"
	"strings"

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

// func AddStockMovement(db *gorm.DB, productID string, productUnitId string, qty int, movementType string) error {
// 	qty, err := ConvertQuantity(db, productID, productUnitId, qty)
// 	if err != nil {
// 		return err
// 	}

// 	var productStock models.ProductStock
// 	if err := db.Where("product_id = ?", productID).First(&productStock).Error; err != nil {
// 		return err
// 	}

// 	productStock.DerivedQty -= qty

// 	return db.Save(&productStock).Error
// }

func AddStockMovement(db *gorm.DB, productID string, productUnitId string, qty int, movementType string) error {
	// Convert to base unit if necessary
	qty, err := ConvertQuantity(db, productID, productUnitId, qty)
	if err != nil {
		return err
	}

	var productStock models.ProductStock
	if err := db.Where("product_id = ?", productID).First(&productStock).Error; err != nil {
		return err
	}

	// Handle increase/decrease by movement type
	switch strings.ToLower(movementType) {
	case "increase", "in", "purchase", "sale_return", "adjust_in":
		productStock.DerivedQty += qty

	case "decrease", "out", "sale", "adjust_out":
		productStock.DerivedQty -= qty

	default:
		return fmt.Errorf("invalid movement type: %s", movementType)
	}

	// Optional: prevent negative stock
	if productStock.DerivedQty < 0 {
		return fmt.Errorf("insufficient stock for product %s", productID)
	}

	return db.Save(&productStock).Error
}
