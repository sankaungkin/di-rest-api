package product

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type ProductServiceInterface interface {
	// CreateSerive(product *models.Product) (*models.Product, error)
	CreateProductWithDetails(dto *Create_Product_UnitConversion_Stock_Price_DTO) (*Create_Product_UnitConversion_Stock_Price_DTO, error)
	GetAllSerive() ([]ResponseProductDTO, error)
	GetAllWithoutStock() ([]ResponseProductDTO, error)
	GetProductsWithoutPrices() ([]ResponseProductDTO, error)
	GetByIdSerive(id string) (*models.Product, error)
	GetAllProductStocks() ([]ResponseProductStockDTO, error)
	GetProductStocksById(productId string) (*ResponseProductStockDTO, error)
	GetAllProductPrices() ([]ResponseProductUnitPriceDTO, error)
	GetProductUnitPricesByIdSerive(productId string) ([]ResponseProductUnitPriceDTO, error)
	GetUnitConversionsById(id string) (models.UnitConversion, error)
	GetAllUnitConversions() ([]models.UnitConversion, error)
	Update(input UpdateProductRequstDTO) (*models.Product, error)
	DeleteSerive(id string) error
	GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error)
	GetUniofMeasurementById(id string) (models.UnitOfMeasure, error)
	UpdateUnit(input *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	UpdateUnitConversion(input *models.UnitConversion) (*models.UnitConversion, error)
	GetProductPriceHistoryByProductId(productId string) ([]ResponseProductHistoryDTO, error)
	GetAllProductPriceHistory() ([]ResponseProductHistoryDTO, error)

	GetAllProductUnits() (*[]models.ProductUnit, error)
	GetProductUnitByProductId(id string) ([]models.ProductUnit, error)
	GetAllProducts() ([]models.Product, error)
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

//	func (s *ProductService) CreateSerive(product *models.Product) (*models.Product, error) {
//		return s.repo.Create(product)
//	}
func (s *ProductService) CreateProductWithDetails(dto *Create_Product_UnitConversion_Stock_Price_DTO) (*Create_Product_UnitConversion_Stock_Price_DTO, error) {
	return s.repo.CreateProductWithDetails(dto)
}

func (s *ProductService) GetAllSerive() ([]ResponseProductDTO, error) {
	return s.repo.GetAll()
}
func (s *ProductService) GetAllWithoutStock() ([]ResponseProductDTO, error) {
	return s.repo.GetAllWithoutStock()
}

func (s *ProductService) GetProductsWithoutPrices() ([]ResponseProductDTO, error) {
	return s.repo.GetProductsWithoutPrices()
}

func (s *ProductService) GetByIdSerive(id string) (*models.Product, error) {
	return s.repo.GetById(id)
}

func (s *ProductService) GetProductUnitPricesByIdSerive(productId string) ([]ResponseProductUnitPriceDTO, error) {
	return s.repo.GetProductUnitPricesById(productId)
}

func (s *ProductService) Update(product UpdateProductRequstDTO) (*models.Product, error) {
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

func (s *ProductService) UpdateUnitConversion(input *models.UnitConversion) (*models.UnitConversion, error) {
	return s.repo.UpdateUnitConversion(input)
}

func (s *ProductService) GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error) {
	return s.repo.GetAllUnitOfMeasurement()
}

func (s *ProductService) GetUniofMeasurementById(id string) (models.UnitOfMeasure, error) {
	return s.repo.GetUniofMeasurementById(id)
}

func (s *ProductService) UpdateUnit(input *models.UnitOfMeasure) (*models.UnitOfMeasure, error) {
	return s.repo.UpdateUnit(input)
}

func (s *ProductService) GetProductPriceHistoryByProductId(productId string) ([]ResponseProductHistoryDTO, error) {
	return s.repo.GetProductPriceHistoryByProductId(productId)
}

func (s *ProductService) GetAllProductPriceHistory() ([]ResponseProductHistoryDTO, error) {
	return s.repo.GetAllProductPriceHistory()
}

func (s *ProductService) GetAllProductUnits() (*[]models.ProductUnit, error) {
	return s.repo.GetAllProductUnits()
}

func (s *ProductService) GetProductUnitByProductId(id string) ([]models.ProductUnit, error) {
	return s.repo.GetProductUnitByProductId(id)
}

func (s *ProductService) GetAllProducts() ([]models.Product, error) {
	return s.repo.GetAllProducts()
}
