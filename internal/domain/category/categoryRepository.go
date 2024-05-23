package category

import (
	"errors"
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/dto"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CategoryRepositoryInterface interface {
	CreateCategory(category *models.Category) (*models.Category, error)
	GetAllCategories() ([]models.Category, error)
	GetCategoryById(id uint) (*models.Category, error)
	UpdateCategory(category *models.Category) (*models.Category, error)
	DeleteCategory(id uint) error
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
	log.Println(Red+"CategoryRepository constructor is called"+Reset)
	repoOnce.Do(func(){
		repoInstance = &CategoryRepository{db: db}
	})
	return repoInstance
}

func (r *CategoryRepository)CreateCategory(category *models.Category) (*models.Category, error) {

	input := new(dto.CreateCategoryRequestDTO)
	log.Println("input:", input)
	newCategory := &models.Category{
		CategoryName: input.CategoryName,
	}
	err := r.db.Create(newCategory)
	if err != nil {
		return nil, err.Error
	}
	return newCategory, nil
}

func (r *CategoryRepository)GetAllCategories() ([]models.Category, error){

	categories := []models.Category{}
	r.db.Model(&models.Category{}).Order("ID asc").Limit(100).Find(&categories)
	if len(categories) == 0 {
		return nil, errors.New("no record found")
	}
	return categories, nil

}

func (r *CategoryRepository) GetCategoryById(id uint) (*models.Category, error){
	var category models.Category
	result := r.db.First(&category, "id = ?", id)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &category, nil
}

func (r *CategoryRepository) UpdateCategory(category *models.Category) (*models.Category, error) {
	var updateCategory *models.Category
	err := r.db.Find(&updateCategory, "id = ?", category.ID)
	if err != nil {
		return nil, err.Error
	}
	r.db.Save(&updateCategory)
	return updateCategory, nil
}

func (r *CategoryRepository) DeleteCategory(id uint) error {
	var deletedCategory *models.Category
	err := r.db.Find(&deletedCategory, "id = ?", id)
	if err != nil {
		return  err.Error
	}
	r.db.Delete(&deletedCategory)
	return nil
}