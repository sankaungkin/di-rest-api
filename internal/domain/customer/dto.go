package customer

import (
	"time"

	"github.com/sankangkin/di-rest-api/internal/models"
)

type CreateCustomerRequestDTO struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type UpdateCustomerRequstDTO struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}
type CustomerSummaryDTO struct {
	ID           uint          `json:"id"`
	Name         string        `json:"name"`
	Phone        string        `json:"phone"`
	OrderCount   int           `json:"orderCount"`
	TotalSpent   int64         `json:"totalSpent"`
	LastPurchase time.Time     `json:"lastPurchase"`
	RecentSales  []models.Sale `json:"recentSales"`
}
