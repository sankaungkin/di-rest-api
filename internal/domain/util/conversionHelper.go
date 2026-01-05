package util

import (
	"fmt"
	"strings"
	"time"

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
	case "increase", "in", "purchase", "sale_return", "adjust_in", "buy":
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

// internal/util/customer_stats.go

func UpdateCustomerBalance(tx *gorm.DB, customerID uint, amount int64, isIncrease bool) error {
	if customerID == 0 {
		return nil
	}

	operation := "+"
	if !isIncrease {
		operation = "-"
	}

	return tx.Model(&models.Customer{}).Where("id = ?", customerID).
		Updates(map[string]interface{}{
			"total_spent":   gorm.Expr(fmt.Sprintf("total_spent %s ?", operation), amount),
			"order_count":   gorm.Expr(fmt.Sprintf("order_count %s ?", operation), 1), // Only if you want to decrease count on returns
			"last_purchase": time.Now(),
		}).Error
}

// internal/util/customer_stats.go

func AdjustCustomerStats(tx *gorm.DB, customerID uint, amount int64, isNewSale bool) error {
	if customerID == 0 {
		return nil
	}

	updates := make(map[string]interface{})

	if isNewSale {
		// New Sale: Increase both Count and Spent
		updates["order_count"] = gorm.Expr("order_count + ?", 1)
		updates["total_spent"] = gorm.Expr("total_spent + ?", amount)
		updates["last_purchase"] = time.Now()
	} else {
		// Return: Only reduce the Spent amount
		updates["total_spent"] = gorm.Expr("total_spent - ?", amount)
	}

	return tx.Model(&models.Customer{}).Where("id = ?", customerID).Updates(updates).Error
}
