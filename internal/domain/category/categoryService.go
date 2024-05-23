package category

import (
	"github.com/sankangkin/di-rest-api/internal/models"
)

type CategoryServiceInterface interface {
	CreateCategory(category *models.Category) (*models.Category, error)
	GetAllCategories() ([]models.Category, error)
	GetCategoryById(id uint) (*models.Category, error)
	UpdateCategory(category *models.Category) (*models.Category, error)
	DeleteCategory(id uint) error
}

type CategoryService struct {
	repo CategoryRepositoryInterface
}

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail
func NewCategoryService(repo CategoryRepositoryInterface) CategoryServiceInterface{
	return &CategoryService{repo: repo}
}

func(s *CategoryService)CreateCategory(category *models.Category) (*models.Category, error) {
	return s.repo.CreateCategory(category)
}

func(s *CategoryService) GetAllCategories() ([]models.Category, error) {
	return s.repo.GetAllCategories()
}

func(s *CategoryService) GetCategoryById(id uint) (*models.Category, error) {
	return s.repo.GetCategoryById(id)
}

func(s *CategoryService) UpdateCategory(category *models.Category) (*models.Category, error) {
	return s.repo.UpdateCategory(category)
}

func(s *CategoryService) DeleteCategory(id uint) error {
	return s.repo.DeleteCategory(id)
}