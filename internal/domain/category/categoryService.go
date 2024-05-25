package category

import (
	"log"
	"sync"

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
//! singleton pattern
var (
	svcInstance *CategoryService
	svcOnce sync.Once
)

//! constructor must be return the Interface, NOT struct, if not, google wire generate fail
func NewCategoryService(repo CategoryRepositoryInterface) CategoryServiceInterface{

	log.Println(Red + "CategoryService constructor is called" + Reset)

	svcOnce.Do(func() {
		svcInstance = &CategoryService{repo: repo}
	})
	return svcInstance
}

func(s *CategoryService)CreateCategory(category *models.Category) (*models.Category, error) {
	return s.repo.Create(category)
}

func(s *CategoryService) GetAllCategories() ([]models.Category, error) {
	return s.repo.GetAll()
}

func(s *CategoryService) GetCategoryById(id uint) (*models.Category, error) {
	return s.repo.GetById(id)
}

func(s *CategoryService) UpdateCategory(category *models.Category) (*models.Category, error) {
	return s.repo.Update(category)
}

func(s *CategoryService) DeleteCategory(id uint) error {
	return s.repo.Delete(id)
}