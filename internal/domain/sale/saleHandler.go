package sale

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
		ID:            input.ID,
		CustomerId:    input.CustomerId,
		Discount:      input.Discount,
		GrandTotal:    input.GrandTotal,
		PaidAmount:    input.PaidAmount,
		Remark:        input.Remark,
		SaleDate:      input.SaleDate,
		PaymentMethod: input.PaymentMethod,
		SaleDetails:   input.SaleDetails,
		Total:         input.Total,
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

// GetSaleStockItemWithPrice godoc
//
//	@Summary		Fetch sale stock item with price
//	@Description	Fetch sale stock item with price
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		ResponseSaleStockItemWithPrice
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/sales/sale-stock-item-with-price	[get]
//	@Security		Bearer
func (h *SaleHandler) GetSaleStockItemWithPrice(c *fiber.Ctx) error {

	results, err := h.svc.GetSaleStockItemWithPrice()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Sale stock item with price fetched successfully",
			"data":    results,
		})
}

// UpdateSaleRemark godoc
//
//	@Summary		Update individual sale
//	@Description	Update individual sale
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"sale Id"
//	@Param			sale				body		UpdateSaleRemarkDTO	true	"Sale Data"
//	@Success		200					{object}	models.Sale
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/sales/{id}	[put]
//
//	@Security		Bearer
func (h *SaleHandler) UpdateSaleRemark(c *fiber.Ctx) error {
	id := c.Params("id")

	// Step 1: Parse incoming payload
	input := new(UpdateSaleRemarkDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	log.Println("inputProduct(Handler): ", input)

	// Step 2: Call service update (only using payload)
	result, err := h.svc.UpdateSale(UpdateSaleRemarkDTO{
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

// @Summary Get Today's Sale List
// @Description Get Today's Sale List
// @Tags Sale
// @Accept json
// @Produce json
// @Success 200 {object} fiber.Map{data=[]models.Sale}
// @Router /api/v1/sales/today-list [get]
func (h *SaleHandler) GetTodaySaleList(c *fiber.Ctx) error {
	results, err := h.svc.GetTodaySaleList()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Today's Sale List fetched successfully",
			"data":    results,
		})
}

// GetSalesByDate godoc
//
//	@Summary		Fetch sales by date
//	@Description	Fetch sales by date
//	@Tags			Sales
//	@Accept			json
//	@Produce		json
//	@Param			date					path		string	true	"sale date"
//	@Success		200				{array}		models.Sale
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/sales/sales-by-date/{date}	[get]
//	@Security		Bearer
func (h *SaleHandler) GetSalesByDate(c *fiber.Ctx) error {
	date := c.Params("date")
	if date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Date is required",
		})
	}

	results, err := h.svc.GetSalesByDate(time.Now())
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Sales by date fetched successfully",
			"data":    results,
		})
}

// @Summary Return sale items
// @Description Return sale items
// @Tags Sales
// @Accept json
// @Produce json
// @Param saleID path string true "Sale ID"
// @Param returnItems body []models.SaleDetail true "Return items"
// @Param remark body string true "Remark"

// @Success 200 {object} fiber.Map{data=models.Sale}
// @Router /api/v1/sales/return-items [post]
func (h *SaleHandler) ReturnSaleItems(c *fiber.Ctx) error {
	var dto SaleReturnDTO

	if err := c.BodyParser(&dto); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format: " + err.Error(),
		})
	}

	result, err := h.svc.ReturnSaleItems(dto)
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Sale items returned successfully",
		"data":    result,
	})
}

// @Summary Get Historical Profit Data
// @Description Get Historical Profit Data
// @Tags Sales
// @Accept json
// @Produce json
// @Success 200 {object} fiber.Map{data=[]models.Sale}
// @Router /api/v1/sales/historical-profit-data [get]
func (h *SaleHandler) GetHistoricalProfitData(c *fiber.Ctx) error {
	results, err := h.svc.GetHistoricalProfitData()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Historical Profit Data fetched successfully",
			"data":    results,
		})
}

// @Summary Collect Debt
// @Description Collect Debt for a specific sale invoice
// @Tags Sales
// @Accept json
// @Produce json
// @Param id path string true "Sale ID"
// @Param payment body models.Payment true "Payment Data"
// @Success 200 {object} fiber.Map
// @Router /api/v1/sales/{id}/collect-debt [post]
func (h *SaleHandler) CollectDebt(c *fiber.Ctx) error {
	saleID := c.Params("id")
	var payment models.PaymentRecord

	if err := c.BodyParser(&payment); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": err.Error()})
	}

	// ✅ Fix 1: Generate a unique numeric ID for the payment record (uint)
	payment.ID = uint(time.Now().UnixNano())

	// ✅ Fix 2: Set the current time if not provided by frontend
	if payment.PaymentDate.IsZero() {
		payment.PaymentDate = time.Now()
	}

	if err := h.svc.CollectDebt(&payment, saleID); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "FAIL", "message": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"status": "SUCCESS"})
}

// @Summary Get Sales With Receivables
// @Description Get Sales With Receivables
// @Tags Sales
// @Accept json
// @Produce json
// @Success 200 {object} fiber.Map{data=[]models.Sale}
// @Router /api/v1/sales/with-receivables [get]
func (h *SaleHandler) GetSalesWithReceivables(c *fiber.Ctx) error {
	results, err := h.svc.GetSalesWithReceivables()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(results)) + " records found",
			"data":    results,
			"count":   len(results),
		})
}

func (h *SaleHandler) GetPaymentHistory(c *fiber.Ctx) error {
	saleID := c.Params("id")
	if saleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Sale ID is required",
		})
	}

	results, err := h.svc.GetPaymentHistory(saleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(results)) + " records found",
			"data":    results,
			"count":   len(results),
		})

}
