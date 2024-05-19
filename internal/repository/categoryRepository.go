package repository

import (
	"github.com/sankangkin/di-rest-api/internal/dto"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	CreateCategory(category *models.Category) (*models.Category, error)
	GetCategories() ([]*models.Category, error)
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

func (r *categoryRepository)GetCategories() ([]*models.Category, error){
	var categories []*models.Category
	err := r.db.Find(&categories)
	if err != nil {
		return nil, err.Error
	}
	return categories, nil

}

func (r *categoryRepository) GetCategoryById(id uint) (*models.Category, error){
	var category *models.Category
	err := r.db.Find(&category, "id = ?", category.ID)
	if err != nil {
		return nil, err.Error
	}
	return category, nil
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