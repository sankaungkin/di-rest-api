package cashbook

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type CashbookHandler struct {
	svc CashbookServiceInterface
}

var (
	hdlInstance *CashbookHandler
	hdlOnce     sync.Once
)

func NewCashbookHandler(svc CashbookServiceInterface) *CashbookHandler {
	log.Println(util.Cyan + "CashbookHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &CashbookHandler{svc: svc}
	})
	return hdlInstance
}

// GetLedger godoc
//
//	@Summary		Fetch all cashbook records
//	@Description	Fetch all cashbook records
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Param			startDate	query		string	true	"Start Date"	Format(date)
//	@Param			endDate		query		string	true	"End Date"		Format(date)
//	@Success		200				{array}		models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetLedger(c *fiber.Ctx) error {
	// 1. Support both "start" and "startDate" for flexibility
	startDate := c.Query("startDate")
	if startDate == "" {
		startDate = c.Query("start")
	}

	endDate := c.Query("endDate")
	if endDate == "" {
		endDate = c.Query("end")
	}

	// 2. Default to today if still empty instead of returning 400
	today := time.Now().Format("2006-01-02")
	if startDate == "" {
		startDate = today
	}
	if endDate == "" {
		endDate = today
	}

	// 3. Call service
	cashbooks, err := h.svc.GetLedger(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// 4. Return the standard wrapper
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Cashbook records fetched successfully",
		"data":    cashbooks,
	})
}

// GetBalance godoc
//
//	@Summary		Fetch current balance
//	@Description	Fetch current balance
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/balance	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetBalance(c *fiber.Ctx) error {
	balance, err := h.svc.GetBalance()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Balance fetched successfully",
			"data":    balance,
		})
}

// CloseDay godoc
//
//	@Summary		Close the cashbook for the day
//	@Description	Close the cashbook for the day
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/close-day	[post]
//	@Security		Bearer
func (h *CashbookHandler) CloseDay(c *fiber.Ctx) error {
	err := h.svc.CloseDay(time.Now())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Cashbook closed for the day successfully",
		})
}

// CreateEntry godoc
//
//	@Summary		Create a new cashbook entry
//	@Description	Create a new cashbook entry
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Param			entry	body		models.Cashbook	true	"Cashbook Entry"
//	@Success		201				{object}	models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/entry	[post]
//	@Security		Bearer
func (h *CashbookHandler) CreateEntry(c *fiber.Ctx) error {
	var entry models.Cashbook
	if err := c.BodyParser(&entry); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	fmt.Printf("Received Entry: Type=%s, Debit=%d, Credit=%d\n",
		entry.TransactionType, entry.Debit, entry.Credit)
	err := h.svc.CreateEntry(&entry)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Cashbook entry created successfully",
			"data":    entry,
		})
}

// GetSettlementReport godoc
//
//	@Summary		Generate monthly settlement report
//	@Description	Generate monthly settlement report
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Param			month	query		int	true	"Month"
//	@Param			year	query		int	true	"Year"
//	@Success		200				{array}	models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/settlement-report	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetSettlementReport(c *fiber.Ctx) error {
	month, err := strconv.Atoi(c.Query("month"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid month parameter"})
	}
	year, err := strconv.Atoi(c.Query("year"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid year parameter"})
	}

	reports, err := h.svc.GetSettlementReport(month, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Cashbook settlement report generated successfully",
			"data":    reports,
		})
}

// GetAllCashbooks godoc
//
//	@Summary		Fetch all cashbook records within a date range
//	@Description	Fetch all cashbook records within a date range
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Param			startDate	query		string	true	"Start Date"	Format(date)
//	@Param			endDate		query		string	true	"End Date"		Format(date)
//	@Success		200				{array}		models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/all	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetAllCashbooks(c *fiber.Ctx) error {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
	}

	cashbooks, err := h.svc.GetAll(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Cashbook records fetched successfully",
			"data":    cashbooks,
		})
}

// GetTransactionHistory godoc
//
//	@Summary		Fetch transaction history
//	@Description	Fetch transaction history
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Param			startDate	query		string	true	"Start Date"	Format(date)
//	@Param			endDate		query		string	true	"End Date"		Format(date)
//	@Success		200				{object}	models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/transaction-history	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetTransactionHistory(c *fiber.Ctx) error {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format"})
	}

	cashbook, err := h.svc.GetTransactionHistory(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Transaction history fetched successfully",
			"data":    cashbook,
		})
}

// GetCashBalance godoc
//
//	@Summary		Fetch current cash balance
//	@Description	Fetch current cash balance
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/balance	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetCashBalance(c *fiber.Ctx) error {
	balance, err := h.svc.GetCashBalance()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Cash balance fetched successfully",
			"data":    balance,
		})
}

// GetDashboardSummary godoc
//
//	@Summary		Fetch dashboard summary
//	@Description	Fetch dashboard summary
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.DashboardSummary
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/summary	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetDashboardSummary(c *fiber.Ctx) error {
	summary, err := h.svc.GetDashboardSummary()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Dashboard summary fetched successfully",
			"data":    summary,
		})
}

// GetCurrentDrawerBalance godoc
//
//	@Summary		Fetch current drawer balance
//	@Description	Fetch current drawer balance
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	int64
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/balance	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetCurrentDrawerBalance(c *fiber.Ctx) error {
	balance, err := h.svc.GetCurrentDrawerBalance()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Current drawer balance fetched successfully",
			"data":    balance,
		})
}

// GetPastSummaries godoc
//
//	@Summary		Fetch past daily summaries
//	@Description	Fetch past daily summaries
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.DailySummaries
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/past-summaries	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetPastSummaries(c *fiber.Ctx) error {
	summaries, err := h.svc.GetPastSummaries()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Past daily summaries fetched successfully",
			"data":    summaries,
		})
}

// ReconcileToday godoc
//
//	@Summary		Reconcile today's cashbook
//	@Description	Reconcile today's cashbook
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		map[string]interface{}
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/reconcile-today	[get]
//	@Security		Bearer
func (h *CashbookHandler) ReconcileToday(c *fiber.Ctx) error {
	results, err := h.svc.ReconcileToday()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Reconcile today's cashbook fetched successfully",
			"data":    results,
		})
}

// @Summary Record owner withdrawal
// @Description Record owner withdrawal
// @Tags Cashbook
// @Accept json
// @Produce json
// @Param amount body int true "Amount"
// @Param description body string true "Description"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} httputil.HttpError400
// @Failure 401 {object} httputil.HttpError401
// @Failure 500 {object} httputil.HttpError500
// @Router /api/cashbook/owner-withdrawal [post]
func (h *CashbookHandler) RecordOwnerWithdrawal(c *fiber.Ctx) error {
	var payload struct {
		Amount      int64  `json:"amount"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := h.svc.RecordOwnerWithdrawal(payload.Amount, payload.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Owner withdrawal recorded successfully",
			"data":    nil,
		})
}

// GetTodayEntries godoc
//
//	@Summary		Fetch today's cashbook entries
//	@Description	Fetch today's cashbook entries
//	@Tags			Cashbooks
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Cashbook
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/cashbook/today-entries	[get]
//	@Security		Bearer
func (h *CashbookHandler) GetTodayOwnerWithdrawalEntries(c *fiber.Ctx) error {
	// today := time.Now().Format("2006-01-02")
	transType := "OWNER_WITHDRAWAL"

	entries, err := h.svc.GetEntriesByDateAndType(time.Now(), transType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Today's cashbook entries fetched successfully",
			"data":    entries,
		})
}
