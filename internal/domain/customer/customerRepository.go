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

	err := r.db.Create(&customer).Error
	return customer, err
	// input := new(CreateCustomerRequestDTO)
	// newCustomer := &models.Customer{
	// 	Name: input.Name,
	// 	Address: input.Address,
	// 	Phone: input.Phone,
	// }

	// err := r.db.Create(newCustomer)
	// if err != nil {
	// 	return nil, err.Error
	// }

	// return newCustomer, nil
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


	func(r *CustomerRepository)UpdateCustomer(input *models.Customer) (*models.Customer, error){
		
		var existingCustomer *models.Customer
		err := r.db.Where("id = ?", input.ID).First(&existingCustomer).Error
		if err != nil {
			// Handle error if customer not found or other issue
			return nil, err
		}

		log.Println("input: ", input)
		if input.Address == "" || input.Name == "" || input.Phone =="" {
			return nil, err
		}
		// Update relevant fields from input data
		existingCustomer.Name = input.Name  // Update other fields as needed
		existingCustomer.Address = input.Address
		existingCustomer.Phone =input.Phone

		// Save the updated customer data
		log.Println("existingCustomer: ", existingCustomer)
		err = r.db.Updates(&existingCustomer).Error
		if err != nil {
			// Handle error if update fails
			return nil, err
		}

		// Return the updated customer object
		return existingCustomer, nil
	}

	func(r *CustomerRepository) DeleteCustomer (id uint) error {
		// return r.db.Delete(&User{}, id).Error
		return r.db.Delete(&models.Customer{}, id).Error
	}