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

type ProductUnit struct {
	gorm.Model
	ID               string        `gorm:"primaryKey" json:"id"`
	ProductId        string        `gorm:"type:varchar(20)" json:"productId" validate:"required"`
	UnitId           uint          `gorm:"type:int" json:"unitId" validate:"required"`
	ConversionToBase int           `json:"conversionToBase" validate:"required,min=1"`
	IsDefaultUnit    bool          `json:"isDefaultUnit" validate:"required"`
	Product          Product       `gorm:"foreignKey:ProductId;references:ID" json:"product"`
	UnitOfMeasure    UnitOfMeasure `gorm:"foreignKey:UnitId;references:ID" json:"unitOfMeasure"`
}
type UnitOfMeasure struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UnitName string `json:"unitName" validate:"required,min=3"`
	// UnitConversion []UnitConversion `gorm:"foreignKey:BaseUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"unitConversions"`
	ProductUnit  []ProductUnit  `gorm:"foreignKey:UnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productUnits"`
	ProductStock []ProductStock `gorm:"foreignKey:BaseUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productStocks"`
	ProductPrice []ProductPrice `gorm:"foreignKey:UnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productPrices"`
}

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
	BrandName        string            `json:"brandName"`
	IsActive         bool              `json:"isActive" gorm:"default:true"`
	Category         Category          `gorm:"foreignKey:CategoryId;references:ID" json:"category"` // 👈 Add this
	ProductUnits     []ProductUnit     `gorm:"foreignKey:ProductId;references:ID" json:"productUnits"`
	ProductPrices    []ProductPrice    `gorm:"foreignKey:ProductId;references:ID" json:"productPrices"`
	SaleDetail       []SaleDetail      `gorm:"foreignKey:ProductId;" json:"saleDetails"`
	PurchaseDetail   []PurchaseDetail  `gorm:"foreignKey:ProductId;" json:"purchaseDetails"`
	ItemTransactions []ItemTransaction `gorm:"foreignKey:ProductId;"  json:"itemTransactions"`
	// Uom              string            `json:"uom"`
	// DeriveUom        string            `json:"deriveUom"`
	// UomId            uint              `json:"uomId"  `
	// DeriveUomId      uint              `json:"deriveUomId"`
	// BuyPrice         int64             `json:"buyPrice"`
	// SellPriceLevel1  int64             `json:"sellPricelvl1" `
	// DeriveUnitPrice  int64             `json:"deriveUnitPrice" `
	CreatedAt int64 `gorm:"autoCreateTime" json:"-"`
	UpdatedAt int64 `gorm:"autoUpdateTime:milli" json:"-"`
}

type ProductPrice struct {
	gorm.Model
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductId string `gorm:"index:idx_product_unit_type,unique" json:"productId" validate:"required"`
	// Product       Product       `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
	UnitId uint `gorm:"index:idx_product_unit_type,unique" json:"unitId" validate:"required"`
	// Unit          UnitOfMeasure `gorm:"foreignKey:UnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"unit"`
	ProductUnitId string `gorm:"index:idx_product_unit_type,unique" json:"productUnitId" validate:"required"`
	PriceType     string `gorm:"index:idx_product_unit_type,unique" json:"priceType" validate:"required,min=1"` // "BUY" or "SELL"
	UnitPrice     int    `json:"price" validate:"required,min=1"`
	Remark        string `json:"remark"`
}

type ProductPriceHistory struct {
	gorm.Model
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductId     string `json:"productId"`
	ProductName   string `json:"productName"`
	ProductUnitId string `json:"productUnitId"`
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
	ProductId     string        `gorm:"type:varchar(20)" json:"productId"`
	Product       Product       `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
	BaseUnitId    int           `json:"baseUnitId" `
	DeriveUnitId  int           `json:"deriveUnitId" validate:"required"`
	BaseQty       int           `json:"baseQty" `
	DerivedQty    int           `json:"derivedQty" validate:"required,min=1"`
	ReorderLvl    int           `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	Remark        string        `json:"remark"`
	UnitOfMeasure UnitOfMeasure `gorm:"foreignKey:BaseUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"unitOfMeasure,omitempty"`
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
	Product   Product   `json:"product"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime" json:"updatedTime"`
}

type ItemTransaction struct {
	gorm.Model
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
	ID      uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	Name    string `json:"name" validate:"required,min=3"`
	Address string `json:"address" validate:"required,min=3"`
	Phone   string `json:"phone" validate:"required,min=3"`

	// --- Refactored Fields ---
	OrderCount   int       `gorm:"default:0" json:"orderCount"`
	TotalSpent   int64     `gorm:"default:0" json:"totalSpent"`
	LastPurchase time.Time `json:"lastPurchase"`

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
	ID              string           `gorm:"primaryKey" json:"id"`
	SupplierId      uint             `json:"supplierId"`
	Supplier        *Supplier        `json:"supplier"`
	PurchaseDetails []PurchaseDetail `gorm:"foreignKey:PurchaseId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"purchaseDetails"`

	Discount   int   `json:"discount"`
	Total      int   `json:"total"`
	GrandTotal int64 `json:"grandTotal"` // The amount you actually owe

	// ✅ Added for Payable Tracking
	PaidAmount    int64  `json:"paidAmount"`    // How much you have paid so far
	BalanceAmount int64  `json:"balanceAmount"` // GrandTotal - PaidAmount
	PaymentStatus string `json:"paymentStatus"` // "DEBT", "PARTIAL", "PAID"

	Remark string `json:"remark"`

	// ✅ Note: Keep these for historical cash flow tracking
	PaymentSource   string `json:"paymentSource"` // "CASH", "OWNER", "CREDIT"
	AmountFromCash  int64  `json:"amountFromCash"`
	AmountFromOwner int64  `json:"amountFromOwner"`
	//`gorm:"type:timestamptz;index" json:"transactionDate"` `gorm:"type:timestamptz;default:now()" json:"saleDate"`
	PurchaseDate time.Time `gorm:"type:timestamptz;default:now()" json:"purchaseDate"`
	CreatedAt    int64     `gorm:"autoCreateTime" json:"-"`
	UpdatedAt    int64     `gorm:"autoUpdateTime:milli" json:"-"`
}
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
	ID            string       `gorm:"primaryKey" json:"id"`
	CustomerId    uint         `json:"customerId"`
	Customer      *Customer    `json:"customer"`
	SaleDetails   []SaleDetail `gorm:"foreignKey:SaleId;reference:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"saleDetails"`
	Discount      int64        `json:"discount"`
	Total         int64        `json:"total"`
	GrandTotal    int64        `json:"grandTotal"`
	PaymentMethod string       `validate:"required" json:"paymentMethod"` // "CASH", "KPAY", "DEBT"
	PaidAmount    int64        `json:"paidAmount"`                        // ✅ Track how much has been paid so far
	BalanceAmount int64        `json:"balanceAmount"`
	PaymentStatus string       `json:"paymentStatus"` // ✅ "UNPAID", "PARTIAL", "PAID"
	ReturnAmount  int64        `json:"returnAmount"`  // ✅ NEW: total refunded
	NetTotal      int64        `json:"netTotal"`      // ✅ NEW: grandTotal - returnAmount
	Status        string       `json:"status"`        // ✅ NEW: partial / full return
	Remark        string       `json:"remark"`
	SaleDate      time.Time    `gorm:"type:timestamptz;default:now()" json:"saleDate"`
	CreatedAt     int64        `gorm:"autoCreateTime" json:"-"`
	UpdatedAt     int64        `gorm:"autoUpdateTime:milli" json:"-"`

	PaymentRecords []PaymentRecord `gorm:"foreignKey:ReferenceID" json:"paymentRecords"`
}
type PaymentRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ReferenceID   string    `gorm:"index" json:"referenceId"` // SaleID or PurchaseID
	Amount        int64     `json:"amount"`
	PaymentDate   time.Time `json:"paymentDate"`
	PaymentMethod string    `json:"paymentMethod"` // CASH, KPAY
	Type          string    `json:"type"`          // "RECEIPT" (from customer) or "PAYMENT" (to supplier)
}
type SaleDetail struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey:autoIncrement" json:"id"`
	ProductId   string `json:"productId"`
	ProductName string `json:"productName"`
	Qty         int    `json:"qty"`
	ReturnedQty int    `json:"returnedQty"` // ✅ how many returned
	NetQty      int    `json:"netQty"`      // ✅ remaining = Qty - ReturnedQty
	Remark      string `json:"remark"`
	// DerivedQty    int         `json:"derivedQty"`
	Uom           string      `json:"uom"`
	Price         int64       `json:"price"`
	Total         int64       `json:"total"`
	SaleId        string      `json:"saleId"`
	ProductUnitId string      `json:"productUnitId"`
	ProductUnit   ProductUnit `gorm:"foreignKey:ProductUnitId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"productUnit"`
	Product       Product     `gorm:"foreignKey:ProductId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product"`
}

type SaleReturn struct {
	ID          string           `gorm:"primaryKey;size:30" json:"id"`
	SaleID      string           `json:"saleId"`
	Remark      string           `json:"remark"`
	ReturnDate  time.Time        `json:"returnDate"`
	TotalAmount int64            `json:"totalAmount"`
	Sale        Sale             `gorm:"foreignKey:SaleID"`
	ReturnItems []SaleReturnItem `gorm:"foreignKey:SaleReturnID"`
	CreatedAt   time.Time        `json:"createdAt"`
}

type SaleReturnItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SaleReturnID string    `json:"saleReturnId"`
	SaleDetailID uint      `json:"saleDetailId"`
	ProductID    string    `json:"productId"`
	Qty          int       `json:"qty"`
	UnitPrice    int64     `json:"unitPrice"`
	Total        int64     `json:"total"`
	CreatedAt    time.Time `json:"createdAt"`
}

type DailySummaries struct {
	ID uint `gorm:"primaryKey:autoIncrement" json:"id"`
	// ✅ Changed type to 'date' and added 'uniqueIndex'
	// This makes it physically impossible for two rows to exist for the same day.
	SummaryDate        time.Time `gorm:"type:date;uniqueIndex" json:"summaryDate"`
	OpeningBalance     int64     `json:"openingBalance"`
	ClosingBalance     int64     `json:"closingBalance"`
	KPayTotal          int64     `json:"kpayTotal"`
	DebtTotal          int64     `json:"debtTotal"`
	DebtCollected      int64     `json:"debtCollected"`
	ExpenseTotal       int64     `json:"expenseTotal"`
	TotalCashSales     int64     `json:"totalCashSales"`
	CashTotal          int64     `json:"cashTotal"`
	TotalSale          int64     `json:"totalSale"`
	TotalWithdrawal    int64     `json:"totalWithdrawal"`
	DebtCollectedTotal int64     `json:"debtCollectedTotal"`
	IsClosed           bool      `json:"isClosed" gorm:"default:false"`
}

type Cashbook struct {
	ID                uint      `gorm:"primaryKey:autoIncrement" json:"id"`
	TransactionDate   time.Time `gorm:"type:timestamptz;index" json:"transactionDate"`
	ReferenceID       string    `json:"referenceId"` // ID from Sale, Purchase, or Expense
	Description       string    `json:"description"`
	Debit             int64     `json:"debit"`   // Cash In (+)
	Credit            int64     `json:"credit"`  // Cash Out (-)
	Balance           int64     `json:"balance"` // Running Balance
	CreatedAt         time.Time `json:"createdAt"`
	TransactionType   string    `json:"transactionType"`                   // OPENING, CLOSING, etc.
	PaymentMethod     string    `validate:"required" json:"paymentMethod"` // CASH, KPAY, DEBT
	TransactionStatus string    `json:"transactionStatus"`
}

type Expense struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey:autoIncrement" json:"id"`
	Category    string    `json:"category"` // Rent, Electricity, etc.
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	ExpenseDate time.Time `json:"expenseDate"`
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
		&Cashbook{},
		&Expense{},
		&User{},
	)
	return err
}
