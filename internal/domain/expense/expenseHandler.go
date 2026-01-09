package expense

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type ExpenseHandler struct {
	svc ExpenseServiceInterface
}

var (
	hdlInstance *ExpenseHandler
	hdlOnce     sync.Once
)

func NewExpenseHandler(svc ExpenseServiceInterface) *ExpenseHandler {
	log.Println(util.Cyan + "ExpenseHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &ExpenseHandler{svc: svc}
	})
	return hdlInstance
}

// CreateExpense godoc
//
//	@Summary		Create a new expense record
//	@Description	Create a new expense record
//	@Tags			Expenses
//	@Accept			json
//	@Produce		json
//	@Param			expense	body		models.Expense	true	"Expense Data"
//	@Success		201		{object}	models.Expense
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Router			/api/expenses	[post]
//	@Security		Bearer
func (h *ExpenseHandler) CreateExpense(c *fiber.Ctx) error {
	var expense models.Expense

	if err := c.BodyParser(&expense); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	createdExpense, err := h.svc.CreateExpenseService(&expense)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "Expense record created successfully",
			"data":    createdExpense,
		})
}

// GetAllExpenses godoc
//
//	@Summary		Get all expense records
//	@Description	Get all expense records
//	@Tags			Expenses
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Expense
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/expenses	[get]
//	@Security		Bearer
func (h *ExpenseHandler) GetAllExpenses(c *fiber.Ctx) error {
	expenses, err := h.svc.GetAllExpensesService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(expenses)) + " records found",
			"data":    expenses,
		})
}
