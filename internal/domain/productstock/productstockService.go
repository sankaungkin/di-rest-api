package productstock

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type ProductStockServiceInterface interface {
	CreateProductStocks(productStock *models.ProductStock) (*models.ProductStock, error)
	// GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetAllProductStocks() ([]models.ProductStock, error)
	GetLowStockProducts() ([]ResponseProductStockDTO, error)
	GetOutOfStockProducts() ([]ResponseProductStockDTO, error)
	// GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	GetProductStocksById(productId string) (models.ProductStock, error)
	UpdateProductStocksById(productStock UpdateProductStockDTO) (*models.ProductStock, error)
	GetDetailsProductStockById(productId string) (*StockResponse, error)
	GetConcreteBlockHeads() ([]ConcreteBlockHead, error)
	GetAllProductStocksWithCategory() ([]ProductStockListInfoWithCategory, error)
}

type ProductStockService struct {
	repo ProductStockRepositoryInterface
}

// ! singleton pattern
var (
	svcInstance *ProductStockService
	svcOnce     sync.Once
)

// func NewProductStockService(repo ProductStockRepositoryInterface) ProductStockServiceInterface{
// 	return &ProductStockService{repo: repo}
// }
//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

func NewProductStockService(repo ProductStockRepositoryInterface) ProductStockServiceInterface {

	log.Println(util.Yellow + "ProductStockService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &ProductStockService{repo: repo}
	})
	return svcInstance
}

func (s *ProductStockService) CreateProductStocks(productStock *models.ProductStock) (*models.ProductStock, error) {
	return s.repo.CreateProductStocks(productStock)
}

// func (s *ProductStockService) GetAllProductStocks() ([]ResponseProductStockDTO, error) {
// 	return s.repo.GetAllProductStocks()
// }

func (s *ProductStockService) GetAllProductStocks() ([]models.ProductStock, error) {
	return s.repo.GetAllProductStocks()
}

func (s *ProductStockService) GetLowStockProducts() ([]ResponseProductStockDTO, error) {
	return s.repo.GetLowStockProducts()
}

func (s *ProductStockService) GetOutOfStockProducts() ([]ResponseProductStockDTO, error) {
	return s.repo.GetOutOfStockProducts()
}

// func (s *ProductStockService) GetProductStocksById(productId string) (*ResponseProductStockDTO, error) {
// 	return s.repo.GetProductStocksById(productId)
// }

func (s *ProductStockService) UpdateProductStocksById(productStock UpdateProductStockDTO) (*models.ProductStock, error) {
	return s.repo.UpdateProductStocksById(productStock)
}

func (s *ProductStockService) GetProductStocksById(productId string) (models.ProductStock, error) {
	return s.repo.GetProductStocksById(productId)
}

func (s *ProductStockService) GetDetailsProductStockById(productId string) (*StockResponse, error) {
	return s.repo.GetDetailsProductStockById(productId)
}

func (s *ProductStockService) GetAllProductStocksWithCategory() ([]ProductStockListInfoWithCategory, error) {
	return s.repo.GetAllProductStocksWithCategory()
}

func (s *ProductStockService) GetConcreteBlockHeads() ([]ConcreteBlockHead, error) {
	return s.repo.GetConcreteBlockHeads()
}
