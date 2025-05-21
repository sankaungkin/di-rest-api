package itemtransactions

type ItemTransactionDTO struct {
	ID          string `gorm:"primaryKey" json:"id"`
	ProductId   string `json:"productId"`
	ReferenceNo string `json:"referenceNo"`
	OutQty      uint   `json:"inQty"`
	InQty       uint   `json:"outQty"`
	TranType    string `json:"tranType"`
	Remark      string `json:"remark"`
}
