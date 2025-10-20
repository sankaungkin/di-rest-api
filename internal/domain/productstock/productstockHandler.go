package productstock

import (
	"errors"
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

// CreateProductStocks godoc
//
//	@Summary		Create new product stock
//	@Description	Create a new product stock with productId, baseQty, derivedQty, and reorderLvl
//	@Tags			ProductStocks
//	@Accept			json
//	@Produce		json
//	@Param			productStock		body		models.ProductStock	true	"Product Stock Input Data"
//	@Success		200				{object}	models.ProductStock
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/productstocks [post]
//	@Security		Bearer
func (h *ProductStockHandler) CreateProductStocks(c *fiber.Ctx) error {
	input := new(models.ProductStock)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	log.Println("inputProductStock(Handler): ", input)

	productStockToCreate := &models.ProductStock{
		ProductId:    input.ProductId,
		BaseQty:      input.BaseQty,
		DerivedQty:   input.DerivedQty,
		ReorderLvl:   input.ReorderLvl,
		BaseUnitId:   input.BaseUnitId,
		DeriveUnitId: input.DeriveUnitId,
		// Add other fields as necessary
	}

	// Step 4: Create and return
	result, err := h.svc.CreateProductStocks(productStockToCreate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Create Successfully",
		"data":    result,
	})
}

// GetAllProductStocks godoc
//
//	@Summary		Get all product stocks
//	@Description	Get all product stocks
//	@Tags			ProductStocks
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

// GetLowStockProductCount godoc
//
//	@Summary		Get low stock products
//	@Description	Get low stock products
//	@Tags			ProductStocks
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	int64
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productstocks/low-stock-products [get]
//	@Security		Bearer
func (h *ProductStockHandler) GetLowStockProduct(c *fiber.Ctx) error {
	lowStockProductCount, err := h.svc.GetLowStockProducts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": " Record found",
		"data":    lowStockProductCount,
		"count":   len(lowStockProductCount),
	})
}

// GetOutOfStockProducts godoc
//
//	@Summary		Get out of stock products
//	@Description	Get out of stock products
//	@Tags			ProductStocks
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	models.ProductStock
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productstocks/out-of-stock-products [get]
//	@Security		Bearer
func (h *ProductStockHandler) GetOutOfStockProducts(c *fiber.Ctx) error {
	outOfStockProducts, err := h.svc.GetOutOfStockProducts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": " Record found",
		"data":    outOfStockProducts,
		"count":   len(outOfStockProducts),
	})
}

// GetProductStocksById godoc
//
//	@Summary		Fetch individual productstock by Id
//	@Description	Fetch individual productstock by Id
//	@Tags			ProductStocks
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"product Id"
//	@Success		200					{object}	ResponseProductStockDTO
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
//	@Tags			ProductStocks
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"product Id"
//	@Param			product				body		UpdateProductStockDTO	true	"ProductStock Data"
//	@Success		200					{object}	models.ProductStock
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
	foundProductStock, err := h.svc.UpdateProductStocksById(UpdateProductStockDTO{
		ProductID:  input.ProductID,
		DerivedQty: input.DerivedQty,
		ReorderLvl: input.ReorderLvl,
	})
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

	log.Println("updateProduct(Handler): ", foundProductStock)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    foundProductStock,
	})
}

// GetDetailsProductStockById godoc
//
//	@Summary		Get details of product stock
//	@Description	Get details of product stock
//	@Tags			ProductStocks
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"product Id"
//	@Success		200					{object}	DisplayStock
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productstocks/{id}	[get]
//	@Security		Bearer
func (h *ProductStockHandler) GetDetailsProductStockById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "ID is required",
		})
	}

	details, err := h.svc.GetDetailsProductStockById(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "No product stocks found for this ID",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "FAIL",
			"message": err.Error(),
		})
	}

	// Format response explicitly
	response := fiber.Map{
		"productId":   details.ProductId,
		"productName": details.ProductName,
		"quantity":    details.Quantity,
		"reorderlvl":  details.ReorderLvl,
		"units":       details.Units, // should be []DisplayStock or []map[string]interface{}
		"message":     "Record found",
		"status":      "SUCCESS",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// @Summary Get all product stocks with category
// @Description Get all product stocks with category
// @Tags ProductStock
// @Accept json
// @Produce json
// @Success 200 {object} ProductStockListInfoWithCategory
// @Router /product-stocks/all-with-category [get]
func (h *ProductStockHandler) GetAllProductStocksWithCategory(c *fiber.Ctx) error {
	productStocks, err := h.svc.GetAllProductStocksWithCategory()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": " Record found",
		"data":    productStocks,
		"count":   len(productStocks),
	})
}
