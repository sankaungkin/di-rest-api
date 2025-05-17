package purchase

import "github.com/sankangkin/di-rest-api/internal/models"
type PurchaseInvoiceRequestDTO struct {
	ID          string              `gorm:"primaryKey" json:"id"`
	SupplierId  uint                `json:"supplierId"`
	PurchaseDetails []models.PurchaseDetail `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"saleDetails"`
	Discount    int64               `json:"discount"`
	Total       int64               `json:"total"`
	GrandTotal  int64               `json:"grandTotal"`
	Remark      string              `json:"remark"`
	PurchaseDate    string          `json:"purchaseDate"`
}