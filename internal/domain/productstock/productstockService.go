package productstock

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type ProductStockServiceInterface interface {
	GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	UpdateProductStocksById(productStock *models.ProductStock) (*models.ProductStock, error)
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

func (s *ProductStockService) GetAllProductStocks() ([]ResponseProductStockDTO, error) {
	return s.repo.GetAllProductStocks()
}

func (s *ProductStockService) GetProductStocksById(productId string) (*ResponseProductStockDTO, error) {
	return s.repo.GetProductStocksById(productId)
}

func (s *ProductStockService) UpdateProductStocksById(productStock *models.ProductStock) (*models.ProductStock, error) {
	return s.repo.UpdateProductStocksById(productStock)
}
