package service

import (
	"log"

	"github.com/sankangkin/di-rest-api/internal/models"
	"github.com/sankangkin/di-rest-api/internal/repository"
)

type CategoryServiceInterface interface {
	CreateCategory(category *models.Category) (*models.Category, error)
	GetAllCategories() ([]models.Category, error)
	GetCategoryById(id uint) (*models.Category, error)
	UpdateCategory(category *models.Category) (*models.Category, error)
	DeleteCategory(id uint) error
}

type categoryService struct {
	repo repository.CategoryRepositoryInterface
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) *categoryService{
	return &categoryService{repo: repo}
}

func(s *categoryService)CreateCategory(category *models.Category) (*models.Category, error) {
	return s.repo.CreateCategory(category)
}

func(s *categoryService) GetAllCategories() ([]models.Category, error) {
	log.Println("Invoking categoryService layer.....")
	return s.repo.GetAllCategories()
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