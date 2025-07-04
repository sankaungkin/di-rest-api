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
	Products     []Product `gorm:"foreignKey:CategoryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt    int64     `gorm:"autoCreateTime" json:"-"`
	UpdatedAt    int64     `gorm:"autoUpdateTime:milli" json:"-"`
}

type Product struct {
	gorm.Model
	ID               string            `gorm:"primaryKey" json:"id"`
	ProductName      string            `json:"productName" validate:"required,min=3"`
	CategoryId       uint              `json:"categoryId"`
	UnitConversion   []UnitConversion  `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Inventories      []Inventory       `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	SaleDetail       []SaleDetail      `gorm:"foreignKey:ProductId;" json:"-"`
	PurchaseDetail   []PurchaseDetail  `gorm:"foreignKey:ProductId;" json:"-"`
	ItemTransactions []ItemTransaction `gorm:"foreignKey:ProductId;"  json:"-"`
	Uom              string            `json:"uom"`
	DeriveUom        string            `json:"deriveUom"`
	UomId            uint              `json:"uomId" validate:"required"`
	DeriveUomId      uint              `json:"deriveUomId" validate:"required"`
	BuyPrice         int64             `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1  int64             `json:"sellPricelvl1" validate:"required,min=1"`
	DeriveUnitPrice  int64             `json:"deriveUnitPrice" validate:"required,min=1"`
	BrandName        string            `json:"brandName"`
	IsActive         bool              `json:"isActive" gorm:"default:true"`
	CreatedAt        int64             `gorm:"autoCreateTime" json:"-"`
	UpdatedAt        int64             `gorm:"autoUpdateTime:milli" json:"-"`
}

type UnitOfMeasure struct {
	gorm.Model
	ID             uint             `gorm:"primaryKey" json:"id"`
	UnitName       string           `json:"unitName" validate:"required,min=3"`
	Product        []Product        `gorm:"foreignKey:UomId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	UnitConversion []UnitConversion `gorm:"foreignKey:BaseUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type ProductPrice struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey" json:"id"`
	ProductId string `gorm:"index:idx_product_unit_type,unique" json:"productId" validate:"required"`
	UnitId    uint   `gorm:"index:idx_product_unit_type,unique" json:"unitId" validate:"required"`
	PriceType string `gorm:"index:idx_product_unit_type,unique" json:"priceType" validate:"required,min=1"` // "BUY" or "SELL"
	UnitPrice int64  `json:"price" validate:"required,min=1"`
}

type ProductPriceHistory struct {
	gorm.Model
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProductId     string    `json:"productId" validate:"required"`
	UnitId        uint      `json:"unitId" validate:"required"`
	PriceType     string    `json:"priceType" validate:"required,min=1"` // "BUY"	or "SELL"
	UnitPrice     int64     `json:"price" validate:"required,min=1"`
	EffectiveDate time.Time `gorm:"not null"`
	CreatedAt     time.Time
}

type ProductStock struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey" json:"id"`
	ProductId    string `gorm:"type:varchar(20)" json:"productId"`
	BaseUnitId   int    `json:"baseUnitId" validate:"required"`
	DeriveUnitId int    `json:"deriveUnitId" validate:"required"`
	BaseQty      int    `json:"baseQty" validate:"required,min=1"`
	DerivedQty   int    `json:"derivedQty" validate:"required,min=1"`
	ReorderLvl   int    `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
}

type UnitConversion struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey" json:"id"`
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
	Sales     []Sale `gorm:"foreignKey:CustomerId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli" json:"-"`
}

type Supplier struct {
	gorm.Model
	ID        uint       `gorm:"primaryKey:autoIncrement" json:"id"`
	Name      string     `json:"name" validate:"required,min=3"`
	Address   string     `json:"address" validate:"required,min=3"`
	Phone     string     `json:"phone" validate:"required,min=3"`
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
	ProductId   string `gorm:"type:varchar(20)" json:"productId"`
	ProductName string `json:"productName"`
	Qty         int    `json:"qty"`
	Price       int64  `json:"price"`
	UnitName    string `json:"unitName"`
	Total       int64  `json:"total"`
	PurchaseId  string `json:"purchaseId"`
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
	DerivedQty  int    `json:"derivedQty"`
	Uom         string `json:"uom"`
	Price       int64  `json:"price"`
	Total       int64  `json:"total"`
	SaleId      string `json:"saleId"`
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
