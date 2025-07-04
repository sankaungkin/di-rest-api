package product

type CreateProductRequstDTO struct {
	ID              string `gorm:"primaryKey" json:"id"`
	ProductName     string `json:"productName" validate:"required,min=3"`
	CategoryId      uint   `json:"categoryId" validate:"required"`
	Uom             string `json:"uom"`
	UomId           uint   `json:"uomId" validate:"required"`
	DeriveUom       string `json:"deriveUom"`
	DeriveUomId     uint   `json:"deriveUomId" validate:"required"`
	BuyPrice        int64  `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64  `json:"sellPriceLevel1" validate:"required,min=1"`
	DeriveUnitPrice int64  `json:"deriveUnitPrice" validate:"required,min=1"`
	ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	QtyOnHand       int    `json:"qtyOhHand" validate:"required"`
	BrandName       string `json:"brandName"`
	IsActive        bool   `json:"isActive" gorm:"default:true"`
}

type UpdateProductRequstDTO struct {
	ProductName string `json:"productName" validate:"required,min=3"`
	CategoryId  uint   `json:"categoryId" validate:"required"`
	UomId       uint   `json:"uomId" validate:"required"`
	// Uom             string `json:"uom" validate:"required,min=2"`
	BuyPrice        int64 `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64 `json:"sellPricelvl1" validate:"required,min=1"`
	DeriveUnitPrice int64 `json:"deriveUnitPrice" validate:"required,min=1"`
	// ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	// QtyOnHand       int    `json:"qtyOhHand" validate:"required"`
	BrandName string `json:"brandName"`
	IsActive  bool   `json:"isActive" gorm:"default:true"`
}

type ResponseProductDTO struct {
	ID          string `gorm:"primaryKey" json:"id"`
	ProductName string `json:"productName" validate:"required,min=3"`
	CategoryId  uint   `json:"categoryId"`
	// Uom             string `json:"uom" validate:"required,min=3"`
	BaseUnit        string `json:"baseUnit"`
	DeriveUnit      string `json:"deriveUnit"`
	UomId           uint   `json:"uomId"`
	DeriveUomId     uint   `json:"deriveUomId"`
	BuyPrice        int64  `json:"buyPrice" validate:"required,min=1"`
	SellPriceLevel1 int64  `json:"sellPricelvl1" `
	DeriveUnitPrice int64  `json:"deriveUnitPrice"`
	ReorderLvl      uint   `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	QtyOnHand       int    `json:"qtyOnHand" validate:"required"`
	BrandName       string `json:"brandName"`
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
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	BaseQty     int    `json:"baseQty"`
	DerivedQty  int    `json:"derivedQty"`
	ReorderLvl  int    `json:"reorderlvl"`
	Factor      int    `json:"factor"`
	BaseUnit    string `json:"baseUnit"`
	DeriveUnit  string `json:"deriveUnit"`
}

type UpdateProductStockDTO struct {
	ProductID  string `json:"productId"`
	BaseQty    int    `json:"baseQty"`
	DerivedQty int    `json:"derivedQty"`
	Reorder    int    `json:"reorder"`
}

type UnitConversionWithProductDTO struct {
	ID          uint   `json:"id"`
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	Description string `json:"description"`
	BaseUnit    string `json:"baseUnit"`
	DeriveUnit  string `json:"deriveUnit"`
	Factor      int    `json:"factor"`
}

type UpdateUnitConversionRequestDTO struct {
	ID           uint   `json:"id"`
	ProductId    string `json:"productId"`
	BaseUnit     string `json:"baseUnit"`
	DeriveUnit   string `json:"deriveUnit"`
	BaseUnitId   int    `json:"baseUnitId"`
	DeriveUnitId int    `json:"deriveUnitId"`
	Factor       int    `json:"factor"`
	Description  string `json:"description"`
}

type UpdateUnitRequstDTO struct {
	ID       uint   `json:"id"`
	UnitName string `json:"unitName"`
}
