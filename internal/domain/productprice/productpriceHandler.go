package productprice

import (
	"log"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductPriceHandler struct {
	svc ProductPriceServiceInterface
}

var (
	hdlInstance *ProductPriceHandler
	hdlOnce     sync.Once
)

func NewProductPriceHandler(svc ProductPriceServiceInterface) *ProductPriceHandler {
	log.Println(util.Yellow + "ProductPriceHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &ProductPriceHandler{svc: svc}
	})
	return hdlInstance
}

// CreateProductPrice godoc
//
//	@Summary		Create new product price
//	@Description	Create a new product price with productId, unitId, and unitPrice
//	@Tags			ProductPrice
//	@Accept			json
//	@Produce		json
//	@Param			productPrice		body		models.ProductPrice	true	"Product Price Input Data"
//	@Success		200				{object}	models.ProductPrice
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/productprices [post]
//	@Security		Bearer
func (h *ProductPriceHandler) CreateProductPrice(c *fiber.Ctx) error {
	input := new(models.ProductPrice)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	log.Println("inputProductPrice(Handler): ", input)
	newProductPrice := models.ProductPrice{
		ID:        input.ID,
		ProductId: input.ProductId,
		UnitId:    input.UnitId,
		UnitPrice: input.UnitPrice,
		PriceType: input.PriceType,
	}

	errors := models.ValidateStruct(newProductPrice)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	log.Println("newProductPrice : ", newProductPrice)

	if _, err := h.svc.Create(&newProductPrice); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "new PRODUCT PRICE has been created successfully",
			"data":    newProductPrice,
		})

}

// GetAllProductPrices godoc
//
//	@Summary		Fetch all product prices
//	@Description	Fetch all product prices
//	@Tags			ProductPrice
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	[]models.ProductPrice
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/productprices [get]
//	@Security		Bearer
func (h *ProductPriceHandler) GetAllProductPrices(c *fiber.Ctx) error {
	productPrices, err := h.svc.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(productPrices)) + " records found",
			"data":    productPrices,
			"count":   len(productPrices),
		})
}

//GetAllProductPricesWithStock godoc
//
//	@Summary		Fetch all product prices with stock

// @Summary		Fetch all product prices with stock
// @Description	Fetch all product prices with stock
// @Tags			ProductPrice
// @Accept			json
// @Produce		json
// @Success		200					{object}	[]models.ProductPrice
// @Failure		400					{object}	httputil.HttpError400
// @Failure		401					{object}	httputil.HttpError401
// @Failure		500					{object}	httputil.HttpError500
// @Router			/api/productprices/with-stock [get]
// @Security		Bearer
func (h *ProductPriceHandler) GetAllProductPricesWithStock(c *fiber.Ctx) error {
	productPrices, err := h.svc.GetAllWithStock()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(productPrices)) + " records found",
			"data":    productPrices,
			"count":   len(productPrices),
		})
}

// GetInventoryTotalValue godoc
//
//	@Summary		Fetch inventory total value
//	@Description	Fetch inventory total value
//	@Tags			ProductPrice
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	int64
//	@Failure		400					{object}	httputil.HttpError400
//	@Router			/api/productprices/inventory-total-value [get]
func (h *ProductPriceHandler) GetInventoryTotalValue(c *fiber.Ctx) error {
	inventoryTotalValue, err := h.svc.GetInventoryTotalValue()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Inventory total value fetched successfully",
		"data":    inventoryTotalValue,
	})
}

// GetProductPriceById godoc
//
//	@Summary		Fetch individual product price by Id
//	@Description	Fetch individual product price by Id
//	@Tags			ProductPrice
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"product price Id"
//	@Success		200					{object}	models.ProductPrice
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productprices/{id} [get]
//	@Security		Bearer
func (h *ProductPriceHandler) GetProductPriceById(c *fiber.Ctx) error {

	productPriceIdStr := c.Params("id")
	// productPriceId, err := strconv.Atoi(productPriceIdStr)
	// if err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"status":  "FAIL",
	// 		"message": "Invalid product price ID",
	// 	})
	// }
	productPrice, err := h.svc.GetById(productPriceIdStr)
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
		"data":    productPrice,
	})
}

// UpdateProductPrice godoc
//
//	@Summary		Update individual product price
//	@Description	Update individual product price
//	@Tags			ProductPrice
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"product price Id"
//	@Param			productPrice		body		UpdateProductPriceRequestDTO	true	"Product Price Data"
//	@Success		200					{object}	models.ProductPrice
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/productprices/{id}	[put]
//	@Security		Bearer
//TODO: update product price
// func (h *ProductPriceHandler) UpdateProductPrice(c *fiber.Ctx) error {
// 	// idUint64, err := strconv.ParseUint(c.Params("id"), 10, 32)
// 	// if err != nil {
// 	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 	// 		"status":  "FAIL",
// 	// 		"message": "Invalid product price ID",
// 	// 	})
// 	// }
// 	// id := int(idUint64)
// 	productId := c.Params("id")
// 	// Step 1: Get the existing product
// 	foundProductPrice, err := h.svc.GetById(productId)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 				"status":  "FAIL",
// 				"message": "Record not found",
// 			})
// 		}
// 		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
// 			"status": "FAIL", "message": err.Error(),
// 		})
// 	}
// 	input := new(UpdateProductPriceRequestDTO)
// 	if err := c.BodyParser(input); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"status":  400,
// 			"message": "Invalid JSON format",
// 		})
// 	}
// 	log.Println("inputProduct(Handler): ", input)

// 	// Step 3: Manually update only intended fields
// 	foundProductPrice.ProductId = input.ProductId
// 	foundProductPrice.UnitId = input.UnitId
// 	foundProductPrice.UnitPrice = input.UnitPrice

// 	log.Println("updateProduct(Handler): ", foundProductPrice)

// 	// Convert foundProductPrice to *models.ProductPrice
// 	updatedProductPrice := &models.ProductPrice{
// 		ID:        foundProductPrice.ID,
// 		ProductId: foundProductPrice.ProductId,
// 		UnitId:    foundProductPrice.UnitId,
// 		UnitPrice: foundProductPrice.UnitPrice,
// 	}

// 	// Step 4: Update and return
// 	result, err := h.svc.UpdateProductPrice(updatedProductPrice)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  "FAIL",
// 			"message": err.Error(),
// 		})
// 	}
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  "SUCCESS",
// 		"message": "Update Successfully",
// 		"data":    result,
// 	})
// }

// GetAllNew godoc
//
//	@Summary		Fetch all product prices with stock

// @Summary		Fetch all product prices with stock
// @Description	Fetch all product prices with stock
// @Tags			ProductPrice
// @Accept			json
// @Produce		json
// @Success		200					{object}	[]models.ProductPrice
// @Failure		400					{object}	httputil.HttpError400
// @Failure		401					{object}	httputil.HttpError401
// @Failure		500					{object}	httputil.HttpError500
// @Router			/api/productprices/with-stock [get]
// @Security		Bearer
func (h *ProductPriceHandler) GetAllNew(c *fiber.Ctx) error {
	productPrices, err := h.svc.GetAllNew()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(productPrices)) + " records found",
			"data":    productPrices,
			"count":   len(productPrices),
		})
}
