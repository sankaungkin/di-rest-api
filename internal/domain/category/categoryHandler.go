package category

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	Svc CategoryServiceInterface
}

// ! singleton pattern
var (
	hdlInstance *CategoryHandler
	hdlOnce     sync.Once
)

// constructor
func NewCategoryHandler(svc CategoryServiceInterface) *CategoryHandler {

	log.Println(util.Blue + "CategoryHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &CategoryHandler{Svc: svc}
	})
	return hdlInstance
	// return &CategoryHandler{svc: svc}
}

// CreateCategory 	godoc
//
//	@Summary		Create new category based on parameters
//	@Description	Create new category based on parameters
//	@Tags			Categories
//	@Accept			json
//	@Param			category	body		CreateCategoryRequestDTO	true	"Category Data"
//	@Success		200			{object}	models.Category
//	@Failure		400			{object}	httputil.HttpError400
//	@Failure		401			{object}	httputil.HttpError401
//	@Failure		500			{object}	httputil.HttpError500
//	@Failure		401			{object}	httputil.HttpError401
//	@Router			/api/categories [post]
//
//	@Security		ApiKeyAuth
//
//	@param			Authorization	header	string	true	"Authorization"
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	newCategory := new(models.Category)
	if err := c.BodyParser(newCategory); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	errors := models.ValidateStruct(newCategory)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	if _, err := h.Svc.CreateCategory(newCategory); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data":    newCategory,
		})
}

// GetCategories godoc
//
//	@Summary		Fetch all Categories
//	@Description	Fetch all Categories
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Category
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/categories	[get]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *CategoryHandler) GetAllCategorie(c *fiber.Ctx) error {
	categories, err := h.Svc.GetAllCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if categories == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No categories found"})
	}

	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(categories)) + " records found",
			"data":    categories,
		})
}

// GetCategoryById godoc
//
//	@Summary		Fetch individual category by Id
//	@Description	Fetch individual category by Id
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"category Id"
//	@Success		200					{object}	models.Category
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/categories/{id}	[get]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *CategoryHandler) GetCategoryById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}

	category, err := h.Svc.GetCategoryById(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "Record not found",
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status": "FAIL", "message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Record found",
		"data":    category,
	})
}

// UpdateCategory godoc
//
//	@Summary		Update individual category
//	@Description	Update individual category
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string							true	"category Id"
//	@Param			category			body		dto.UpdateCategoryRequestDTO	true	"Category Data"
//	@Success		200					{object}	models.Category
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/categories/{id}	[put]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *CategoryHandler) UpdateCatagory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}

	foundCategory, err := h.Svc.GetCategoryById(uint(id))
	log.Println("foundCategory: ", foundCategory)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "ရှာမတွေ့ပါ။",
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status": "FAIL", "message": err.Error(),
		})
	}

	input := new(UpdateCategoryRequestDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	updateCategory := models.Category{
		ID:           foundCategory.ID,
		CategoryName: foundCategory.CategoryName,
	}

	if err := c.BodyParser(&updateCategory); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	result, err := h.Svc.UpdateCategory(&updateCategory)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})

}

// DeleteCategory godoc
//
//	@Summary		Delete individual category
//	@Description	Delete individual category
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"category Id"
//	@Success		200					{object}	models.Category
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/categories/{id}	[delete]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}
	category, err := h.Svc.GetCategoryById(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "Record not found",
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status": "FAIL", "message": err.Error(),
		})
	}
	err = h.Svc.DeleteCategory(uint(category.ID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Internal server error",
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}
