package product

import "github.com/sankangkin/di-rest-api/internal/models"

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

func NewProductService(repo ProductRepositoryInterface) ProductServiceInterface{
	return &ProductService{repo: repo}
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

