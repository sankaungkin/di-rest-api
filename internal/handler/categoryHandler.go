package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber"
	"github.com/sankangkin/di-rest-api/internal/models"
	"github.com/sankangkin/di-rest-api/internal/service"
	"gorm.io/gorm"
)

type categoryHandler struct {
	srv service.CategoryService
}

func NewCategoryHandler(srv service.CategoryService) *categoryHandler{
	return &categoryHandler{srv: srv}
}

func(h *categoryHandler) CreateCategoryHandler(c *fiber.Ctx) error{
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

	if err := h.srv.CreateCategory(newCategory); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data":    newCategory,
		})
}

func(h *categoryHandler) GetAllCategoriesHandler(c *fiber.Ctx) error{
	categories, err := h.srv.GetCategories() 
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": len(categories),
			"data":    categories,
		})
}

func(h *categoryHandler) GetCategoryById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}

	category, err := h.srv.GetCategoryById(uint(id))
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

func(h *categoryHandler) UpdateCatagory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	category, err := h.srv.GetCategoryById(uint(id))
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
	h.srv.UpdateCategory(category)
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Update successfully",
	})
	
}

func(h *categoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	category, err := h.srv.GetCategoryById(uint(id))
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
	h.srv.DeleteCategory(uint(category.ID))
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}