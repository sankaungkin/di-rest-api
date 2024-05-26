package supplier

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

type SupplierHandler struct {
	svc SupplierServiceInterface
}

var(
	hdlInstance *SupplierHandler
	hdlOnce sync.Once
)

func NewSupplierHandler(svc SupplierServiceInterface) *SupplierHandler{
	log.Println(util.Green + "SupplierHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &SupplierHandler{svc: svc}
	})
	return hdlInstance
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
	if _, err := h.svc.CreateSupplier(newSupplier); err != nil {
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
	Suppliers, err := h.svc.GetAllSuppliers()
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

	Supplier, err := h.svc.GetSupplierById(uint(id))
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
	foundSupplier, err := h.svc.GetSupplierById(uint(id))
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

	input := new(UpdateSupplierRequstDTO)
	log.Println("input: ", input)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	updateSupplier := models.Supplier{
		ID: foundSupplier.ID,
		Name: input.Name,
		Address: input.Address,
		Phone: input.Phone,
	}
	log.Println("updateCustomer: ", &updateSupplier)
	if err := c.BodyParser(&updateSupplier); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}	

	result, err :=	h.svc.UpdateSupplier(&updateSupplier)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})
}

func(h *SupplierHandler)DeleteSupplier(c *fiber.Ctx) error{
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	Supplier, err := h.svc.GetSupplierById(uint(id))
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
	h.svc.DeleteSupplier(uint(Supplier.ID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status" : "FAIL",
			"message" : "Internal server error",
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}