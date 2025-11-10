package productstock

type ProductStockListInfoWithCategory struct {
	ProductId      string `json:"productId"`
	ProductName    string `json:"productName"`
	CategoryName   string `json:"categoryName"`
	UomId          int    `json:"uomId"`
	QuantityOnHand int    `json:"quantityOnHand"`
	ReorderLvl     int    `json:"reorderlvl"`
	Remark         string `json:"remark"`
}

type ResponseProductStockDTO struct {
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	// BaseQty      int    `json:"baseQty"`
	DerivedQty int `json:"derivedQty"`
	ReorderLvl int `json:"reorderlvl"`
	// BaseUnitId   int    `json:"baseUnitId"`
	DeriveUnitId int `json:"deriveUnitId"`
}

type OutOfStockDTO struct {
	ProductId      string `json:"productId"`
	ProductName    string `json:"productName"`
	QuantityOnHand int    `json:"quantityOnHand"`
	UnitName       string `json:"unitName"`
	ReorderLvl     int    `json:"reorderLvl"`
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

type UpdateProductStockDTO struct {
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	DerivedQty  int    `json:"derivedQty"`
	ReorderLvl  int    `json:"reorderlvl"`
	Remark      string `json:"remark"`
}

type DisplayStock struct {
	UnitName string `json:"unitName"`
	Quantity int    `json:"quantity"`
}

type StockResponse struct {
	ProductId   string         `json:"productId"`
	ProductName string         `json:"productName"`
	Quantity    int            `json:"quantity"`
	ReorderLvl  int            `json:"reorderlvl"`
	Units       []DisplayStock `json:"units"`
	Message     string         `json:"message"`
	Status      string         `json:"status"`
}

type ConcreteBlockHead struct {
	ProductId      string `json:"productId"`
	ProductName    string `json:"productName"`
	QuantityOnHand int    `json:"quantityOnHand"`
	ReorderLvl     int    `json:"reorderLvl"`
}

type ActiveProductStockDTO struct {
	ProductId      string `json:"productId"`
	ProductName    string `json:"productName"`
	UnitName       string `json:"unitName"`
	QuantityOnHand int    `json:"quantityOnHand"`
}
