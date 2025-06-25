package unitofmeasurement

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

type UnitOfMeasurementHandler struct {
	svc UnitOfMeasurementServiceInterface
}

var (
	handlerInstance *UnitOfMeasurementHandler
	handlerOnce     sync.Once
)

func NewUnitOfMeasurementHandler(svc UnitOfMeasurementServiceInterface) *UnitOfMeasurementHandler {
	log.Println(util.Gray + "UnitOfMeasurementHandler constructor is called " + util.Reset)
	handlerOnce.Do(func() {
		handlerInstance = &UnitOfMeasurementHandler{svc: svc}
	})
	return handlerInstance
}

// CreateUnitOfMeasurement godoc
//
//	@Summary		Create new unit of measurement based on parameters
//	@Description	Create new unit of measurement based on parameters
//	@Tags			UnitOfMeasurements
//	@Accept			json
//	@Param			product	body		CreateUnitOfMeasurementDTO	true	"Product Data"
//	@Success		200		{object}	models.UnitOfMeasure
//	@Failure		400		{object}	httputil.HttpError400
//	@Failure		401		{object}	httputil.HttpError401
//	@Failure		500		{object}	httputil.HttpError500
//	@Failure		401		{object}	httputil.HttpError401
//	@Router			/api/unitofmeasurements [post]
//	@Security		Bearer
func (h *UnitOfMeasurementHandler) CreateUnitOfMeasurement(c *fiber.Ctx) error {

	input := new(models.UnitOfMeasure)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	newUnitOfMeasurement := models.UnitOfMeasure{
		ID:       input.ID,
		UnitName: input.UnitName,
	}

	errors := models.ValidateStruct(newUnitOfMeasurement)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	if _, err := h.svc.CreateUnitOfMeasurement(&newUnitOfMeasurement); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": "unit of measurement has been created successfully",
			"data":    newUnitOfMeasurement,
		})
}

// GetAllUnitOfMeasurement godoc

// @Summary		Fetch all unit of measurement
// @Description	Fetch all unit of measurement
// @Tags			UnitOfMeasurements
// @Accept			json
// @Produce		json
// @Success		200				{array}		models.UnitOfMeasure
// @Failure		400				{object}	httputil.HttpError400
// @Failure		401				{object}	httputil.HttpError401
// @Failure		500				{object}	httputil.HttpError500
// @Router			/api/unitofmeasurements	[get]
// @Security		Bearer
func (h *UnitOfMeasurementHandler) GetAllUnitOfMeasurement(c *fiber.Ctx) error {
	unitOfMeasurements, err := h.svc.GetAllUnitOfMeasurement()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusOK).JSON(
		&fiber.Map{
			"status":  "SUCCESS",
			"message": strconv.Itoa(len(unitOfMeasurements)) + " records found",
			"data":    unitOfMeasurements,
		})
}

// GetUnitOfMeasurementById godoc
//
//	@Summary		Fetch individual unit of measurement by Id
//	@Description	Fetch individual unit of measurement by Id
//	@Tags			UnitOfMeasurements
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"unit of measurement Id"
//	@Success		200					{object}	models.UnitOfMeasure
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/unitofmeasurements/{id}	[get]

// @Security       Bearer
func (h *UnitOfMeasurementHandler) GetUnitOfMeasurementById(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}

	unitOfMeasurement, err := h.svc.GetUnitOfMeasurementById(int(id))
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
		"data":    unitOfMeasurement,
	})
}

// UpdateUnitOfMeasurement godoc
//
//	@Summary		Update individual unit of measurement
//	@Description	Update individual unit of measurement
//	@Tags			UnitOfMeasurements
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string						true	"unit of measurement Id"
//	@Param			product				body		UpdateUnitOfMeasurementDTO	true	"Product Data"
//	@Success		200					{object}	models.UnitOfMeasure
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/unitofmeasurements/{id}	[put]
//	@Security		Bearer
func (h *UnitOfMeasurementHandler) UpdateUnitOfMeasurement(c *fiber.Ctx) error {

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}

	foundUnitOfMeasurement, err := h.svc.GetUnitOfMeasurementById(int(id))
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

	input := new(models.UnitOfMeasure)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}
	updateUnitOfMeasurement := models.UnitOfMeasure{
		ID:       foundUnitOfMeasurement.ID,
		UnitName: input.UnitName,
	}
	log.Println("updateUnitOfMeasurement: ", &updateUnitOfMeasurement)
	if err := c.BodyParser(&updateUnitOfMeasurement); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  400,
			"message": "Invalid JSON format",
		})
	}

	result, err := h.svc.UpdateUnitOfMeasurement(&updateUnitOfMeasurement)
	if err != nil {
		log.Fatal(err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Update Successfully",
		"data":    result,
	})
}

// DeleteUnitOfMeasurement godoc
//
//	@Summary		Delete individual unit of measurement
//	@Description	Delete individual unit of measurement
//	@Tags			UnitOfMeasurements
//	@Accept			json
//	@Produce		json
//	@Param			id					path		string	true	"unit of measurement Id"
//	@Success		200					{object}	models.UnitOfMeasure
//	@Failure		400					{object}	httputil.HttpError400
//	@Failure		401					{object}	httputil.HttpError401
//	@Failure		500					{object}	httputil.HttpError500
//	@Router			/api/unitofmeasurements/{id}	[delete]
//	@Security		Bearer
func (h *UnitOfMeasurementHandler) DeleteUnitOfMeasurement(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "invalid ID parameter",
			"detail": err.Error(),
		})
	}
	unitOfMeasurement, err := h.svc.GetUnitOfMeasurementById(int(id))
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
	err = h.svc.DeleteUnitOfMeasurement(int(unitOfMeasurement.ID))
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
