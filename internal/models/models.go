package models

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type Category struct {
	gorm.Model
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryName string    `json:"categoryName" validate:"required,min=3"`
	Products     []Product `gorm:"foreignKey:CategoryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt    int64     `gorm:"autoCreateTime" json:"-"`
	UpdatedAt    int64     `gorm:"autoUpdateTime:milli" json:"-"`
}

type Product struct {
	gorm.Model
	ID              string           `gorm:"primaryKey" json:"id"`
	ProductName     string           `json:"productName" validate:"required,min=3"`
	CategoryId      uint             `json:"categoryId"`
	Inventories     []Inventory      `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	SaleDetail      []SaleDetail     `gorm:"foreignKey:ProductId;" json:"-"`
	PurchaseDetail  []PurchaseDetail `gorm:"foreignKey:ProductId;" json:"-"`
	Uom             string           `json:"uom" validate:"required,min=3"`
	BuyPrice        int64            `josn:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64            `josn:"sellPricelvl1" validate:"required,min=1"`
	SellPriceLevel2 int64            `josn:"sellPricelvl2" validate:"required,min=1"`
	ReorderLvl      uint             `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	QtyOnHand       int              `json:"qtyOhHand" validate:"required"`
	BrandName       string           `json:"brand"`
	IsActive        bool             `json:"isActive" gorm:"default:true"`
	CreatedAt       int64            `gorm:"autoCreateTime" json:"-"`
	UpdatedAt       int64            `gorm:"autoUpdateTime:milli" json:"-"`
}

type User struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string `gorm:"uniqueIndex;" json:"email" validate:"required,email"`
	UserName  string `json:"username" validate:"required,min=3"`
	Password  string `json:"password" validate:"required,min=3"`
	IsAdmin   bool   `json:"isAdmin" validate:"required"`
	Role      string `json:"role" validate:"required" gorm:"default:user"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"-"`
}

type Customer struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Sales     []Sale `gorm:"foreignKey:CustomerId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"-"`
}

type Supplier struct {
	gorm.Model
	ID        uint       `gorm:"primaryKey:autoIncrement" json:"id"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	Phone     string     `json:"phone"`
	Purchases []Purchase `gorm:"foreignKey:SupplierId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt int64      `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64      `gorm:"autoUpdateTime:milli" json:"-"`
}

type Purchase struct {
	gorm.Model
	ID              string           `gorm:"primaryKey" json:"id"`
	SupplierId      uint             `json:"supplierId"`
	Supplier        *Supplier        `json:"supplier"`
	PurchaseDetails []PurchaseDetail `gorm:"foreignKey:PurchaseId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchaseDetails"`
	Discount        int64            `json:"discount"`
	Total           int64            `json:"total"`
	GrandTotal      int64            `json:"grandTotal"`
	Remark          string           `json:"remark"`
	PurchaseDate    string           `json:"purchaseDate"`
	CreatedAt       int64            `gorm:"autoCreateTime" json:"-"`
	UpdatedAt       int64            `gorm:"autoUpdateTime:milli" json:"-"`
}

type PurchaseDetail struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	ProductId   string `json:"productId"`
	ProductName string `json:"productName"`
	Qty         int    `json:"qty"`
	Price       int64  `json:"price"`
	Total       int64  `json:"total"`
	PurchaseId  string `json:"purchaseId"`
}

type Inventory struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	OutQty    uint   `json:"inQty"`
	InQty     uint   `json:"outQty"`
	ProductId string `json:"productId"`
	Remark    string `json:"remark"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"-"`
}

type Sale struct {
	gorm.Model
	ID          string       `gorm:"primaryKey" json:"id"`
	CustomerId  uint         `json:"customerId"`
	Customer    *Customer    `json:"customer"`
	SaleDetails []SaleDetail `gorm:"foreignKey:SaleId;reference:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"saleDetails"`
	Discount    int64        `json:"discount"`
	Total       int64        `json:"total"`
	GrandTotal  int64        `json:"grandTotal"`
	Remark      string       `json:"remark"`
	SaleDate    string       `json:"saleDate"`
	CreatedAt   int64        `gorm:"autoCreateTime" json:"-"`
	UpdatedAt   int64        `gorm:"autoUpdateTime:milli" json:"-"`
}

type SaleDetail struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	ProductId   string `json:"productId"`
	ProductName string `json:"productName"`
	Qty         int    `json:"qty"`
	Price       int64  `json:"price"`
	Total       int64  `json:"total"`
	SaleId      string `json:"saleId"`
}

type ItemTransaction struct {
	gorm.Model
	// TODO to enhance with UUID
	ID          uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	ProductId   string `json:"productId"`
	ReferenceNo string `json:"referenceNo"`
	InQty       int    `json:"inQty"`
	OutQty      int    `json:"outQty"`
	TranType    string `json:"tranType"`
	Remark      string `json:"remark"`
	CreatedAt   int64  `gorm:"autoCreateTiem" json:"createdTime"`
}

type ErrorResponse struct {
	Field string                                 `json:"field"`
	Tag   string                                 `json:"tag"`
	Value string                                 `json:"value,omitempty"`
	Info  validator.ValidationErrorsTranslations `json:"info"`
}

func ValidateStruct[T any](payload T) []*ErrorResponse {

	en := en.New()
	uni := ut.New(en, en)

	trans, _ := uni.GetTranslator("en")

	validate := validator.New()
	en_translations.RegisterDefaultTranslations(validate, trans)

	var errors []*ErrorResponse
	err := validate.Struct(payload)

	if err != nil {

		errTran := err.(validator.ValidationErrors)
		fmt.Println(errTran.Translate(trans))
		info := errTran.Translate(trans)

		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			element.Info = info
			errors = append(errors, &element)
		}

	}
	return errors
}

func MigrateModels(db *gorm.DB) error {
	err := db.AutoMigrate(&Category{}, &Product{}, &User{}, &Customer{}, &Supplier{}, &Sale{}, &SaleDetail{}, &Purchase{}, &PurchaseDetail{}, &ItemTransaction{}, &User{})
	return err
}
