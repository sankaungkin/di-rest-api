package handler

import (
	"net/http"

	"github.com/gofiber/fiber"
	"github.com/sankangkin/di-rest-api/internal/models"
	"github.com/sankangkin/di-rest-api/internal/service"
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