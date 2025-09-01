package purchase

import (
	"github.com/sankangkin/di-rest-api/internal/models"
)

type PurchaseInvoiceRequestDTO struct {
	ID              string                  `gorm:"primaryKey" json:"id"`
	SupplierId      uint                    `json:"supplierId"`
	PurchaseDetails []models.PurchaseDetail `gorm:"foreignKey:purchaseId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchaseDetails"`
	Discount        int                     `json:"discount"`
	Total           int                     `json:"total"`
	GrandTotal      int                     `json:"grandTotal"`
	Remark          string                  `json:"remark"`
	PurchaseDate    string                  `json:"purchaseDate"`
}
