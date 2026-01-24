package purchase

import (
	"time"

	"github.com/sankangkin/di-rest-api/internal/models"
)

type PurchaseInvoiceRequestDTO struct {
	ID              string                  `gorm:"primaryKey" json:"id"`
	SupplierId      uint                    `json:"supplierId"`
	PurchaseDetails []models.PurchaseDetail `gorm:"foreignKey:purchaseId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchaseDetails"`
	PaymentSource   string                  `json:"paymentSource"`
	AmountFromCash  int64                   `json:"amountFromCash"`
	AmountFromOwner int64                   `json:"amountFromOwner"`
	Discount        int                     `json:"discount"`
	Total           int                     `json:"total"`
	GrandTotal      int64                   `json:"grandTotal"`
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
	StockQty      int    `json:"stockQty"`
}

type UpdateRemarkPurchaseDTO struct {
	ID     string `json:"id"`
	Remark string `json:"remark"`
}

type ResponseHistoricalCOGS struct {
	MonthYear string `json:"month"` // e.g., "2025-11"
	COGS      int64  `json:"cogs"`
}
type AgingSummary struct {
	Category     string  `json:"category"`
	TotalBalance float64 `json:"totalBalance"`
	PoCount      int     `json:"poCount"`
}

type PaymentRequest struct {
	// Matches "amount" in JSON
	Amount int64 `json:"amount" validate:"required,gt=0"`

	// Matches "paymentMethod" in JSON ("CASH" or "KPAY")
	PaymentMethod string `json:"paymentMethod" validate:"required"`

	// Matches "paymentDate" ISO string from Angular
	PaymentDate time.Time `json:"paymentDate"`

	// Matches "purchaseId" in JSON
	PurchaseId string `json:"purchaseId" validate:"required"`
}
