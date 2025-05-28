package inventory

type IncreaseInventoryDTO struct {
	OutQty    int    `json:"outQty"`
	InQty     int    `json:"inQty"`
	ProductId string `json:"productId"`
	Remark    string `json:"remark"`
}

type ResponseInventoryDTO struct {
	ProductName string `json:"productName"`
	OutQty      int    `json:"outQty"`
	InQty       int    `json:"inQty"`
	TranType    string `json:"tranType"`
	ProductId   string `json:"productId"`
	Remark      string `json:"remark"`
	QtyOnHand   int    `json:"qtyOnHand"`
	CreatedAt   string `json:"createdAt"`
}
