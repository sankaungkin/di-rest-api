package repository

import (
	"errors"
	"log"

	"github.com/sankangkin/di-rest-api/internal/dto"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	CreateCategory(category *models.Category) (*models.Category, error)
	GetAllCategories() ([]models.Category, error)
	GetCategoryById(id uint) (*models.Category, error)
	UpdateCategory(category *models.Category) (*models.Category, error)
	DeleteCategory(id uint) error
}

type categoryRepository struct{
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *categoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository)CreateCategory(category *models.Category) (*models.Category, error) {

	input := new(dto.CreateCategoryRequestDTO)
	newCategory := &models.Category{
		CategoryName: input.CategoryName,
	}
	err := r.db.Create(newCategory)
	if err != nil {
		return nil, err.Error
	}
	return newCategory, nil
}

func (r *categoryRepository)GetAllCategories() ([]models.Category, error){

	log.Println("Invoking repository layer .....")
	categories := []models.Category{}
	r.db.Model(&models.Category{}).Order("ID asc").Limit(100).Find(&categories)
	if len(categories) == 0 {
		return nil, errors.New("no record found")
	}
	return categories, nil

}

func (r *categoryRepository) GetCategoryById(id uint) (*models.Category, error){
	var category models.Category
	result := r.db.First(&category, "id = ?", id)
	if err := result.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	return &category, nil
}

func (r *categoryRepository) UpdateCategory(category *models.Category) (*models.Category, error) {
	var updateCategory *models.Category
	err := r.db.Find(&updateCategory, "id = ?", category.ID)
	if err != nil {
		return nil, err.Error
	}
	r.db.Save(&updateCategory)
	return updateCategory, nil
}

func (r *categoryRepository) DeleteCategory(id uint) error {
	var deletedCategory *models.Category
	err := r.db.Find(&deletedCategory, "id = ?", id)
	if err != nil {
		return  err.Error
	}
	r.db.Delete(&deletedCategory)
	return nil
}