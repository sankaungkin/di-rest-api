package customer

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type CustomerServiceInterface interface {
	CreateCustomer(customer *models.Customer) (*models.Customer, error)
	GetAllCustomers() ([]models.Customer, error)
	GetCustomerById(id uint) (*models.Customer, error)
	UpdateCustomer(customer *models.Customer) (*models.Customer, error)
	DeleteCustomer(id uint) error
}

type CustomerService struct {
	repo CustomerRepositoryInterface
}

//! for singletone pattern
var (
	svcInstance *CustomerService
	svcOnce sync.Once
)

// func NewCustomerService(repo CustomerRepositoryInterface) CustomerServiceInterface{
// 	return &CustomerService{repo:repo}
// }

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail
func NewCustomerService(repo CustomerRepositoryInterface) CustomerServiceInterface {
	log.Println(util.Gray + "CustomerService constructor is called " + util.Reset)
	svcOnce.Do(func() {
		svcInstance = &CustomerService{repo: repo}
	})
	return svcInstance
}

func (s *CustomerService)CreateCustomer(customer *models.Customer) (*models.Customer, error){
	return s.repo.CreateCustomer(customer)
}



func (s *CustomerService)GetAllCustomers() ([]models.Customer, error){
	return s.repo.GetAllCustomers()
}

func (s *CustomerService)GetCustomerById(id uint) (*models.Customer, error){
	return s.repo.GetCustomerById(id)
}

func (s *CustomerService)UpdateCustomer(customer *models.Customer) (*models.Customer, error){
	return s.repo.UpdateCustomer(customer)
}


func (s *CustomerService)DeleteCustomer(id uint) error{
	return s.repo.DeleteCustomer(id)
}
