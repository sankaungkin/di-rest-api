package category_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	c "github.com/sankangkin/di-rest-api/internal/domain/category"
	"github.com/sankangkin/di-rest-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock service for CategoryService
type MockCategoryService struct {
	mock.Mock
	c.CategoryServiceInterface
}

func (m *MockCategoryService) GetAllCategories() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryService) GetCategoryById(id uint) (*models.Category, error) {
	args := m.Called(id)
	if res := args.Get(0); res != nil {
		return res.(*models.Category), nil
	}
	return &models.Category{}, args.Error(1)
}

// Test GetAllCategories Handler
func TestGetAllCategories(t *testing.T) {
	// Set up the Fiber app
	app := fiber.New()

	// Create the mock service
	mockService := new(MockCategoryService)

	// Create the handler and inject the mock service
	handler := &c.CategoryHandler{Svc: mockService}

	// Register the handler to the app
	app.Get("/category", handler.GetAllCategorie)

	t.Run("Success", func(t *testing.T) {
		// Define expected response
		mockCategories := []models.Category{
			{ID: 1, CategoryName: "Category 1"},
			{ID: 2, CategoryName: "Category 2"},
		}
		mockService.On("GetAllCategories").Return(mockCategories, nil)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/category", nil)
		resp, _ := app.Test(req, -1) // -1 disables request timeout

		// Assert status code
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse the response body
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		// Assert the response
		assert.Equal(t, "SUCCESS", response["status"])
		assert.Equal(t, strconv.Itoa(len(mockCategories))+" records found", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("No Categories Found", func(t *testing.T) {
		// Define nil response for categories
		mockService.On("GetAllCategories").Return([]models.Category{}, nil)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/category", nil)
		resp, _ := app.Test(req, -1)

		// Assert status code
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		// Parse the response body
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		// Assert the response
		assert.Equal(t, "No categories found", response["error"])

		mockService.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		// Define expected error response
		mockService.On("GetAllCategories").Return(nil, errors.New("database error"))

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/categories", nil)
		resp, _ := app.Test(req, -1)

		// Assert status code
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		// Parse the response body
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		// Assert the error message
		assert.Equal(t, "database error", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestGetAllCategories_Error(t *testing.T) {
	// Create mock service
	mockSvc := &MockCategoryService{}

	// Mock service to return error
	mockSvc.On("GetAllCategories").Return(nil, errors.New("internal error"))

	// Create handler with mock service
	handler := c.CategoryHandler{Svc: mockSvc}

	// Create fiber app for testing
	app := fiber.New()
	app.Get("/categories", handler.GetAllCategorie)

	// Simulate GET request
	req, err := http.NewRequest(http.MethodGet, "/categories", nil)
	assert.NoError(t, err)

	// Record the response
	recorder := httptest.NewRecorder()
	app.Test(req, 1)

	// Assert response status code
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	// Assert error message in response body
	body := bytes.TrimSpace(recorder.Body.Bytes())
	assert.Equal(t, `{"error":"internal error"}`, string(body))
}

func TestGetCategoryById_Success(t *testing.T) {

	app := fiber.New()

	// Create the mock service
	mockService := new(MockCategoryService)

	// Create the handler and inject the mock service
	handler := &c.CategoryHandler{Svc: mockService}

	// Register the handler to the app
	app.Get("/category/:id", handler.GetCategoryById)

	// Define expected category
	expectedCategory := &models.Category{ID: 1, CategoryName: "Category 2"}

	// Mock service behavior
	mockService.On("GetCategoryById", uint(1)).Return(expectedCategory, nil)

	t.Run("SUCCESS", func(t *testing.T) {
		// Define expected response
		mockCategory := &models.Category{ID: 1, CategoryName: "Category 1"}
		mockService.On("GetCategoryById", uint(1)).Return(mockCategory, nil)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/category/1", nil)
		resp, _ := app.Test(req, -1) // -1 disables request timeout

		// Assert status code
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse the response body
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		// Assert the response
		assert.Equal(t, "SUCCESS", response["status"])
		assert.Equal(t, "Record found", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})
}
