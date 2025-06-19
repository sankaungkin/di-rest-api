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

var (
	hdlInstance *SupplierHandler
	hdlOnce     sync.Once
)

func NewSupplierHandler(svc SupplierServiceInterface) *SupplierHandler {
	log.Println(util.Green + "SupplierHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &SupplierHandler{svc: svc}
	})
	return hdlInstance
}

// CreateSupplier 	godoc
//
//	@Summary		Create new supplier based on parameters
//	@Description	Create new supplier based on parameters
//	@Tags			Suppliers
//	@Accept			json
//	@Param			supplier	body		CreateSupplierRequestDTO	true	"Supplier Data"
//	@Success		200		{object}	models.Supplier
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/suppliers [post]
//	@Security		Bearer
func (h *SupplierHandler) CreateSupplier(c *fiber.Ctx) error {
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

// GetAllSuppliers godoc
//
//	@Summary		Fetch all supplier
//	@Description	Fetch all supplier
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Supplier
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/suppliers	[get]
//	@Security		Bearer
func (h *SupplierHandler) GetAllSuppliers(c *fiber.Ctx) error {
	Suppliers, err := h.svc.GetAllSuppliers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(Suppliers)) + " records found",
			"data":    Suppliers,
		})
}

// GetSupplierById godoc
//
//	@Summary		Fetch individual supplier by Id
//	@Description	Fetch individual supplier by Id
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"supplier Id"
//	@Success		200					{object}	models.Supplier
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/suppliers/{id}	[get]
//	@Security		Bearer
func (h *SupplierHandler) GetSupplierById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
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

// UpdateSupplier godoc
//
//	@Summary		Update individual supplier
//	@Description	Update individual supplier
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"Supplier Id"
//	@Param			product				body		UpdateSupplierRequstDTO	true	"Supplier Data"
//	@Success		200					{object}	models.Supplier
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/suppliers/{id}	[put]
//	@Security		Bearer
func (h *SupplierHandler) UpdateSupplier(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
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
		ID:      foundSupplier.ID,
		Name:    input.Name,
		Address: input.Address,
		Phone:   input.Phone,
	}
	log.Println("updateCustomer: ", &updateSupplier)
	if err := c.BodyParser(&updateSupplier); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	result, err := h.svc.UpdateSupplier(&updateSupplier)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})
}

// DeleteSupplier godoc
//
//	@Summary		Delete individual supplier
//	@Description	Delete individual supplier
//	@Tags			Suppliers
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"supplier Id"
//	@Success		200					{object}	models.Supplier
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/suppliers/{id}	[delete]
//	@Security		Bearer
func (h *SupplierHandler) DeleteSupplier(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
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
			"status":  "FAIL",
			"message": "Internal server error",
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}
