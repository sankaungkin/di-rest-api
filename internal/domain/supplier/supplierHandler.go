package supplier

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type SupplierHandler struct {
	srv SupplierServiceInterface
}

func NewSupplierHandler(svc SupplierServiceInterface) *SupplierHandler{
	return &SupplierHandler{srv: svc}
}

func(h *SupplierHandler)CreateSupplier(c *fiber.Ctx) error {
	newSupplier := new(models.Supplier)
	if err := c.BodyParser(newSupplier); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	errors := models.ValidateStruct(newSupplier)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	if _, err := h.srv.CreateSupplier(newSupplier); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data":    newSupplier,
		})
}

func(h *SupplierHandler) GetAllSuppliers(c *fiber.Ctx) error {
	Suppliers, err := h.srv.GetAllSuppliers()
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(Suppliers)) + " records found",
			"data":    Suppliers,
		})
}

func(h *SupplierHandler) GetSupplierById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}

	Supplier, err := h.srv.GetSupplierById(uint(id))
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
		"data":    Supplier,
	})
}

func(h *SupplierHandler)UpdateSupplier(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	Supplier, err := h.srv.GetSupplierById(uint(id))
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
	h.srv.UpdateSupplier(Supplier)
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Update successfully",
	})
}

func(h *SupplierHandler)DeleteSupplier(c *fiber.Ctx) error{
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	Supplier, err := h.srv.GetSupplierById(uint(id))
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
	h.srv.DeleteSupplier(uint(Supplier.ID))
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}