package sale

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sankangkin/di-rest-api/internal/models"
)

type SaleInvoiceRequestDTO struct {
	ID            string              `gorm:"primaryKey" json:"id"`
	CustomerId    uint                `json:"customerId"`
	SaleDetails   []models.SaleDetail `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"saleDetails"`
	Discount      int64               `json:"discount"`
	Total         int64               `json:"total"`
	GrandTotal    int64               `json:"grandTotal"`
	PaidAmount    int64               `json:"paidAmount"`
	Remark        string              `json:"remark"`
	PaymentMethod string              `json:"paymentMethod"`
	SaleDate      time.Time           `json:"saleDate"`
}

type ResponseTopCustomerDTO struct {
	Name       string `json:"name"`
	TotalSpent int64  `json:"totalSpent"`
}

type ResponseDailySalesDTO struct {
	SaleDate string `json:"saleDate"`
	Total    int64  `json:"total"`
}

type ResponseTopTenSoleProductsDTO struct {
	ProductId    string `json:"productId"`
	ProductName  string `json:"productName"`
	TotalQtySold int64  `json:"totalQtySold"`
}

type ResponseSaleStockItemWithPrice struct {
	ProductUnitId  string `json:"productUnitId"`
	ProductName    string `json:"productName"`
	ProductId      string `json:"productId"`
	UnitId         int    `json:"unitId"`
	UnitName       string `json:"uom"`
	PriceType      string `json:"priceType"`
	UnitPrice      int    `json:"unitPrice"`
	QuantityOnHand int    `json:"quantityOnHand"`
}

type UpdateSaleRemarkDTO struct {
	ID     string `json:"id"`
	Remark string `json:"remark"`
}

type ReturnItem struct {
	ID        int    `json:"id"`
	SaleID    string `json:"saleId"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unitPrice"`
	Total     int    `json:"total"`
}

type SaleReturnDTO struct {
	SaleID      string       `json:"id"`
	ReturnItems []ReturnItem `json:"returnItems"`
	Remark      string       `json:"remark"`
}

type ResponseMonthlyProfitDataDTO struct {
	// MonthYear string `json:"month"` // e.g., "November 2025" or "2025-11"
	Month   sql.NullString `json:"month"`
	Revenue int64          `json:"revenue"`
	COGS    int64          `json:"cogs"`
}

// Add this method to your DTO struct
func (dto ResponseMonthlyProfitDataDTO) MarshalJSON() ([]byte, error) {
	type Alias ResponseMonthlyProfitDataDTO // Create an alias to avoid infinite recursion

	return json.Marshal(&struct {
		Month interface{} `json:"month"` // Override the marshaling for the Month field
		Alias
	}{
		Month: dto.Month.String, // Marshal only the String content
		Alias: (Alias)(dto),
	})
}
