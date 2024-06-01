package category

import (
	"errors"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {
	Create(category *models.Category) (*models.Category, error)
	GetAll() ([]models.Category, error)
	GetById(id uint) (*models.Category, error)
	Update(category *models.Category) (*models.Category, error)
	Delete(id uint) error
}

type CategoryRepository struct{
	db *gorm.DB
}

var (
	repoInstance *CategoryRepository
	repoOnce sync.Once
)

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail
func NewCategoryRepository(db *gorm.DB) CategoryRepositoryInterface {
	log.Println(util.Blue+"CategoryRepository constructor is called"+util.Reset)
	repoOnce.Do(func(){
		repoInstance = &CategoryRepository{db: db}
	})
	return repoInstance
}

func (r *CategoryRepository)Create(category *models.Category) (*models.Category, error) {

	err := r.db.Create(&category).Error

	return category, err
}

func (r *CategoryRepository)GetAll() ([]models.Category, error){

	categories := []models.Category{}
	r.db.Model(&models.Category{}).Order("ID asc").Limit(100).Find(&categories)
	if len(categories) == 0 {
		return nil, errors.New("no record found")
	}
	return categories, nil

}

func (r *CategoryRepository) GetById(id uint) (*models.Category, error){
	var category models.Category
	result := r.db.First(&category, "id = ?", id)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &category, nil
}

func (r *CategoryRepository) Update(input *models.Category) (*models.Category, error) {

	log.Println("input from CategoryRepository: ", input)
	var existingCategory *models.Category
		err := r.db.Where("id = ?", input.ID).First(&existingCategory).Error
		if err != nil {
			// Handle error if customer not found or other issue
			return nil, err
		}

		log.Println("input: ", input)
		if input.CategoryName == ""  {
			return nil, err
		}
		// Update relevant fields from input data
		existingCategory.CategoryName = input.CategoryName  // Update other fields as needed

		// Save the updated customer data
		log.Println("existingCustomer: ", existingCategory)
		err = r.db.Updates(&existingCategory).Error
		if err != nil {
			// Handle error if update fails
			return nil, err
		}

		// Return the updated customer object
		return existingCategory, nil
}

func (r *CategoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.Category{}, id).Error
}