package product

type CreateProductRequstDTO struct {
	ID              string `gorm:"primaryKey" json:"id"`
	ProductName     string `json:"productName" validate:"required,min=3"`
	CategoryId      uint   `json:"categoryId" validate:"required"`
	Uom             string `json:"uom" validate:"required,min=3"`
	BuyPrice        int64  `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64  `json:"sellPriceLevel1" validate:"required,min=1"`
	SellPriceLevel2 int64  `json:"sellPriceLevel2" validate:"required,min=1"`
	ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	QtyOnHand       int    `json:"qtyOhHand" validate:"required"`
	BrandName       string `json:"brand"`
	IsActive        bool   `json:"isActive" gorm:"default:true"`
}

type UpdateProductRequstDTO struct {
	ProductName     string `json:"productName" validate:"required,min=3"`
	CategoryId      uint   `json:"categoryId" validate:"required"`
	Uom             string `json:"uom" validate:"required,min=2"`
	BuyPrice        int64  `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64  `json:"sellPricelvl1" validate:"required,min=1"`
	SellPriceLevel2 int64  `json:"sellPricelvl2"`
	// ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	// QtyOnHand       int    `json:"qtyOhHand" validate:"required"`
	BrandName string `json:"brandName"`
	IsActive  bool   `json:"isActive" gorm:"default:true"`
}

type ResponseProductDTO struct {
	ID              string `gorm:"primaryKey" json:"id"`
	ProductName     string `json:"productName" validate:"required,min=3"`
	CategoryId      uint   `json:"categoryId"`
	Uom             string `json:"uom" validate:"required,min=3"`
	BuyPrice        int64  `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64  `json:"sellPricelvl1" validate:"required,min=1"`
	SellPriceLevel2 int64  `json:"sellPricelvl2" validate:"required,min=1"`
	ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	QtyOnHand       int    `json:"qtyOnHand" validate:"required"`
	BrandName       string `json:"brand"`
	IsActive        bool   `json:"isActive" gorm:"default:true"`
	CreatedAt       string `json:"createdAt"`
}

type ResponseProductUnitPriceDTO struct {
	Serial      int    `json:"serial"`
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	Uom         string `json:"uom"`
	UnitPrice   int64  `json:"unitPrice"`
}

type ResponseProductStockDTO struct {
	ProductID         string `json:"productId"`
	ProductName       string `json:"productName"`
	BaseUomInStock    int    `json:"baseUomInStock"`
	DerivedUomInStock int    `json:"derivedUomInStock"`
	Reorder           int    `json:"reorder"`
}
