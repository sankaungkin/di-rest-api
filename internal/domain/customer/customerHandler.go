package customer

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	svc CustomerServiceInterface
}
//! singleton pattern
var (
	hdlInstance *CustomerHandler
	hdlOnce sync.Once
)
// constructor
func NewCustomerHandler(svc CustomerServiceInterface) *CustomerHandler{
	log.Println(Magenta + "CustomerHandler constructor is called " +Reset)
	hdlOnce.Do(func(){
		hdlInstance = &CustomerHandler{svc: svc}
	})
	return hdlInstance
}
// func NewCustomerHandler(svc CustomerServiceInterface) *CustomerHandler{
// 	return &CustomerHandler{srv: svc}
// }

func(h *CustomerHandler)CreateCustomer(c *fiber.Ctx) error {
	newCustomer := new(models.Customer)
	if err := c.BodyParser(newCustomer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	errors := models.ValidateStruct(newCustomer)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	if _, err := h.svc.CreateCustomer(newCustomer); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data":    newCustomer,
		})
}

func(h *CustomerHandler) GetAllCustomers(c *fiber.Ctx) error {
	customers, err := h.svc.GetAllCustomers()
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(customers)) + " records found",
			"data":    customers,
		})
}

func(h *CustomerHandler) GetCustomerById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}

	customer, err := h.svc.GetCustomerById(uint(id))
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
		"data":    customer,
	})
}

func(h *CustomerHandler)UpdateCustomer(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	customer, err := h.svc.GetCustomerById(uint(id))
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
	h.svc.UpdateCustomer(customer)
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Update successfully",
	})
}

func(h *CustomerHandler)DeleteCustomer(c *fiber.Ctx) error{
	id, err := strconv.ParseUint(c.Params("id"), 10,32)
	if err != nil {
		log.Fatal(err)
	}
	customer, err := h.svc.GetCustomerById(uint(id))
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
	h.svc.DeleteCustomer(uint(customer.ID))
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}