package productprice

type ProductPriceResponseDTO struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	ProductId     string `json:"productId" `
	ProductUnitId string `json:"productUnitId" `
	ProductName   string `json:"productName"`
	UnitId        uint   `json:"unitId" `
	UnitName      string `json:"unitName"`
	UnitPrice     int    `json:"price" `
	PriceType     string `json:"priceType"`
}

type UpdateProductPriceRequestDTO struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ProductId   string `json:"productId" `
	ProductName string `json:"productName"`
	UnitId      uint   `json:"unitId" `
	UnitName    string `json:"unitName"`
	UnitPrice   int    `json:"price" `
	PriceType   string `json:"priceType"`
}

type UpdateProductPriceDTO struct {
	ProductId     string `json:"productId" `
	ProductUnitId string `json:"productUnitId"`
	PriceType     string `json:"priceType"`
	Price         int    `json:"price" `
}
