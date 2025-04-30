package test_test

import (
	"testing"

	"github.com/sankangkin/di-rest-api/internal/domain/category/mocks"
	"github.com/sankangkin/di-rest-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAll(t *testing.T) {
	mockTestCategoryRepo := new(mocks.CategoryRepositoryInterface)

	categories := []models.Category{
		{ID: 1, CategoryName: "Category 1"},
		{ID: 2, CategoryName: "Category 2"},
	}

	mockTestCategoryRepo.On("GetAll", mock.Anything).Return(categories, nil).Once()
	result, err := mockTestCategoryRepo.GetAll()

	assert.NoError(t, err)
	assert.Equal(t, categories, result)

	mockTestCategoryRepo.AssertExpectations(t)
}

func TestGetById(t *testing.T) {
	mockTestCategoryRepo := new(mocks.CategoryRepositoryInterface)

	category := &models.Category{ID: 1, CategoryName: "Category 1"}

	mockTestCategoryRepo.On("GetById", uint(1)).Return(category, nil)
	result, err := mockTestCategoryRepo.GetById(1)

	assert.NoError(t, err)
	assert.Equal(t, category, result)

	mockTestCategoryRepo.AssertExpectations(t)
}
