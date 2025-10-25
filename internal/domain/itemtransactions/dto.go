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

type ResquestAdjustInventoryDTO struct {
	ProductId   string `json:"productId"`
	BaseQty     int    `json:"baseQty"`
	DerivedQty  int    `json:"derivedQty"`
	InQty       int    `json:"inQty"`  // Quantity to be added
	OutQty      int    `json:"outQty"` // Quantity to be removed
	Uom         string `json:"uom"`    // Unit of Measure (e.g., EACH, KG)
	Remark      string `json:"remark"`
	TranType    string `json:"tranType"`    // DEBIT or CREDIT
	ReferenceNo string `json:"referenceNo"` // Reference number for the transaction
	CreatedAt   string `json:"createdAt"`   // Timestamp of the transaction
}

type ResponseItemTransactionDTO struct {
	ProductId   string `json:"productId"`
	ProductName string `json:"productName"`
	ReferenceNo string `json:"referenceNo"`
	InQty       int    `json:"inQty"`
	OutQty      int    `json:"outQty"`
	Uom         string `json:"uom"`
	TranType    string `json:"tranType"`
	// CreatedAt   time.Time `json:"createdAt"`
	CreatedAt string `json:"createdAt"`
	Remark    string `json:"remark"`
}
