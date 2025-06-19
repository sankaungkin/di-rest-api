package inventory

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type InventoryHandler struct {
	svc InventoryServiceInterface
}

var (
	hdlInstance *InventoryHandler
	hdlOnce     sync.Once
)

func NewInventoryHandler(svc InventoryServiceInterface) *InventoryHandler {
	log.Println(util.Cyan + "InventoryHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &InventoryHandler{svc: svc}
	})
	return hdlInstance
}

// GetAllInventories godoc
//
//	@Summary		Fetch all inventory records
//	@Description	Fetch all inventory records
//	@Tags			Inventories
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Inventory
//	@Failure		400				{object}	httputil.HttpError400
//	@Failure		401				{object}	httputil.HttpError401
//	@Failure		500				{object}	httputil.HttpError500
//	@Router			/api/inventories	[get]
//	@Security		Bearer
func (h *InventoryHandler) GetAllInventories(c *fiber.Ctx) error {
	// inventories, err := h.svc.GetAllService()
	inventories, err := h.svc.GetInvData()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(inventories)) + " records found",
			"data":    inventories,
		})
}

// IncreaseInventory 	godoc
//
//	@Summary		Create increase inventory record based on parameters
//	@Description	Create increase inventory record based on parameters
//	@Tags			Inventories
//	@Accept			json
//	@Param			inventory	body		IncreaseInventoryDTO	true	"Inventory Data"
//	@Success		200		{object}	models.Inventory
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/inventories/increase [post]
//	@Security		Bearer
func (h *InventoryHandler) IncreaseInventory(c *fiber.Ctx) error {

	input := new(IncreaseInventoryDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	newInventory := models.Inventory{
		InQty:     input.InQty,
		OutQty:    input.OutQty,
		ProductId: input.ProductId,
		Remark:    input.Remark,
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
			"status":  "FAIL",
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": msg,
	})
}

// DecreaseInventory 	godoc
//
//	@Summary		Create decrease inventory record based on parameters
//	@Description	Create decrease inventory record based on parameters
//	@Tags			Inventories
//	@Accept			json
//	@Param			inventory	body		IncreaseInventoryDTO	true	"Inventory Data"
//	@Success		200		{object}	models.Inventory
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/inventories/decrease [post]
//	@Security		Bearer
func (h *InventoryHandler) DecreaseInventory(c *fiber.Ctx) error {

	input := new(IncreaseInventoryDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	newInventory := models.Inventory{
		InQty:     input.InQty,
		OutQty:    input.OutQty,
		ProductId: input.ProductId,
		Remark:    input.Remark,
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
			"status":  "FAIL",
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": msg,
	})
}
