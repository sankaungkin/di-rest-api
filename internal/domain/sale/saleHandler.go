package sale

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

type SaleHandler struct{
	svc SaleServiceInterface
}

var (
	hdlInstance *SaleHandler
	hdlOnce sync.Once
)


func NewSaleHandler(svc SaleServiceInterface) *SaleHandler {
	log.Println(util.Blue + "SaleHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &SaleHandler{svc: svc}
	})
	return hdlInstance
}

func (h *SaleHandler)CreateSale(c *fiber.Ctx) error{
	
	input := new(SaleInvoiceRequestDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	newSale := models.Sale{
		ID:          input.ID,
		CustomerId:  input.CustomerId,
		Discount:    input.Discount,
		GrandTotal:  input.GrandTotal,
		Remark:      input.Remark,
		SaleDate:    input.SaleDate,
		SaleDetails: input.SaleDetails,
		Total:       input.Total,
	}
	errors := models.ValidateStruct(newSale)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "operation failed",
		})
	}

	if _, err := h.svc.CreateService(&newSale); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data" : newSale,
		})
	
	
}
//TODO implement get sales process
func(h *SaleHandler)GetAllSales(c *fiber.Ctx) error{

	sales, err := h.svc.GetAllService()
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(sales)) + " records found",
			"data":    sales,
		})
}

func (h *SaleHandler)GetById(c *fiber.Ctx) error {

	sale, err := h.svc.GetById(c.Params("id"))
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
		"data":    sale,
	})

}