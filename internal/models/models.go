package models

import (
	"fmt"
	"time"

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
	Products     []Product `gorm:"foreignKey:CategoryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"products"`
	CreatedAt    int64     `gorm:"autoCreateTime" json:"-"`
	UpdatedAt    int64     `gorm:"autoUpdateTime:milli" json:"-"`
}

type Product struct {
	gorm.Model
	ID               string            `gorm:"primaryKey" json:"id"`
	ProductName      string            `json:"productName" validate:"required,min=3"`
	CategoryId       uint              `json:"categoryId"`
	UnitConversion   []UnitConversion  `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"unitConversion"`
	Inventories      []Inventory       `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"inventories"`
	SaleDetail       []SaleDetail      `gorm:"foreignKey:ProductId;" json:"saleDetails"`
	PurchaseDetail   []PurchaseDetail  `gorm:"foreignKey:ProductId;" json:"purchaseDetails"`
	ItemTransactions []ItemTransaction `gorm:"foreignKey:ProductId;"  json:"itemTransactions"`
	Uom              string            `json:"uom"`
	DeriveUom        string            `json:"deriveUom"`
	UomId            uint              `json:"uomId" validate:"required"`
	DeriveUomId      uint              `json:"deriveUomId" validate:"required"`
	BuyPrice         int64             `json:"buyPrice"`
	SellPriceLevel1  int64             `json:"sellPricelvl1" `
	DeriveUnitPrice  int64             `json:"deriveUnitPrice" `
	BrandName        string            `json:"brandName"`
	IsActive         bool              `json:"isActive" gorm:"default:true"`
	CreatedAt        int64             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt        int64             `gorm:"autoUpdateTime:milli" json:"-"`
}

type ProductPrice struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductId string `gorm:"index:idx_product_unit_type,unique" json:"productId" validate:"required"`
	UnitId    uint   `gorm:"index:idx_product_unit_type,unique" json:"unitId" validate:"required"`
	PriceType string `gorm:"index:idx_product_unit_type,unique" json:"priceType" validate:"required,min=1"` // "BUY" or "SELL"
	UnitPrice int    `json:"price" validate:"required,min=1"`
	Remark    string `json:"remark"`
}

type ProductPriceHistory struct {
	gorm.Model
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductId     string `json:"productId"`
	ProductName   string `json:"productName"`
	UnitId        uint   `json:"unitId" `
	UnitName      string `json:"unitName" `
	PriceType     string `json:"priceType" ` // "BUY"	or "SELL"
	UnitPrice     int    `json:"price" `
	Remark        string `json:"remark"`
	EffectiveDate string `gorm:"not null"`
	CreatedAt     time.Time
}

type ProductStock struct {
	gorm.Model
	// ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductId    string `gorm:"type:varchar(20)" json:"productId"`
	BaseUnitId   int    `json:"baseUnitId" validate:"required"`
	DeriveUnitId int    `json:"deriveUnitId" validate:"required"`
	BaseQty      int    `json:"baseQty" validate:"required,min=1"`
	DerivedQty   int    `json:"derivedQty" validate:"required,min=1"`
	ReorderLvl   int    `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
}

type UnitConversion struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Description  string `gorm:"type:varchar(20)" json:"description"`
	ProductId    string `gorm:"type:varchar(20)" json:"productId" validate:"required"`
	BaseUnit     string `json:"baseUnit"`
	DeriveUnit   string `json:"deriveUnit" `
	BaseUnitId   int    `json:"baseUnitId" validate:"required"`
	DeriveUnitId int    `json:"deriveUnitId" validate:"required"`
	Factor       int    `json:"factor" validate:"required,min=1"`
}

type Inventory struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey:autoIncrement" json:"id"`
	OutQty    int       `json:"inQty"`
	InQty     int       `json:"outQty"`
	ProductId string    `json:"productId"`
	Product   Product   `gorm:"foreignKey:ProductId;" json:"product"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime" json:"updatedTime"`
}

type ItemTransaction struct {
	gorm.Model
	// TODO to enhance with UUID
	ID          uint      `gorm:"primaryKey:autoIncrement" json:"id"`
	ProductId   string    `json:"productId"`
	ReferenceNo string    `json:"referenceNo"`
	InQty       int       `json:"inQty"`
	OutQty      int       `json:"outQty"`
	Uom         string    `json:"uom"`
	TranType    string    `json:"tranType"`
	Remark      string    `json:"remark"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdTime"`
}

type Role string

const (
	ADMIN Role = "admin"
	USER  Role = "user"
)

type User struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string `gorm:"uniqueIndex;" json:"email" validate:"required,email"`
	UserName  string `json:"userName" validate:"required,min=3"`
	Password  string `json:"password" validate:"required,min=3"`
	IsAdmin   bool   `json:"isAdmin" validate:"required"`
	Role      Role   `json:"role" validate:"required" gorm:"default:user"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"-"`
}

type Customer struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	Name      string `json:"name" validate:"required,min=3"`
	Address   string `json:"address" validate:"required,min=3"`
	Phone     string `json:"phone" validate:"required,min=3"`
	Sales     []Sale `gorm:"foreignKey:CustomerId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sales"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"-"`
}

type Supplier struct {
	gorm.Model
	ID        uint       `gorm:"primaryKey:autoIncrement" json:"id"`
	Name      string     `json:"name" validate:"required,min=3"`
	Address   string     `json:"address" validate:"required,min=3"`
	Phone     string     `json:"phone" validate:"required,min=3"`
	Purchases []Purchase `gorm:"foreignKey:SupplierId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchases"`
	CreatedAt int64      `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64      `gorm:"autoUpdateTime:milli" json:"-"`
}

type Purchase struct {
	gorm.Model
	ID         string    `gorm:"primaryKey" json:"id"`
	SupplierId uint      `json:"supplierId"`
	Supplier   *Supplier `json:"supplier"`
	// Supplier        Supplier         `gorm:"foreignKey:SupplierId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"supplier"`
	PurchaseDetails []PurchaseDetail `gorm:"foreignKey:PurchaseId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchaseDetails"`
	Discount        int              `json:"discount"`
	Total           int              `json:"total"`
	GrandTotal      int              `json:"grandTotal"`
	Remark          string           `json:"remark"`
	PurchaseDate    string           `json:"purchaseDate"`
	CreatedAt       int64            `gorm:"autoCreateTime" json:"-"`
	UpdatedAt       int64            `gorm:"autoUpdateTime:milli" json:"-"`
}

// type PurchaseDetail struct {
// 	// gorm.Model includes ID, CreatedAt, UpdatedAt, DeletedAt
// 	// so you don't need to redefine ID again
// 	gorm.Model

// 	ProductId     string      `gorm:"type:varchar(20)" json:"productId"`
// 	ProductName   string      `json:"productName"`
// 	Uom           string      `json:"uom"`
// 	Qty           int         `json:"qty"`
// 	Price         int         `json:"price"`
// 	UnitId        uint        `json:"unitId"`
// 	UnitName      string      `json:"unitName"`
// 	Total         int         `json:"total"`
// 	BaseQty       int         `json:"baseQty"`
// 	PurchaseId    string      `json:"purchaseId"`
// 	ProductUnitId string      `json:"productUnitId"`
// 	ProductUnit   ProductUnit `gorm:"foreignKey:ProductUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productUnit"`
// 	Product       Product     `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
// }

type PurchaseDetail struct {
	gorm.Model
	ID            uint        `gorm:"primaryKey:autoIncrement" json:"-"`
	ProductId     string      `gorm:"type:varchar(20)" json:"productId"`
	ProductName   string      `json:"productName"`
	Uom           string      `json:"uom"`
	Qty           int         `json:"qty"`
	Price         int         `json:"price"`
	UnitId        uint        `json:"unitId"`
	UnitName      string      `json:"unitName"`
	Total         int         `json:"total"`
	BaseQty       int         `json:"baseQty"` // Converted to smallest/base unit
	PurchaseId    string      `json:"purchaseId"`
	ProductUnitId string      `json:"productUnitId"`
	ProductUnit   ProductUnit `gorm:"foreignKey:ProductUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productUnit"`
	Product       Product     `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
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
	// DerivedQty    int         `json:"derivedQty"`
	Uom           string      `json:"uom"`
	Price         int64       `json:"price"`
	Total         int64       `json:"total"`
	SaleId        string      `json:"saleId"`
	ProductUnitId string      `json:"productUnitId"`
	ProductUnit   ProductUnit `gorm:"foreignKey:ProductUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productUnit"`
	Product       Product     `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
}

type UnitOfMeasure struct {
	gorm.Model
	ID             uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	UnitName       string           `json:"unitName" validate:"required,min=3"`
	Product        []Product        `gorm:"foreignKey:UomId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"products"`
	UnitConversion []UnitConversion `gorm:"foreignKey:BaseUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"unitConversions"`
	ProductUnit    []ProductUnit    `gorm:"foreignKey:UnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productUnits"`
}

type ProductUnit struct {
	gorm.Model
	ID               string        `gorm:"primaryKey" json:"id"`
	ProductId        string        `gorm:"type:varchar(20)" json:"productId" validate:"required"`
	UnitId           uint          `gorm:"type:int" json:"unitId" validate:"required"`
	ConversionToBase int           `json:"conversionToBase" validate:"required,min=1"`
	IsDefaultUnit    bool          `json:"isDefaultUnit" validate:"required"`
	Product          Product       `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
	UnitOfMeasure    UnitOfMeasure `gorm:"foreignKey:UnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"unitOfMeasure"`
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
	err := db.AutoMigrate(
		&Category{},
		&Product{},
		&UnitOfMeasure{},
		&UnitConversion{},
		&ProductPrice{},
		&ProductStock{},
		&User{},
		&Customer{},
		&Supplier{},
		&Sale{},
		&SaleDetail{},
		&Purchase{},
		&PurchaseDetail{},
		&ItemTransaction{},
		&User{},
	)
	return err
}
