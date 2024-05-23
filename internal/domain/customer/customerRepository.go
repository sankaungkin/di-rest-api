package customer

import (
	"errors"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CustomerRepositoryInterface interface{
	CreateCustomer(customer *models.Customer) (*models.Customer, error)
	GetAllCustomers() ([]models.Customer, error)
	GetCustomerById(id uint) (*models.Customer, error)
	UpdateCustomer(customer *models.Customer) (*models.Customer, error)
	DeleteCustomer(id uint) error
}

type CustomerRepository struct {
	db *gorm.DB
}


var(
	repoInstance *CustomerRepository
	repoOnce sync.Once
	Reset = "\033[0m" 
	Magenta = "\033[35m"
)

// func NewCustomerRepository(db *gorm.DB) CustomerRepositoryInterface{
// 	return &CustomerRepository{db: db}
// }

func NewCustomerRepository(db *gorm.DB) CustomerRepositoryInterface{
	log.Println(Magenta + "CustomerRepository constructor is called" +Reset)
	repoOnce.Do(func(){
		repoInstance = &CustomerRepository{db: db}
	})
	return repoInstance
	// return &CustomerRepository{db: db}
}

func(r *CustomerRepository)CreateCustomer(customer *models.Customer) (*models.Customer, error){
	input := new(CreateCustomerRequestDTO)
	newCustomer := &models.Customer{
		Name: input.Name,
		Address: input.Address,
		Phone: input.Phone,
	}

	err := r.db.Create(newCustomer)
	if err != nil {
		return nil, err.Error
	}

	return newCustomer, nil
}

	func(r *CustomerRepository)GetAllCustomers() ([]models.Customer, error){
		customers := []models.Customer{}
		r.db.Model(&models.Customer{}).Order("ID asc").Find(&customers)
		if len(customers) == 0 {
			return nil, errors.New("NO records found")
		}
		return customers, nil
	}

	func(r *CustomerRepository)GetCustomerById(id uint) (*models.Customer, error){
		var customer models.Customer
		result := r.db.First(&customer, "id = ?",id)
		if err := result.Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, err
			}
		}
		return &customer, nil
	}


	func(r *CustomerRepository)UpdateCustomer(customer *models.Customer) (*models.Customer, error){
		var updateCustomer *models.Customer
		err := r.db.First(&updateCustomer,"id = ?",customer.ID)
		if err != nil {
			return nil, err.Error
		}
		r.db.Save(&updateCustomer)
		return updateCustomer, nil
	}

	func(r *CustomerRepository)DeleteCustomer(id uint) error {
		var deleteCustomer *models.Customer
		err := r.db.First(&deleteCustomer, "id = ?", id)
		if err != nil {
			return err.Error
		}
		r.db.Delete(&deleteCustomer)
		return nil
	}