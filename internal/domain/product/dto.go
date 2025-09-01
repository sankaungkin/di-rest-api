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
	ID          string `gorm:"primaryKey" json:"id"`
	ProductName string `json:"productName" validate:"required,min=3"`
	CategoryId  uint   `json:"categoryId" validate:"required"`
	Uom         string `json:"uom"`
	UomId       uint   `json:"uomId" validate:"required"`
	DeriveUom   string `json:"deriveUom"`
	DeriveUomId uint   `json:"deriveUomId" gorm:"default:1"`
	BrandName   string `json:"brandName"`
	IsActive    bool   `json:"isActive" gorm:"default:true"`

	//Product Unit Conversion DTO
	/**
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
	*/
	BaseUnitName   string `json:"baseUnitName"`
	BaseUnitId     uint   `json:"baseUnitId" validate:"required"`
	DeriveUnitName string `json:"deriveUnitName"`
	DeriveUnitId   uint   `json:"deriveUnitId" validate:"required"`
	Description    string `json:"description"`
	Factor         int    `json:"factor" gorm:"default:1"`

	//Product Price DTO
	/**
	type ProductPrice struct {
		gorm.Model
		ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
		ProductId string `gorm:"index:idx_product_unit_type,unique" json:"productId" validate:"required"`
		UnitId    uint   `gorm:"index:idx_product_unit_type,unique" json:"unitId" validate:"required"`
		PriceType string `gorm:"index:idx_product_unit_type,unique" json:"priceType" validate:"required,min=1"` // "BUY" or "SELL"
		UnitPrice int64  `json:"price" validate:"required,min=1"`
		Remark    string `json:"remark"`
	}
	*/

	// PriceType string `json:"priceType" gorm:"default:BUY"`
	// Price     int64  `json:"price"`

	Prices []struct {
		PriceType string `json:"priceType" validate:"required,oneof=BUY SELL"`
		UnitId    uint   `json:"unitId" validate:"required"`
		Price     int    `json:"price" validate:"required,min=1"`
	} `json:"prices" validate:"required,min=1"`

	//Inventory DTO
	// BasesUnitId uint `json:"baseUnitId"`
	// DerivedUnitId uint `json:"deriveUnitId"`
	/**
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
	*/
	BaseQty    int `json:"baseQty"`
	DerivedQty int `json:"derivedQty"`
	ReorderLvl int `json:"reorderlvl" gorm:"default:5"`
}

type UpdateProductRequstDTO struct {
	ProductName string `json:"productName" validate:"required,min=3"`
	CategoryId  uint   `json:"categoryId" validate:"required"`
	UomId       uint   `json:"uomId" `
	Uom         string `json:"uom" validate:"required,min=2"`
	DeriveUom   string `json:"deriveUom" `
	DeriveUomId uint   `json:"deriveUomId" gorm:"default:1"`
	// UomId          uint   `json:"uomId"
	// BuyPrice        int64 `json:"buyPrice" `
	// SellPriceLevel1 int64 `json:"sellPricelvl1" `
	// DeriveUnitPrice int64 `json:"deriveUnitPrice" `
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
