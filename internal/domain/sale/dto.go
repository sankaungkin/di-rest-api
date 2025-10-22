package sale

import (
	"time"

	"github.com/sankangkin/di-rest-api/internal/models"
)

type SaleInvoiceRequestDTO struct {
	ID          string              `gorm:"primaryKey" json:"id"`
	CustomerId  uint                `json:"customerId"`
	SaleDetails []models.SaleDetail `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"saleDetails"`
	Discount    int64               `json:"discount"`
	Total       int64               `json:"total"`
	GrandTotal  int64               `json:"grandTotal"`
	Remark      string              `json:"remark"`
	SaleDate    time.Time           `json:"saleDate"`
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
	ProductUnitId string `json:"productUnitId"`
	ProductName   string `json:"productName"`
	ProductId     string `json:"productId"`
	UnitId        int    `json:"unitId"`
	UnitName      string `json:"uom"`
	PriceType     string `json:"priceType"`
	UnitPrice     int    `json:"unitPrice"`
}

type UpdateSaleRemarkDTO struct {
	ID     string `json:"id"`
	Remark string `json:"remark"`
}
