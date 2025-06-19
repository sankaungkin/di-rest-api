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

type PurchaseHandler struct {
	svc PurchaseServiceInterface
}

var (
	hdlInstance *PurchaseHandler
	hdlOnce     sync.Once
)

func NewSaleHandler(svc PurchaseServiceInterface) *PurchaseHandler {
	log.Println(util.Magenta + "SaleHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &PurchaseHandler{svc: svc}
	})
	return hdlInstance
}

// CreatePurchase 	godoc
//
//	@Summary		Create new purchase based on parameters
//	@Description	Create new purchase based on parameters
//	@Tags			Purchases
//	@Accept			json
//	@Param			purchase	body		PurchaseInvoiceRequestDTO	true	"Product Data"
//	@Success		200		{object}	models.Purchase
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/purchases [post]
//	@Security		Bearer
func (h *PurchaseHandler) CreatePurchase(c *fiber.Ctx) error {

	input := new(PurchaseInvoiceRequestDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	newPurchase := models.Purchase{
		ID:              input.ID,
		SupplierId:      input.SupplierId,
		Discount:        input.Discount,
		GrandTotal:      input.GrandTotal,
		Remark:          input.Remark,
		PurchaseDate:    input.PurchaseDate,
		PurchaseDetails: input.PurchaseDetails,
		Total:           input.Total,
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
			"message": "purchase operation has been created successfully",
			"data":    newPurchase,
		})

}

// GetAllPurchases godoc
//
//	@Summary		Fetch all purchases
//	@Description	Fetch all purchases
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Purchase
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/purchases	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetAllPurchases(c *fiber.Ctx) error {

	purchases, err := h.svc.GetAllService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(purchases)) + " records found",
			"data":    purchases,
			"count":   len(purchases),
		})
}

// GetPurchaseById godoc
//
//	@Summary		Fetch individual purchase by Id
//	@Description	Fetch individual purchase by Id
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"purchase Id"
//	@Success		200					{object}	models.Purchase
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/purchases/{id}	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetById(c *fiber.Ctx) error {

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
