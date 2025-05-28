package product

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductHandler struct {
	svc ProductServiceInterface
}

// ! singleton pattern
var (
	hdlInstance *ProductHandler
	hdlOnce     sync.Once
)

// func NewProductHandler(srv ProductRepositoryInterface) *ProductHandler{
// 	return &ProductHandler{srv:srv}
// }

func NewProductHandler(svc ProductServiceInterface) *ProductHandler {
	log.Println(util.Yellow + "ProductHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &ProductHandler{svc: svc}
	})
	return hdlInstance
}

// CreateProduct 	godoc
//
//	@Summary		Create new product based on parameters
//	@Description	Create new product based on parameters
//	@Tags			Products
//	@Accept			json
//	@Param			product	body		CreateProductRequstDTO	true	"Product Data"
//	@Success		200		{object}	models.Product
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/product [post]
//
//	@Security		ApiKeyAuth
//
//	@param			Authorization	header	string	true	"Authorization"
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	input := new(CreateProductRequstDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	log.Println("New product input: ", input)
	newProduct := models.Product{
		ID:              input.ID,
		ProductName:     input.ProductName,
		CategoryId:      input.CategoryId,
		Uom:             input.Uom,
		BuyPrice:        input.BuyPrice,
		SellPriceLevel1: input.SellPriceLevel1,
		SellPriceLevel2: input.SellPriceLevel2,
		// ReorderLvl:      input.ReorderLvl,
		IsActive: input.IsActive,
	}

	err := c.BodyParser(&newProduct)
	if err != nil {
		c.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	errors := models.ValidateStruct(newProduct)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}
	log.Println("newProduct : ", newProduct)

	if _, err := h.svc.CreateSerive(&newProduct); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "new PRODUCT has been created successfully",
			"data":    newProduct,
		})

}

// GetProducts godoc
//
//	@Summary		Fetch all products
//	@Description	Fetch all products
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Product
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/product	[get]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *ProductHandler) GetAllProducts(c *fiber.Ctx) error {
	products, err := h.svc.GetAllSerive()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(products)) + " records found",
			"data":    products,
			"count":   len(products),
		})

}

// GetProductById godoc
//
//	@Summary		Fetch individual product by Id
//	@Description	Fetch individual product by Id
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"product Id"
//	@Success		200					{object}	models.Product
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/product/{id}	[get]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *ProductHandler) GetProductById(c *fiber.Ctx) error {

	product, err := h.svc.GetByIdSerive(c.Params("id"))
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
		"data":    product,
	})
}

func (h *ProductHandler) GetProductUnitPricesById(c *fiber.Ctx) error {
	productId := c.Params("id")
	if productId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Product ID is required",
		})
	}

	unitPrices, err := h.svc.GetProductUnitPricesByIdSerive(c.Params("id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "FAIL",
				"message": "No unit prices found for this product",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "FAIL", "message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": strconv.Itoa(len(unitPrices)) + " unit prices found",
		"data":    unitPrices,
		"count":   len(unitPrices),
	})
}

// UpdateProduct godoc
//
//	@Summary		Update individual product
//	@Description	Update individual product
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"product Id"
//	@Param			product				body		UpdateProductRequstDTO	true	"Product Data"
//	@Success		200					{object}	models.Product
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/product/{id}	[put]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {

	foundProduct, err := h.svc.GetByIdSerive(c.Params("id"))
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

	input := new(UpdateProductRequstDTO)
	log.Println("input: ", input)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	updateProduct := models.Product{
		ID:              foundProduct.ID,
		ProductName:     foundProduct.ProductName,
		CategoryId:      foundProduct.CategoryId,
		Uom:             foundProduct.Uom,
		BuyPrice:        foundProduct.BuyPrice,
		SellPriceLevel1: foundProduct.SellPriceLevel1,
		SellPriceLevel2: foundProduct.SellPriceLevel2,
		// ReorderLvl:      foundProduct.ReorderLvl,
		IsActive: foundProduct.IsActive,
	}
	log.Println("updateCustomer: ", &updateProduct)
	if err := c.BodyParser(&updateProduct); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	result, err := h.svc.Update(&updateProduct)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})
}

// DeleteProduct godoc
//
//	@Summary		Delete individual product
//	@Description	Delete individual product
//	@Tags			Products
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"product Id"
//	@Success		200					{object}	models.Product
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/product/{id}	[delete]
//
//	@Security		ApiKeyAuth
//
//	@Security		Bearer  <-----------------------------------------add this in all controllers that need authentication
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {

	id := strings.ToUpper(c.Params("id"))
	product, err := h.svc.GetByIdSerive(id)
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
	err = h.svc.DeleteSerive(product.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "FAIL",
			"message": "Internal server error",
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}
