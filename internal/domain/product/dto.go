package product

type CreateProductRequstDTO struct {
	ID          string `gorm:"primaryKey" json:"id"`
	ProductName string `json:"productName" validate:"required,min=3"`
	CategoryId  uint   `json:"categoryId" validate:"required"`
	Uom         string `json:"uom"`
	UomId       uint   `json:"uomId" validate:"required"`
	DeriveUom   string `json:"deriveUom"`
	DeriveUomId uint   `json:"deriveUomId" gorm:"default:1"`
	BrandName   string `json:"brandName"`
	IsActive    bool   `json:"isActive" gorm:"default:true"`
	Facor       int    `json:"factor" gorm:"default:1"`
}

type Create_Product_UnitConversion_Stock_Price_DTO struct {
	//Product DTO
	ID           string `gorm:"primaryKey" json:"id"`
	ProductName  string `json:"productName" validate:"required,min=3"`
	CategoryId   uint   `json:"categoryId" validate:"required"`
	BrandName    string `json:"brandName"`
	IsActive     bool   `json:"isActive" gorm:"default:true"`
	ReorderLvl   int    `json:"reorderlvl" gorm:"default:1" validate:"required,min=1"`
	DeriveUnitId int    `json:"deriveUnitId" validate:"required,min=1"`
	Qty          int    `json:"qty" validate:"required,min=1"`
	//Product Units array DTO
	ProductUnits []ProductUnit `json:"productUnits"`
	//Product Prices array DTO
	ProductPrices []ProductPrice `json:"productPrices"`
}

type UpdateProductUnitDTO struct {
	ProductId    string                  `json:"productId" validate:"required"`
	ProductUnits []ProductUnitRequestDTO `json:"productUnits"`
}
type ProductUnitRequestDTO struct {
	ID               string `json:"id"`
	UnitID           uint   `json:"unitId"`
	ConversionToBase int    `json:"conversionToBase"`
	IsDefaultUnit    bool   `json:"isDefaultUnit"`
	ProductID        string `json:"productId"`
}

type UpdateProductRequstDTO struct {
	ProductId   string `json:"productId" validate:"required"`
	ProductName string `json:"productName" validate:"required,min=3"`
	CategoryId  uint   `json:"categoryId" validate:"required"`
	BrandName   string `json:"brandName"`
	IsActive    bool   `json:"isActive" gorm:"default:true"`
	// ProductUnits []ProductUnit `json:"productUnits" gorm:"many2many:product_units;"`
}

type ProductUnit struct {
	Id               string `json:"productUnitId" `
	ProductId        string `json:"productId" `
	UnitId           uint   `json:"unitId" `
	ConversionToBase int    `json:"conversionToBase" gorm:"default:1"`
	IsDefaultUnit    bool   `json:"isDefaultUnit" gorm:"default:false"`
}

type ProductPrice struct {
	ProductUnitId string `json:"productUnitId" `
	ProductId     string `json:"productId" `
	UnitId        uint   `json:"unitId" `
	PriceType     string `json:"priceType" ` // "BUY" or "SELL"
	UnitPrice     int    `json:"unitPrice" `
	Remark        string `json:"remark"`
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

//SELECT product_id, p.product_name, uom.unit_name, unit_id, unit_price, effective_date, price_type

type ResponseProductHistoryDTO struct {
	ProductId     string `json:"productId"`
	ProductName   string `json:"productName"`
	UnitId        uint   `json:"unitId"`
	UnitName      string `json:"unitName"`
	PriceType     string `json:"priceType"`
	UnitPrice     int64  `json:"unitPrice"`
	EffectiveDate string `json:"effectiveDate"`
}
