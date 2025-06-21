package dto

import "github.com/sankangkin/di-rest-api/internal/models"

type LoginDTO struct {
	Email    string `json:"email" validate:"required,email" example:"user@email.com"`
	Password string `json:"password" validate:"required" example:"pass1234"`
}

type LoginResponseDTO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type CreateUserRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
	Role     string `json:"role"`
}

type UpdateCategoryRequestDTO struct {
	CategoryName string `json:"categoryName"`
}

type CreateCategoryRequestDTO struct {
	CategoryName string `json:"categoryName"`
}

type CreateProductRequstDTO struct {
	ID              string `gorm:"primaryKey" json:"id"`
	ProductName     string `json:"productName" validate:"required,min=3"`
	CategoryId      uint   `json:"categoryId" validate:"required"`
	Uom             string `json:"uom" validate:"required,min=3"`
	BuyPrice        int64  `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64  `json:"sellPricelvl1" validate:"required,min=1"`
	SellPriceLevel2 int64  `json:"sellPricelvl2" validate:"required,min=1"`
	ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	QtyOnHand       int    `json:"qtyOhHand" validate:"required"`
	BrandName       string `json:"brandName"`
	IsActive        bool   `json:"isActive" gorm:"default:true"`
}

type UpdateProductRequstDTO struct {
	ProductName     string `json:"productName" validate:"required,min=3"`
	CategoryId      uint   `json:"categoryId" validate:"required"`
	Uom             string `json:"uom" validate:"required,min=2"`
	BuyPrice        int64  `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64  `json:"sellPricelvl1" validate:"required,min=1"`
	SellPriceLevel2 int64  `json:"sellPricelvl2" validate:"required,min=1"`
	ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	// QtyOnHand       int    `json:"qtyOhHand" validate:"required"`
	BrandName string `json:"brandName"`
	IsActive  bool   `json:"isActive" gorm:"default:true"`
}

type SaleInvoiceRequestDTO struct {
	ID          string              `gorm:"primaryKey" json:"id"`
	CustomerId  uint                `json:"customerId"`
	SaleDetails []models.SaleDetail `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"saleDetails"`
	Discount    int64               `json:"discount"`
	Total       int64               `json:"total"`
	GrandTotal  int64               `json:"grandTotal"`
	Remark      string              `json:"remark"`
	SaleDate    string              `json:"saleDate"`
}

type PurchaseInvoiceRequestDTO struct {
	ID              string                  `gorm:"primaryKey" json:"id"`
	SupplierId      uint                    `json:"supplierId"`
	PurchaseDetails []models.PurchaseDetail `gorm:"foreignKey:PurchaseId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchaseDetails"`
	Discount        int64                   `json:"discount"`
	Total           int64                   `json:"total"`
	GrandTotal      int64                   `json:"grandTotal"`
	Remark          string                  `json:"remark"`
	PurchaseDate    string                  `json:"purchaseDate"`
}
