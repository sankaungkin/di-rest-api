package inventory

type IncreaseInventoryDTO struct {
	OutQty    uint   `json:"outQty"`
	InQty     uint   `json:"inQty"`
	ProductId string `json:"productId"`
	Remark    string `json:"remark"`
}

type ResponseInventoryDTO struct {
	ProductName string `json:"productName"`
	OutQty      uint   `json:"outQty"`
	InQty       uint   `json:"inQty"`
	TranType    string `json:"tranType"`
	ProductId   string `json:"productId"`
	Remark      string `json:"remark"`
	QtyOnHand   int    `json:"qtyOnHand"`
	CreatedAt   string `json:"createdAt"`
}
