package service

import (
	"github.com/sankangkin/di-rest-api/internal/models"
	"github.com/sankangkin/di-rest-api/internal/repository"
)

type CategoryService interface {
	CreateCategory(category *models.Category) error
	GetCategories() ([]*models.Category, error)
	GetCategoryById(id uint) (*models.Category, error)
	UpdateCategory(category *models.Category) (*models.Category, error)
	DeleteCategory(id uint) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) *categoryService{
	return &categoryService{repo: repo}
}

func(s *categoryService)CreateCategory(category *models.Category) (*models.Category, error) {
	return s.repo.CreateCategory(category)
}

func(s *categoryService) GetCategories() ([]*models.Category, error) {
	return s.repo.GetCategories()
}

func(s *categoryService) GetCategoryById(id uint) (*models.Category, error) {
	return s.repo.GetCategoryById(id)
}

func(s *categoryService) UpdateCategory(category *models.Category) (*models.Category, error) {
	return s.repo.UpdateCategory(category)
}

func(s *categoryService) DeleteCategory(id uint) error {
	return s.repo.DeleteCategory(id)
}