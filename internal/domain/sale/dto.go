package sale

import "github.com/sankangkin/di-rest-api/internal/models"

type SaleInvoiceRequestDTO struct {
	ID          string              `gorm:"primaryKey" json:"id"`
	CustomerId  uint                `json:"customerId"`
	SaleDetails []models.SaleDetail `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"saleDetails"`
	Discount    int64               `json:"discount"`
	Total       int64               `json:"total"`
	GrandTotal  int64               `json:"grandTotal"`
	Remark      string              `json:"remark"`
	SaleDate    string              `jsong:"saleDate"`
}