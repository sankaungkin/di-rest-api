package unitconversion

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type UnitConversionServiceInterface interface {
	CreateUnitConversion(unitConversion *models.UnitConversion) (*models.UnitConversion, error)
	GetAllUnitConversions() ([]models.UnitConversion, error)
	GetUnitConversionById(id int) (*models.UnitConversion, error)
	UpdateUnitConversion(unitConversion *models.UnitConversion) (*models.UnitConversion, error)
	DeleteUnitConversion(id int) error
}

type UnitConversionService struct {
	repo UnitConversionRepositoryInterface
}

// ! for singletone pattern
var (
	svcInstance *UnitConversionService
	svcOnce     sync.Once
)

// func NewUnitConversionService(repo UnitConversionRepositoryInterface) UnitConversionServiceInterface{
// 	return &UnitConversionService{repo:repo}
// }

// ! constructor must be return the Interface, NOT struct, if not, google wire generate fail
func NewUnitConversionService(repo UnitConversionRepositoryInterface) UnitConversionServiceInterface {
	log.Println(util.Gray + "UnitConversionService constructor is called " + util.Reset)
	svcOnce.Do(func() {
		svcInstance = &UnitConversionService{repo: repo}
	})
	return svcInstance
}

func (s *UnitConversionService) CreateUnitConversion(unitConversion *models.UnitConversion) (*models.UnitConversion, error) {
	return s.repo.Create(unitConversion)
}

func (s *UnitConversionService) GetAllUnitConversions() ([]models.UnitConversion, error) {
	return s.repo.GetAll()
}

func (s *UnitConversionService) GetUnitConversionById(id int) (*models.UnitConversion, error) {
	return s.repo.GetById(id)
}

func (s *UnitConversionService) UpdateUnitConversion(unitConversion *models.UnitConversion) (*models.UnitConversion, error) {
	return s.repo.Update(unitConversion)
}

func (s *UnitConversionService) DeleteUnitConversion(id int) error {
	return s.repo.Delete(id)
}
