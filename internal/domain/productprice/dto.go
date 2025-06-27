package productprice

type ProductPriceResponseDTO struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	ProductId   string `json:"productId" `
	ProductName string `json:"productName"`
	UnitId      uint   `json:"unitId" `
	UnitName    string `json:"unitName"`
	UnitPrice   int64  `json:"price" `
}
