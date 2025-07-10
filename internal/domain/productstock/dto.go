package productstock

type ResponseProductStockDTO struct {
	ProductID    string `json:"productId"`
	ProductName  string `json:"productName"`
	BaseQty      int    `json:"baseQty"`
	DerivedQty   int    `json:"derivedQty"`
	ReorderLvl   int    `json:"reorderlvl"`
	BaseUnitId   int    `json:"baseUnitId"`
	DeriveUnitId int    `json:"deriveUnitId"`
}

type UpdateProductStockDTO struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ProductID  string `json:"productId"`
	BaseQty    int    `json:"baseQty"`
	DerivedQty int    `json:"derivedQty"`
	ReorderLvl int    `json:"reorderlvl"`
}
