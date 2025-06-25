package unitofmeasurement

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type UnitOfMeasurementServiceInterface interface {
	CreateUnitOfMeasurement(unitOfMeasurement *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error)
	GetUnitOfMeasurementById(id int) (*models.UnitOfMeasure, error)
	UpdateUnitOfMeasurement(unitOfMeasurement *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	DeleteUnitOfMeasurement(id int) error
}

type UnitOfMeasurementService struct {
	repo UnitOfMeasurementRepositoryInterface
}

// ! singleton pattern
var (
	svcInstance *UnitOfMeasurementService
	svcOnce     sync.Once
)

// func NewUnitOfMeasurementService(repo UnitOfMeasurementRepositoryInterface) UnitOfMeasurementServiceInterface{
// 	return &UnitOfMeasurementService{repo: repo}
// }
//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

func NewUnitOfMeasurementService(repo UnitOfMeasurementRepositoryInterface) UnitOfMeasurementServiceInterface {

	log.Println(util.Yellow + "UnitOfMeasurementService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &UnitOfMeasurementService{repo: repo}
	})
	return svcInstance
}

func (s *UnitOfMeasurementService) CreateUnitOfMeasurement(unitOfMeasurement *models.UnitOfMeasure) (*models.UnitOfMeasure, error) {

	return s.repo.Create(unitOfMeasurement)
}
func (s *UnitOfMeasurementService) GetAllUnitOfMeasurement() ([]models.UnitOfMeasure, error) {
	return s.repo.GetAll()
}
func (s *UnitOfMeasurementService) GetUnitOfMeasurementById(id int) (*models.UnitOfMeasure, error) {
	return s.repo.GetById(id)
}

func (s *UnitOfMeasurementService) UpdateUnitOfMeasurement(unitOfMeasurement *models.UnitOfMeasure) (*models.UnitOfMeasure, error) {
	return s.repo.Update(unitOfMeasurement)
}

func (s *UnitOfMeasurementService) DeleteUnitOfMeasurement(id int) error {
	return s.repo.Delete(id)
}
