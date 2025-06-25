package unitconversion

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

type UnitConversionHandler struct {
	svc UnitConversionServiceInterface
}

var (
	handlerInstance *UnitConversionHandler
	handlerOnce     sync.Once
)

func NewUnitConversionHandler(svc UnitConversionServiceInterface) *UnitConversionHandler {
	log.Println(util.Gray + "UnitConversionHandler constructor is called " + util.Reset)
	handlerOnce.Do(func() {
		handlerInstance = &UnitConversionHandler{svc: svc}
	})
	return handlerInstance
}

// CreateUnitConversion godoc
//
//	@Summary		Create new unit conversion based on parameters
//	@Description	Create new unit conversion based on parameters
//	@Tags			UnitConversions
//	@Accept			json
//	@Param			product	body		CreateUnitConversionDTO	true	"Product Data"
//	@Success		200		{object}	models.UnitConversion
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/unitconversions [post]
//	@Security		Bearer
func (h *UnitConversionHandler) CreateUnitConversion(c *fiber.Ctx) error {

	input := new(CreateUnitConversionDTO)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	newUnitConversion := models.UnitConversion{
		Factor:       input.Factor,
		Description:  input.Description,
		ProductId:    input.ProductId,
		BaseUnitId:   input.BaseUnitId,
		DeriveUnitId: input.DeriveUnitId,
	}

	errors := models.ValidateStruct(newUnitConversion)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	if _, err := h.svc.CreateUnitConversion(&newUnitConversion); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "unit conversion has been created successfully",
			"data":    newUnitConversion,
		})
}

// GetAllUnitConversions godoc

// @Summary		Fetch all unit conversions
// @Description	Fetch all unit conversions
// @Tags			UnitConversions
// @Accept			json
// @Produce		json
// @Success		200				{array}		models.UnitConversion
// @Failure		400				{object}	httputil.HttpError400
// @Failure		401				{object}	httputil.HttpError401
// @Failure		500				{object}	httputil.HttpError500
// @Router			/api/unitconversions	[get]
// @Security		Bearer
func (h *UnitConversionHandler) GetAllUnitConversions(c *fiber.Ctx) error {
	unitConversions, err := h.svc.GetAllUnitConversions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(unitConversions)) + " records found",
			"data":    unitConversions,
		})
}

// GetUnitConversionById godoc
//
//	@Summary		Fetch individual unit conversion by Id
//	@Description	Fetch individual unit conversion by Id
//	@Tags			UnitConversions
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"unit conversion Id"
//	@Success		200					{object}	models.UnitConversion
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/unitconversions/{id}	[get]
//	@Security		Bearer
func (h *UnitConversionHandler) GetUnitConversionById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}

	unitConversion, err := h.svc.GetUnitConversionById(int(id))
	log.Println("id : ", id)
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
		"data":    unitConversion,
	})
}

// UpdateUnitConversion godoc
//
//	@Summary		Update individual unit conversion
//	@Description	Update individual unit conversion
//	@Tags			UnitConversions
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"unit conversion Id"
//	@Param			product				body		UpdateUnitConversionRequestDTO	true	"Product Data"
//	@Success		200					{object}	models.UnitConversion
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/unitconversions/{id}	[put]
//	@Security		Bearer
func (h *UnitConversionHandler) UpdateUnitConversion(c *fiber.Ctx) error {

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}

	foundUnitConversion, err := h.svc.GetUnitConversionById(int(id))
	log.Println("id : ", id)
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

	input := new(UpdateUnitConversionRequestDTO)
	log.Println("input: ", input)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	updateUnitConversion := models.UnitConversion{
		ID:           foundUnitConversion.ID,
		ProductId:    input.ProductId,
		BaseUnitId:   input.BaseUnitId,
		DeriveUnitId: input.DeriveUnitId,
		Factor:       input.Factor,
		Description:  input.Description,
	}
	log.Println("updateUnitConversion: ", &updateUnitConversion)
	if err := c.BodyParser(&updateUnitConversion); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	result, err := h.svc.UpdateUnitConversion(&updateUnitConversion)
	if err != nil {
		log.Fatal(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})

}

// DeleteUnitConversion godoc
//
//	@Summary		Delete individual unit conversion
//	@Description	Delete individual unit conversion
//	@Tags			UnitConversions
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"unit conversion Id"
//	@Success		200					{object}	models.UnitConversion
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/unitconversions/{id}	[delete]
//	@Security		Bearer
func (h *UnitConversionHandler) DeleteUnitConversion(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}
	conversion, err := h.svc.GetUnitConversionById(int(id))
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
	err = h.svc.DeleteUnitConversion(int(conversion.ID))
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
