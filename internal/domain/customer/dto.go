package customer

type CreateCustomerRequestDTO struct{
	Name string `json:"name"`
	Address string `json:"address"`
	Phone string `json:"phone"`
}



type UpdateCustomerRequstDTO struct{
	Name string `json:"name"`
	Address string `json:"address"`
	Phone string `json:"phone"`
}