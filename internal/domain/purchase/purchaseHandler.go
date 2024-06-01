package purchase

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


type PurchaseHandler struct{
	svc PurchaseServiceInterface
}

var (
	hdlInstance *PurchaseHandler
	hdlOnce sync.Once
)


func NewSaleHandler(svc PurchaseServiceInterface) *PurchaseHandler {
	log.Println(util.Magenta + "SaleHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &PurchaseHandler{svc: svc}
	})
	return hdlInstance
}

func (h *PurchaseHandler)CreateSale(c *fiber.Ctx) error{
	
	input := new(PurchaseInvoiceRequestDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	newPurchase := models.Purchase{
		ID:          input.ID,
		SupplierId:  input.SupplierId,
		Discount:    input.Discount,
		GrandTotal:  input.GrandTotal,
		Remark:      input.Remark,
		PurchaseDate:    input.PurchaseDate,
		PurchaseDetails: input.PurchaseDetails,
		Total:       input.Total,
	}
	errors := models.ValidateStruct(newPurchase)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "operation failed",
		})
	}

	if _, err := h.svc.CreateService(&newPurchase); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data" : newPurchase,
		})
	
	
}

func(h *PurchaseHandler)GetAllPurchases(c *fiber.Ctx) error{

	purchases, err := h.svc.GetAllService()
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(purchases)) + " records found",
			"data":    purchases,
		})
}

func (h *PurchaseHandler)GetById(c *fiber.Ctx) error {

	purchase, err := h.svc.GetById(c.Params("id"))
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
		"data":    purchase,
	})

}