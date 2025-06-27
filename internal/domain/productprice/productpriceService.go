package productprice

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type ProductPriceServiceInterface interface {
	Create(productPrice *models.ProductPrice) (*models.ProductPrice, error)
	GetAll() ([]ProductPriceResponseDTO, error)
	GetById(id int) (*ProductPriceResponseDTO, error)
	UpdateProductPrice(input *models.ProductPrice) (*models.ProductPrice, error)
	DeleteProductPrice(id int) error
}

type ProductPriceService struct {
	repo ProductPriceRepositoryInterface
}

// ! singleton pattern
var (
	svcInstance *ProductPriceService
	svcOnce     sync.Once
)

// func NewProductPriceService(repo ProductPriceRepositoryInterface) ProductPriceServiceInterface{
// 	return &ProductPriceService{repo: repo}
// }
//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

func NewProductPriceService(repo ProductPriceRepositoryInterface) ProductPriceServiceInterface {

	log.Println(util.Yellow + "ProductPriceService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &ProductPriceService{repo: repo}
	})
	return svcInstance
}

func (s *ProductPriceService) Create(productPrice *models.ProductPrice) (*models.ProductPrice, error) {
	return s.repo.Create(productPrice)
}
func (s *ProductPriceService) GetAll() ([]ProductPriceResponseDTO, error) {
	return s.repo.GetAll()
}
func (s *ProductPriceService) GetById(id int) (*ProductPriceResponseDTO, error) {
	return s.repo.GetById(id)
}

func (s *ProductPriceService) UpdateProductPrice(input *models.ProductPrice) (*models.ProductPrice, error) {
	return s.repo.UpdateProductPrice(input)
}

func (s *ProductPriceService) DeleteProductPrice(id int) error {
	return s.repo.DeleteProductPrice(id)
}
