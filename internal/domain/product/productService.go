package product

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type ProductServiceInterface interface {
	CreateSerive(product *models.Product) (*models.Product, error)
	GetAllSerive() ([]ResponseProductDTO, error)
	GetByIdSerive(id string) (*models.Product, error)
	GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	GetAllProductPrices() ([]ResponseProductUnitPriceDTO, error)
	GetProductUnitPricesByIdSerive(productId string) ([]ResponseProductUnitPriceDTO, error)
	GetUnitConversionsById(id string) (models.UnitConversion, error)
	GetAllUnitConversions() ([]models.UnitConversion, error)
	Update(product *models.Product) (*models.Product, error)
	DeleteSerive(id string) error
	GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error)
}

type ProductService struct {
	repo ProductRepositoryInterface
}

// ! singleton pattern
var (
	svcInstance *ProductService
	svcOnce     sync.Once
)

// func NewProductService(repo ProductRepositoryInterface) ProductServiceInterface{
// 	return &ProductService{repo: repo}
// }
//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

func NewProductService(repo ProductRepositoryInterface) ProductServiceInterface {

	log.Println(util.Yellow + "ProductService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &ProductService{repo: repo}
	})
	return svcInstance
}

func (s *ProductService) CreateSerive(product *models.Product) (*models.Product, error) {

	return s.repo.Create(product)
}
func (s *ProductService) GetAllSerive() ([]ResponseProductDTO, error) {
	return s.repo.GetAll()
}
func (s *ProductService) GetByIdSerive(id string) (*models.Product, error) {
	return s.repo.GetById(id)
}

func (s *ProductService) GetProductUnitPricesByIdSerive(productId string) ([]ResponseProductUnitPriceDTO, error) {
	return s.repo.GetProductUnitPricesById(productId)
}

func (s *ProductService) Update(product *models.Product) (*models.Product, error) {
	return s.repo.Update(product)
}

func (s *ProductService) DeleteSerive(id string) error {
	return s.repo.Delete(id)
}

func (s *ProductService) GetAllProductStocks() ([]ResponseProductStockDTO, error) {
	return s.repo.GetAllProductStocks()
}

func (s *ProductService) GetProductStocksById(productId string) (*ResponseProductStockDTO, error) {
	return s.repo.GetProductStocksById(productId)
}

func (s *ProductService) GetAllProductPrices() ([]ResponseProductUnitPriceDTO, error) {
	return s.repo.GetAllProductPrices()
}
func (s *ProductService) GetUnitConversionsById(id string) (models.UnitConversion, error) {
	return s.repo.GetUnitConversionsById(id)
}

func (s *ProductService) GetAllUnitConversions() ([]models.UnitConversion, error) {
	return s.repo.GetAllUnitConversions()
}

func (s *ProductService) GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error) {
	return s.repo.GetAllUnitOfMeasurement()
}
