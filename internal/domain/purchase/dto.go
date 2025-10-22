package purchase

import (
	"time"

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
	PurchaseDate    time.Time               `json:"purchaseDate"`
}

type ResponsePurchaseLineItemDTO struct {
	ProductUnitId string `json:"productUnitId"`
	ProductName   string `json:"productName"`
	ProductId     string `json:"productId"`
	UnitId        int    `json:"unitId"`
	UnitName      string `json:"unitName"`
	PriceUnitId   int    `json:"priceUnitId"`
	PriceType     string `json:"priceType"`
	UnitPrice     int    `json:"unitPrice"`
}

type UpdateRemarkPurchaseDTO struct {
	ID     string `json:"id"`
	Remark string `json:"remark"`
}
