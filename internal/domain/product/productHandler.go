package product

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type ProductHandler struct {
	svc ProductServiceInterface
}
//! singleton pattern
var(
	hdlInstance *ProductHandler
	hdlOnce sync.Once
)

// func NewProductHandler(srv ProductRepositoryInterface) *ProductHandler{
// 	return &ProductHandler{srv:srv}
// }

func NewProductHandler(svc ProductServiceInterface) *ProductHandler{
	log.Println(Yellow + "ProductHandler constructor is called" + Reset)
	hdlOnce.Do(func() {
		hdlInstance = &ProductHandler{svc: svc}
	})
	return hdlInstance
}

func (h *ProductHandler)CreateProduct(c *fiber.Ctx) error {
	input := new(CreateProductRequstDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	log.Println("New customer input: ", input)
	newProduct := models.Product{
		ID: input.ID,
		ProductName: input.ProductName,
		CategoryId: input.CategoryId,
		Uom: input.Uom,
		BuyPrice: input.BuyPrice,
		SellPriceLevel1: input.SellPriceLevel1,
		SellPriceLevel2: input.SellPriceLevel2,
		ReorderLvl: input.ReorderLvl,
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
	log.Println("newCustomer: ", newProduct)

	if _, err := h.svc.CreateSerive(&newProduct); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "category has been created successfully",
			"data" : newProduct,
		})

}

func(h *ProductHandler) GetAllProducts(c *fiber.Ctx) error {
	products, err := h.svc.GetAllSerive()
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(products)) + " records found",
			"data":    products,
		})

}

func(h *ProductHandler) GetProductById(c *fiber.Ctx) error {
	

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

func(h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {

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
		ID: foundProduct.ID,
		ProductName: foundProduct.ProductName,
		CategoryId: foundProduct.CategoryId,
		Uom: foundProduct.Uom,
		BuyPrice: foundProduct.BuyPrice,
		SellPriceLevel1: foundProduct.SellPriceLevel1,
		SellPriceLevel2: foundProduct.SellPriceLevel2,
		ReorderLvl: foundProduct.ReorderLvl,
		IsActive: foundProduct.IsActive,
	}
	log.Println("updateCustomer: ", &updateProduct)
	if err := c.BodyParser(&updateProduct); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}	

	result, err :=	h.svc.Update(&updateProduct)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})
}

func(h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {

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
			"status" : "FAIL",
			"message" : "Internal server error",
		})
	}
	return c.JSON(fiber.Map{
		"code":    200,
		"message": "Delete successfully",
	})
}