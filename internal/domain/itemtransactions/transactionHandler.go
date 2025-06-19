package itemtransactions

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

type TransactionHandler struct {
	svc TransactionServiceInterface
}

var (
	hdlInstance *TransactionHandler
	hdlOnce     sync.Once
)

func NewTransactionHandler(svc TransactionServiceInterface) *TransactionHandler {
	log.Println(util.Green + "TransactionHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &TransactionHandler{svc: svc}
	})
	return hdlInstance
}

// GetAllTransactions godoc
//
//	@Summary		Fetch all transactions
//	@Description	Fetch all transactions
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.ItemTransaction
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/transactions	[get]
//	@Security		Bearer
func (h *TransactionHandler) GetAll(c *fiber.Ctx) error {
	newTransaction := models.ItemTransaction{}
	log.Println(newTransaction)
	transactions, err := h.svc.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(transactions)) + " records found",
			"data":    transactions,
			"count":   len(transactions),
		})
}

// GetTransactionsById godoc
//
//	@Summary		Fetch individual transaction by productId
//	@Description	Fetch individual transaction by productId
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"product Id"
//	@Success		200					{array}	models.ItemTransaction
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/transactions/by-product/{productId} [get]
//	@Security		Bearer
func (h *TransactionHandler) GetTransactionsByProductId(c *fiber.Ctx) error {
	productId := c.Params("productId")

	transactions, err := h.svc.GetByProductId(productId)
	if err != nil {
		if err == gorm.ErrRecordNotFound || len(transactions) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "No transactions found for the given product ID",
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Transactions retrieved successfully",
		"data":    transactions,
		"count":   len(transactions),
	})
}

// GetTransactionsByTransactionType godoc
//
//	@Summary		Fetch individual transaction by transactionType
//	@Description	Fetch individual transaction by protransactionType
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"transactionType"
//	@Success		200					{array}	models.ItemTransaction
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/transactions/by-type/{tranType} [get]
//	@Security		Bearer
func (h *TransactionHandler) GetTransactionsByTransactionType(c *fiber.Ctx) error {
	tranType := c.Params("tranType")

	transactions, err := h.svc.GetByTransactionType(tranType)
	if err != nil {
		if err == gorm.ErrRecordNotFound || len(transactions) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "No transactions found for the given product ID",
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Transactions retrieved successfully",
		"data":    transactions,
		"count":   len(transactions),
	})
}

// GetByProductIdAndTranType godoc
//
//	@Summary		Fetch individual transaction by productId and tranType
//	@Description	Fetch individual transaction by productId and tranType
//	@Tags			Transactions
//	@Accept			json
//	@Produce		json
//	@Param			productId					path		string	true	"Product ID"
//	@Param			tranType					path		string	true	"Transaction Type"
//	@Success		200					{array}	models.ItemTransaction
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/transactions/by-product-type/{productId}/{tranType} [get]
//	@Security		Bearer
func (h *TransactionHandler) GetByProductIdAndTranType(c *fiber.Ctx) error {
	tranType := c.Params("tranType")
	productId := c.Params("productId")

	transactions, err := h.svc.GetByProductIdAndTranType(productId, tranType)
	if err != nil {
		if err == gorm.ErrRecordNotFound || len(transactions) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "No transactions found for the given product ID",
			})
		}
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Transactions retrieved successfully",
		"data":    transactions,
		"count":   len(transactions),
	})
}

// CreateAdjustmentTransaction godoc
// @Summary      Create adjustment transaction
// @Description  Create
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        Authorization header   string                      true  "Bearer token"
// @Param        transaction      body      ResquestAdjustInventoryDTO     true  "Product input data"
// @Success      200          {object}  models.ItemTransaction
// @Failure      400          {object}  httputil.HttpError400
// @Failure      401          {object}  httputil.HttpError401
// @Failure      500          {object}  httputil.HttpError500
// @Router       /api/transactions/adjustment [post]
//
//	@Security		Bearer
func (h *TransactionHandler) CreateAdjustmentTransaction(c *fiber.Ctx) error {
	var transaction ResquestAdjustInventoryDTO
	if err := c.BodyParser(&transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Invalid request body",
		})
	}

	createdTransaction, err := h.svc.CreateAdjustmentTransaction(transaction)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Transaction created successfully",
		"data":    createdTransaction,
	})
}
