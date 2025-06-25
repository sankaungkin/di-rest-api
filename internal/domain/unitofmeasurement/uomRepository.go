package unitofmeasurement

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type UnitOfMeasurementRepositoryInterface interface {
	Create(unitOfMeasurement *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	GetAll() ([]models.UnitOfMeasure, error)
	GetById(id int) (*models.UnitOfMeasure, error)
	Update(unitOfMeasurement *models.UnitOfMeasure) (*models.UnitOfMeasure, error)
	Delete(id int) error
}

type UnitOfMeasurementRepository struct {
	db *gorm.DB
}

// ! singleton pattern
var (
	repoInstance *UnitOfMeasurementRepository
	repoOnce     sync.Once
)

// func NewUnitOfMeasurementRepository(db *gorm.DB) UnitOfMeasurementRepositoryInterface {
// 	return &UnitOfMeasurementRepository{db: db}
// }

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

// constructor
func NewUnitOfMeasurementRepository(db *gorm.DB) UnitOfMeasurementRepositoryInterface {
	log.Println(util.Yellow + "UnitOfMeasurementRepository constructor is called " + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &UnitOfMeasurementRepository{db: db}
	})
	return repoInstance
}

func (r *UnitOfMeasurementRepository) Create(unitOfMeasurement *models.UnitOfMeasure) (*models.UnitOfMeasure, error) {
	err := r.db.Create(&unitOfMeasurement).Error
	return unitOfMeasurement, err
}

func (r *UnitOfMeasurementRepository) GetAll() ([]models.UnitOfMeasure, error) {
	var unitOfMeasurements []models.UnitOfMeasure
	err := r.db.Model(&models.UnitOfMeasure{}).Order("id DESC").Find(&unitOfMeasurements).Error
	if err != nil {
		return nil, err
	}
	if len(unitOfMeasurements) == 0 {
		return nil, errors.New("no records found")
	}

	return unitOfMeasurements, nil
}

func (r *UnitOfMeasurementRepository) GetById(id int) (*models.UnitOfMeasure, error) {

	var unitOfMeasurement models.UnitOfMeasure
	result := r.db.First(&unitOfMeasurement, "id = ?", int(id))
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &unitOfMeasurement, nil
}

func (r *UnitOfMeasurementRepository) Update(input *models.UnitOfMeasure) (*models.UnitOfMeasure, error) {
	var existingUnit models.UnitOfMeasure
	err := r.db.Where("id = ?", input.ID).First(&existingUnit).Error
	if err != nil {
		return nil, err
	}

	log.Println("input from Repository: ", input)
	if input.UnitName == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	existingUnit.UnitName = input.UnitName

	log.Println("existingUnit to update: ", existingUnit)
	err = r.db.Save(&existingUnit).Error
	if err != nil {
		return nil, err
	}

	return &existingUnit, nil
}

func (r *UnitOfMeasurementRepository) Delete(id int) error {
	// return r.db.Delete(&User{}, id).Error

	var unitOfMeasurement models.UnitOfMeasure
	result := r.db.First(&unitOfMeasurement, "id = ?", id)

	if err := result.Error; err != nil {
		return err
	}

	// return r.db.Delete(&unitOfMeasurement).Error
	return r.db.Unscoped().Delete(&unitOfMeasurement).Error

}
