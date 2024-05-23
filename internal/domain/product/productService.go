package product

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/models"
)

type ProductServiceInterface interface {
	CreateProduct(product *models.Product) (*models.Product, error)
	GetAllProducts() ([]models.Product, error)
	GetProductById(id string) (*models.Product, error)
	UpdateProduct(product *models.Product) (*models.Product, error)
	DeleteProduct(id string)  error
}

type ProductService struct {
	repo ProductRepositoryInterface
}
//! singleton pattern
var (
	svcInstance *ProductService
	svcOnce sync.Once
)

// func NewProductService(repo ProductRepositoryInterface) ProductServiceInterface{
// 	return &ProductService{repo: repo}
// }
//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

func NewProductService(repo ProductRepositoryInterface) ProductServiceInterface{
	
	log.Println(Yellow + "ProductService constructor is called"+ Reset) 
	
	svcOnce.Do(func() {
		svcInstance = &ProductService{repo: repo}
	})
	return svcInstance
}

func (s *ProductService)CreateProduct(product *models.Product) (*models.Product, error){

	return s.repo.CreateProduct(product)
}
func (s *ProductService)GetAllProducts() ([]models.Product, error){
	return s.repo.GetAllProducts()
}
func (s *ProductService) GetProductById(id string) (*models.Product, error){
	return s.repo.GetProductById(id)
}

func (s *ProductService)UpdateProduct(product *models.Product) (*models.Product, error){
	return s.repo.UpdateProduct(product)
}

func (s *ProductService)DeleteProduct(id string)  error {
	return s.repo.DeleteProduct(id)
}

