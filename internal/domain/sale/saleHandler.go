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

type SaleHandler struct {
	svc SaleServiceInterface
}

var (
	hdlInstance *SaleHandler
	hdlOnce     sync.Once
)

func NewSaleHandler(svc SaleServiceInterface) *SaleHandler {
	log.Println(util.Blue + "SaleHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &SaleHandler{svc: svc}
	})
	return hdlInstance
}

// CreateSale 	godoc
//
//	@Summary		Create new sale based on parameters
//	@Description	Create new sale based on parameters
//	@Tags			Sales
//	@Accept			json
//	@Param			sale	body		SaleInvoiceRequestDTO	true	"Product Data"
//	@Success		200		{object}	models.Sale
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/sales [post]
//	@Security		Bearer
func (h *SaleHandler) CreateSale(c *fiber.Ctx) error {

	input := new(SaleInvoiceRequestDTO)
	log.Println("input", input)
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
			"message": "Sale has been created successfully",
			"data":    newSale,
		})

}

// GetAllSales godoc
//
//	@Summary		Fetch all sales
//	@Description	Fetch all sales
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Sale
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/sales	[get]
//	@Security		Bearer
func (h *SaleHandler) GetAllSales(c *fiber.Ctx) error {

	sales, err := h.svc.GetAllService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(sales)) + " records found",
			"data":    sales,
			"count":   len(sales),
		})
}

// GetTodaySales godoc
//
//	@Summary		Fetch today's sales
//	@Description	Fetch today's sales
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Sale
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/sales/today	[get]
//	@Security		Bearer
func (h *SaleHandler) GetTodaySales(c *fiber.Ctx) error {

	sales, err := h.svc.GetTodaySales()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(sales)) + " records found",
			"data":    sales,
			"count":   len(sales),
		})
}

// GetById godoc
//
//	@Summary		Fetch individual sale by Id
//	@Description	Fetch individual sale by Id
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"sale Id"
//	@Success		200					{object}	models.Sale
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/sales/{id}	[get]
//	@Security		Bearer
func (h *SaleHandler) GetById(c *fiber.Ctx) error {

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

	if sale == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Record not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Record found",
		"data":    sale,
	})

}

// GetTodayGrandTotal godoc
//
//	@Summary		Fetch today's grand total
//	@Description	Fetch today's grand total
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	int64
//	@Failure		400					{object}	httputil.HttpError400
//	@Router			/api/sales/today-grand-total	[get]
//	@Security		Bearer
func (h *SaleHandler) GetTodayGrandTotal(c *fiber.Ctx) error {
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

// GetMonthlySales godoc
//
//	@Summary		Fetch monthly sales
//	@Description	Fetch monthly sales
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Sale
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/sales/monthly	[get]
//	@Security		Bearer
func (h *SaleHandler) GetMonthlySales(c *fiber.Ctx) error {

	sales, err := h.svc.GetMonthlySales()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(sales)) + " records found",
			"data":    sales,
			"count":   len(sales),
		})
}

// GetMonthlyGrandTotal godoc
//
//	@Summary		Fetch monthly grand total
//	@Description	Fetch monthly grand total
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	int64
//	@Failure		400					{object}	httputil.HttpError400
//	@Router			/api/sales/monthly-grand-total	[get]
//	@Security		Bearer
func (h *SaleHandler) GetMonthlyGrandTotal(c *fiber.Ctx) error {
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

// TopCustomers godoc
//
//	@Summary		Fetch top customers
//	@Description	Fetch top customers
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	ResponseTopCustomerDTO
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/sales/top-customers	[get]
//	@Security		Bearer
func (h *SaleHandler) TopCustomers(c *fiber.Ctx) error {

	customers, err := h.svc.GetTopCustomers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Top customers fetched successfully",
		"data":    customers,
	})
}

// GetDailySales godoc
//
//	@Summary		Fetch daily sales
//	@Description	Fetch daily sales
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		ResponseDailySalesDTO
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/sales/daily	[get]
//	@Security		Bearer
func (h *SaleHandler) GetDailySales(c *fiber.Ctx) error {

	sales, err := h.svc.GetDailySales()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(sales)) + " records found",
			"data":    sales,
			"count":   len(sales),
		})
}

// @Summary Get Top Ten Sole Products
// @Description Get Top Ten Sole Products
// @Tags Sale
// @Accept json
// @Produce json
// @Success 200 {object} fiber.Map{data=[]ResponseTopTenSoleProductsDTO}
// @Router /api/v1/sales/toptensoleproducts [get]
func (h *SaleHandler) GetTopTenSoleProducts(c *fiber.Ctx) error {
	results, err := h.svc.GetTopTenSoleProducts()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Top Ten Sole Products fetched successfully",
			"data":    results,
		})
}
