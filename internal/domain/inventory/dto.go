package inventory

type IncreaseInventoryDTO struct {

	OutQty    uint   `json:"outQty"`
	InQty     uint   `json:"inQty"`
	ProductId string `json:"productId"`
	Remark    string `json:"remark"`

}