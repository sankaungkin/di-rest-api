package category

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CategoryHandler struct {
	svc CategoryServiceInterface
}
//! singleton pattern
var (
	hdlInstance *CategoryHandler
	hdlOnce sync.Once
	Red = "\033[31m" 
	Reset = "\033[0m" 
)
// constructor 
func NewCategoryHandler(svc CategoryServiceInterface) *CategoryHandler{
	
	log.Println(Red + "CategoryHandler constructor is called"+ Reset)
	hdlOnce.Do(func() {
		hdlInstance = &CategoryHandler{svc: svc}
	})
	return hdlInstance
	// return &CategoryHandler{svc: svc}
}

func(h *CategoryHandler) CreateCategory(c *fiber.Ctx) error{
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

	if _, err := h.svc.CreateCategory(newCategory); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data":    newCategory,
		})
}

func(h *CategoryHandler) GetAllCategorie(c *fiber.Ctx) error{
	categories, err := h.svc.GetAllCategories() 
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(categories)) + " records found",
			"data":    categories,
		})
}

func(h *CategoryHandler) GetCategoryById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}

	category, err := h.svc.GetCategoryById(uint(id))
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

func(h *CategoryHandler) UpdateCatagory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	category, err := h.svc.GetCategoryById(uint(id))
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
	h.svc.UpdateCategory(category)
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Update successfully",
	})
	
}

func(h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	category, err := h.svc.GetCategoryById(uint(id))
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
	h.svc.DeleteCategory(uint(category.ID))
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}