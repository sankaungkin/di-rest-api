package purchase

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

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

// GetById godoc
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

// GetTodayPurchases godoc
//
//	@Summary		Fetch today's purchases
//	@Description	Fetch today's purchases
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Purchase
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/purchases/today	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetTodayPurchases(c *fiber.Ctx) error {

	purchases, err := h.svc.GetTodayPurchases()
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

// GetTodayGrandTotal godoc
//
//	@Summary		Fetch today's grand total
//	@Description	Fetch today's grand total
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	int64
//	@Failure		400					{object}	httputil.HttpError400
//	@Router			/api/purchases/today-grand-total	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetTodayGrandTotal(c *fiber.Ctx) error {
	grandTotal, err := h.svc.GetTodayGrandTotal()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Today's grand total fetched successfully",
		"data":    grandTotal,
	})
}

// GetMonthlyPurchases godoc
//
//	@Summary		Fetch monthly purchases
//	@Description	Fetch monthly purchases
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Purchase
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/purchases/monthly	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetMonthlyPurchases(c *fiber.Ctx) error {

	purchases, err := h.svc.GetMonthlyPurchases()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(purchases)) + " records found",
			"data":    purchases,
			"count":   len(purchases),
		})
}

// GetMonthlyGrandTotal godoc
//
//	@Summary		Fetch monthly grand total
//	@Description	Fetch monthly grand total
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	int64
//	@Failure		400					{object}	httputil.HttpError400
//	@Router			/api/purchases/monthly-grand-total	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetMonthlyGrandTotal(c *fiber.Ctx) error {
	grandTotal, err := h.svc.GetMonthlyGrandTotal()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Monthly grand total fetched successfully",
		"data":    grandTotal,
	})
}

// GetPurchaseLineItems godoc
//
//	@Summary		Fetch purchase line items
//	@Description	Fetch purchase line items
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		ResponsePurchaseLineItemDTO
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/purchases/line-items	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetPurchaseLineItems(c *fiber.Ctx) error {

	results, err := h.svc.GetPurchaseLineItems()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Purchase line items fetched successfully",
			"data":    results,
		})
}

// UpdatePurchaseRemark godoc
//
//	@Summary		Update individual purchase
//	@Description	Update individual purchase
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"purchase Id"
//	@Param			purchase				body		UpdateRemarkPurchaseDTO	true	"Purchase Data"
//	@Success		200					{object}	models.Purchase
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/purchases/{id}	[put]
//	@Security		Bearer
func (h *PurchaseHandler) UpdatePurchaseRemark(c *fiber.Ctx) error {
	id := c.Params("id")
	input := new(UpdateRemarkPurchaseDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	log.Println("inputProduct(Handler): ", input)

	// Step 2: Call service update (only using payload)
	result, err := h.svc.UpdatePurchaseRemark(UpdateRemarkPurchaseDTO{
		ID:     id,
		Remark: input.Remark,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "Record not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}

	// Step 3: Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})
}

// GetTodayPurchaseList godoc
//
//	@Summary		Fetch today's purchase list
//	@Description	Fetch today's purchase list
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Purchase
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/purchases/today-list	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetTodayPurchaseList(c *fiber.Ctx) error {
	results, err := h.svc.GetTodayPurchaseList()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Today's Purchase List fetched successfully",
			"data":    results,
		})
}

// GetPurchasesByDate godoc
//
//	@Summary		Fetch purchases by date
//	@Description	Fetch purchases by date
//	@Tags			Purchases
//	@Accept			json
//	@Produce		json
//	@Param			date					path		string	true	"purchase date"
//	@Success		200				{array}		models.Purchase
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/purchases/purchases-by-date/{date}	[get]
//	@Security		Bearer
func (h *PurchaseHandler) GetPurchasesByDate(c *fiber.Ctx) error {
	date := c.Params("date")
	if date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Date is required",
		})
	}

	results, err := h.svc.GetPurchasesByDate(time.Now())
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Purchases by date fetched successfully",
			"data":    results,
		})
}
