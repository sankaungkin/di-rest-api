package customer

import (
	"errors"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CustomerRepositoryInterface interface {
	CreateCustomer(customer *models.Customer) (*models.Customer, error)
	GetAllCustomers() ([]models.Customer, error)
	GetCustomerById(id uint) (*models.Customer, error)
	UpdateCustomer(customer *models.Customer) (*models.Customer, error)
	DeleteCustomer(id uint) error
	GetCustomerSummary(id uint) (*CustomerSummaryDTO, error)
}

type CustomerRepository struct {
	db *gorm.DB
}

var (
	repoInstance *CustomerRepository
	repoOnce     sync.Once
)

func NewCustomerRepository(db *gorm.DB) CustomerRepositoryInterface {
	log.Println(util.Gray + "CustomerRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &CustomerRepository{db: db}
	})
	return repoInstance
	// return &CustomerRepository{db: db}
}

func (r *CustomerRepository) CreateCustomer(customer *models.Customer) (*models.Customer, error) {

	err := r.db.Create(&customer).Error
	return customer, err
}

func (r *CustomerRepository) GetAllCustomers() ([]models.Customer, error) {
	customers := []models.Customer{}
	r.db.Model(&models.Customer{}).Order("ID asc").Find(&customers)
	if len(customers) == 0 {
		return nil, errors.New("NO records found")
	}
	return customers, nil
}

func (r *CustomerRepository) GetCustomerById(id uint) (*models.Customer, error) {
	var customer models.Customer
	result := r.db.First(&customer, "id = ?", id)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &customer, nil
}

func (r *CustomerRepository) UpdateCustomer(input *models.Customer) (*models.Customer, error) {

	var existingCustomer *models.Customer
	err := r.db.Where("id = ?", input.ID).First(&existingCustomer).Error
	if err != nil {
		// Handle error if customer not found or other issue
		return nil, err
	}

	log.Println("input: ", input)
	if input.Address == "" || input.Name == "" || input.Phone == "" {
		return nil, err
	}
	// Update relevant fields from input data
	existingCustomer.Name = input.Name // Update other fields as needed
	existingCustomer.Address = input.Address
	existingCustomer.Phone = input.Phone

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

func (r *CustomerRepository) DeleteCustomer(id uint) error {
	// return r.db.Delete(&User{}, id).Error
	return r.db.Delete(&models.Customer{}, id).Error
}

func (r *CustomerRepository) GetCustomerSummary(id uint) (*CustomerSummaryDTO, error) {
	var customer models.Customer

	// Fetch customer and the 5 most recent sales
	err := r.db.Preload("Sales", func(db *gorm.DB) *gorm.DB {
		return db.Order("sale_date DESC").Limit(5)
	}).First(&customer, id).Error

	if err != nil {
		return nil, err
	}

	return &CustomerSummaryDTO{
		ID:           customer.ID,
		Name:         customer.Name,
		Phone:        customer.Phone,
		OrderCount:   customer.OrderCount,
		TotalSpent:   customer.TotalSpent,
		LastPurchase: customer.LastPurchase,
		RecentSales:  customer.Sales,
	}, nil
}
