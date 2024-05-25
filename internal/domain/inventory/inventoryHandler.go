package inventory

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type InventoryHandler struct{
	svc InventoryServiceInterface
}

var (
	hdlInstance *InventoryHandler
	hdlOnce sync.Once
)

func NewInventoryHandler(svc InventoryServiceInterface) *InventoryHandler{
	log.Println(Cyan + "InventoryHandler constructor is called" + Reset)
	hdlOnce.Do(func() {
		hdlInstance = &InventoryHandler{svc: svc}
	})
	return hdlInstance
}

func (h *InventoryHandler)GetInventory(c *fiber.Ctx) error {
	inventories, err := h.svc.GetAllService()
	if err != nil {
		return  c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(inventories)) + " records found",
			"data":    inventories,
		})
}

func (h *InventoryHandler) IncreaseInventory(c *fiber.Ctx) error {

	input := new(IncreaseInventoryDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	newInventory := models.Inventory{
		InQty: input.InQty,
		OutQty: input.OutQty,
		ProductId: input.ProductId,
		Remark: input.Remark,
	}
	errors := models.ValidateStruct(newInventory)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "operation failed",
		})
	}

	msg, err := h.svc.IncreaseInventoryService(&newInventory)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status" : "FAIL",
			"message" : err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status" : "success",
		"message" : msg,
	})
}

func (h *InventoryHandler) DecreaseInventory(c *fiber.Ctx) error {

	input := new(IncreaseInventoryDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	newInventory := models.Inventory{
		InQty: input.InQty,
		OutQty: input.OutQty,
		ProductId: input.ProductId,
		Remark: input.Remark,
	}
	errors := models.ValidateStruct(newInventory)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "operation failed",
		})
	}

	msg, err := h.svc.DecreaseInventoryService(&newInventory)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status" : "FAIL",
			"message" : err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status" : "success",
		"message" : msg,
	})
}
