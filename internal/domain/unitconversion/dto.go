package unitconversion

type CreateUnitConversionDTO struct {
	ProductId    string `json:"productId"`
	BaseUnitId   int    `json:"baseUnitId"`
	DeriveUnitId int    `json:"deriveUnitId"`
	Factor       int    `json:"factor"`
	Description  string `json:"description"`
}

type UpdateUnitConversionRequestDTO struct {
	ID           uint   `json:"id"`
	ProductId    string `json:"productId"`
	BaseUnitId   int    `json:"baseUnitId"`
	DeriveUnitId int    `json:"deriveUnitId"`
	Factor       int    `json:"factor"`
	Description  string `json:"description"`
}
