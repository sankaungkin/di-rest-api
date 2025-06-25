package unitconversion

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type UnitConversionRepositoryInterface interface {
	Create(unitConversion *models.UnitConversion) (*models.UnitConversion, error)
	GetAll() ([]models.UnitConversion, error)
	GetById(id int) (*models.UnitConversion, error)
	Update(unitConversion *models.UnitConversion) (*models.UnitConversion, error)
	Delete(id int) error
}

type UnitConversionRepository struct {
	db *gorm.DB
}

// ! singleton pattern
var (
	repoInstance *UnitConversionRepository
	repoOnce     sync.Once
)

// func NewUnitConversionRepository(db *gorm.DB) UnitConversionRepositoryInterface {
// 	return &UnitConversionRepository{db: db}
// }

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

// constructor
func NewUnitConversionRepository(db *gorm.DB) UnitConversionRepositoryInterface {
	log.Println(util.Yellow + "UnitConversionRepository constructor is called " + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &UnitConversionRepository{db: db}
	})
	return repoInstance
}

func (r *UnitConversionRepository) Create(unitConversion *models.UnitConversion) (*models.UnitConversion, error) {
	err := r.db.Create(&unitConversion).Error
	return unitConversion, err
}

func (r *UnitConversionRepository) GetAll() ([]models.UnitConversion, error) {
	var unitConversions []models.UnitConversion
	err := r.db.Model(&models.UnitConversion{}).Order("id DESC").Find(&unitConversions).Error
	if err != nil {
		return nil, err
	}
	if len(unitConversions) == 0 {
		return nil, errors.New("no records found")
	}

	return unitConversions, nil
}

func (r *UnitConversionRepository) GetById(id int) (*models.UnitConversion, error) {

	var unitConversion models.UnitConversion
	result := r.db.First(&unitConversion, "id = ?", id)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &unitConversion, nil
}

func (r *UnitConversionRepository) Update(unitConversion *models.UnitConversion) (*models.UnitConversion, error) {
	var existingUnitConversion models.UnitConversion
	err := r.db.Where("id = ?", unitConversion.ID).First(&existingUnitConversion).Error
	if err != nil {
		return nil, err
	}

	if unitConversion.BaseUnitId == 0 || unitConversion.DeriveUnitId == 0 || unitConversion.Factor == 0 {
		return nil, fmt.Errorf("missing required fields")
	}

	// existingUnitConversion.BaseUnit = unitConversion.BaseUnit
	// existingUnitConversion.DeriveUnit = unitConversion.DeriveUnit
	existingUnitConversion.BaseUnitId = unitConversion.BaseUnitId
	existingUnitConversion.DeriveUnitId = unitConversion.DeriveUnitId
	existingUnitConversion.Factor = unitConversion.Factor
	existingUnitConversion.Description = unitConversion.Description
	existingUnitConversion.ProductId = unitConversion.ProductId

	err = r.db.Save(&existingUnitConversion).Error
	if err != nil {
		return nil, err
	}

	return &existingUnitConversion, nil
}

func (r *UnitConversionRepository) Delete(id int) error {
	// return r.db.Delete(&User{}, id).Error

	var unitConversion models.UnitConversion
	result := r.db.First(&unitConversion, "id = ?", id)

	if err := result.Error; err != nil {
		return err
	}

	// return r.db.Delete(&unitConversion).Error
	return r.db.Unscoped().Delete(&unitConversion).Error

}
