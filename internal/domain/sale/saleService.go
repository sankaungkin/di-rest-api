package sale

import (
	"log"
	"sync"
	"time"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type SaleServiceInterface interface {
	CreateService(sale *models.Sale) (*models.Sale, error)
	GetAllService() ([]models.Sale, error)
	GetTodaySales() ([]models.Sale, error)
	GetTopTenSoleProducts() ([]ResponseTopTenSoleProductsDTO, error)
	GetDailySales() ([]ResponseDailySalesDTO, error)
	GetSalesByDate(date time.Time) ([]models.Sale, error)
	GetTodaySaleList() ([]models.Sale, error)
	GetById(id string) (*models.Sale, error)
	GetTodayGrandTotal() (int64, error)
	GetMonthlySales() ([]models.Sale, error)
	GetMonthlyGrandTotal() (int64, error)
	GetTopCustomers() (*ResponseTopCustomerDTO, error)
	UpdateSale(sale UpdateSaleRemarkDTO) (*models.Sale, error)

	GetSaleStockItemWithPrice() ([]ResponseSaleStockItemWithPrice, error)
	ReturnSaleItems(returnItem SaleReturnDTO) (*models.SaleReturn, error)
}

type SaleService struct {
	repo SaleRepositoryInterface
}

var (
	svcInstance *SaleService
	svcOnce     sync.Once
)

func NewSaleService(repo SaleRepositoryInterface) SaleServiceInterface {
	log.Println(util.Blue + "SaleService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &SaleService{repo: repo}
	})
	return svcInstance
}

func (s *SaleService) CreateService(sale *models.Sale) (*models.Sale, error) {
	return s.repo.Create(sale)
}

func (s *SaleService) GetAllService() ([]models.Sale, error) {
	return s.repo.GetAll()
}

func (s *SaleService) GetTodaySales() ([]models.Sale, error) {
	return s.repo.GetTodaySales()
}

func (s *SaleService) GetById(id string) (*models.Sale, error) {
	return s.repo.GetById(id)
}

func (s *SaleService) GetTodayGrandTotal() (int64, error) {
	return s.repo.GetTodayGrandTotal()
}

func (s *SaleService) GetMonthlySales() ([]models.Sale, error) {
	return s.repo.GetMonthlySales()
}

func (s *SaleService) GetMonthlyGrandTotal() (int64, error) {
	return s.repo.GetMonthlyGrandTotal()
}

func (s *SaleService) GetTopCustomers() (*ResponseTopCustomerDTO, error) {
	return s.repo.TopCustomers()
}

func (s *SaleService) GetTopTenSoleProducts() ([]ResponseTopTenSoleProductsDTO, error) {
	return s.repo.GetTopTenSoleProducts()
}

func (s *SaleService) GetDailySales() ([]ResponseDailySalesDTO, error) {
	return s.repo.GetDailySales()
}

func (s *SaleService) GetSaleStockItemWithPrice() ([]ResponseSaleStockItemWithPrice, error) {
	return s.repo.GetSaleStockItemWithPrice()
}

func (s *SaleService) UpdateSale(sale UpdateSaleRemarkDTO) (*models.Sale, error) {
	return s.repo.UpdateSale(sale)
}

func (s *SaleService) GetSalesByDate(date time.Time) ([]models.Sale, error) {
	return s.repo.GetSalesByDate(date)
}

func (s *SaleService) GetTodaySaleList() ([]models.Sale, error) {
	return s.repo.GetTodaySaleList()
}

func (s *SaleService) ReturnSaleItems(returnItem SaleReturnDTO) (*models.SaleReturn, error) {
	return s.repo.ReturnSaleItems(returnItem)
}
