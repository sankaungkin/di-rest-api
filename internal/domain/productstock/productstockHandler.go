package productstock

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

type ProductStockHandler struct {
	svc ProductStockRepositoryInterface
}

// ! singleton pattern
var (
	handlerInstance *ProductStockHandler
	handlerOnce     sync.Once
)

// func NewProductStockHandler(repo ProductStockRepositoryInterface) ProductStockHandlerInterface{
// 	return &ProductStockHandler{repo: repo}
// }
//! constructor must be return the Interface, NOT struct, if not, google wire generate fail

func NewProductStockHandler(svc ProductStockRepositoryInterface) *ProductStockHandler {

	log.Println(util.Yellow + "ProductStockHandler constructor is called" + util.Reset)

	handlerOnce.Do(func() {
		handlerInstance = &ProductStockHandler{svc: svc}
	})
	return handlerInstance
}

// GetAllProductStocks godoc
//
//	@Summary		Get all product stocks
//	@Description	Get all product stocks
//	@Tags			Productstocks
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	models.ProductStock
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productstocks [get]
//	@Security		Bearer
func (h *ProductStockHandler) GetAllProductStocks(c *fiber.Ctx) error {
	productStocks, err := h.svc.GetAllProductStocks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(productStocks)) + " records found",
			"data":    productStocks,
			"count":   len(productStocks),
		})
}

// GetProductStocksById godoc
//
//	@Summary		Fetch individual productstock by Id
//	@Description	Fetch individual productstock by Id
//	@Tags			Productstocks
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"product Id"
//	@Success		200					{object}	models.Productstock
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productstocks/{id} [get]
//	@Security		Bearer
func (h *ProductStockHandler) GetProductStocksById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "ID is required",
		})
	}

	productStocks, err := h.svc.GetProductStocksById(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "No product stocks found for this ID",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "FAIL", "message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": " Record found",
		"data":    productStocks,
	})
}

// UpdateProductStocksById godoc
//
//	@Summary		Update individual productstock
//	@Description	Update individual products
//	@Tags			Productstocks
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"product Id"
//	@Param			product				body		UpdateProductStockDTO	true	"ProductStock Data"
//	@Success		200					{object}	models.Productstock
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productstocks/{id}	[put]
//
//	@Security		Bearer
func (h *ProductStockHandler) UpdateProductStocksById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "ID is required",
		})
	}

	input := new(UpdateProductStockDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	log.Println("inputProduct(Handler): ", input)

	// Step 3: Manually update only intended fields
	foundProductStock, err := h.svc.GetProductStocksById(id)
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
	foundProductStock.BaseQty = input.BaseQty
	foundProductStock.DerivedQty = input.DerivedQty
	foundProductStock.ReorderLvl = input.ReorderLvl

	log.Println("updateProduct(Handler): ", foundProductStock)

	// Map foundProductStock (*ResponseProductStockDTO) to *models.ProductStock
	productStockToUpdate := &models.ProductStock{
		ProductId:  foundProductStock.ProductID,
		BaseQty:    foundProductStock.BaseQty,
		DerivedQty: foundProductStock.DerivedQty,
		ReorderLvl: foundProductStock.ReorderLvl,
		// Add other fields as necessary
	}

	// Step 4: Update and return
	result, err := h.svc.UpdateProductStocksById(productStockToUpdate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})
}
