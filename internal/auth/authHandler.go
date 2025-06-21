package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/models"
	mylog "github.com/sirupsen/logrus"
)

type AuthHandler struct {
	svc AuthServiceInterface
}

var (
	hdlInstance *AuthHandler
	hdlOnce     sync.Once
)

func init() {
	mylog.SetReportCaller(true)
	Formatter := new(mylog.JSONFormatter)
	Formatter.TimestampFormat = "15:04:05 01/02/06 "
	mylog.SetFormatter(Formatter)
}

func NewAuthHandler(svc AuthServiceInterface) *AuthHandler {

	mylog.Info("AuthHandler constructor is called")
	// log.Println(util.Red + "AuthHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &AuthHandler{
			svc: svc,
		}
	})
	return hdlInstance
}

// Register	godoc
// @Summary		Create new user based on parameters
// @Description	Register new user based on parameters
//
// @Tags			Auth
// @Accept			json
// @Param			info	body		SignUpDTO	true	"Signup Data"
// @Success		200		{object}	SignUpResponseDTO
// @Failure		400		{object}	httputil.HttpError400
// @Failure		401		{object}	httputil.HttpError401
// @Failure		500		{object}	httputil.HttpError500
// @Failure		401		{object}	httputil.HttpError401
// @Router			/api/auth/register [post]
func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid request body")
	}
	email := user.Email
	if _, err := h.svc.FindUserByEmail(email); err != nil {

		createdUser, err := h.svc.Signup(&user)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(err)
		}

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"status": "SUCCESS",
			"data":   createdUser,
		})
	}
	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"status":  "FAIL",
		"message": "user name has been already taken",
	})
}

// Login	godoc
//
//	@Summary	Login to the api with email and password
//	@Tags		Auth
//	@Accept		json
//	@Param		info	body		SignInRequestDTO	true	"Login Data"
//	@Success	200		{object}	SignInResponseDTO
//	@Failure	400		{object}	httputil.HttpError400
//	@Failure	401		{object}	httputil.HttpError401
//	@Failure	500		{object}	httputil.HttpError500
//	@Failure	401		{object}	httputil.HttpError401
//	@Router		/api/auth/login [post]
func (h *AuthHandler) SignIn(c *fiber.Ctx) error {
	input := new(SignInRequestDTO)

	if err := c.BodyParser(input); err != nil {
		return c.JSON(fiber.Map{
			"code":    400,
			"message": "Invalid JSON",
		})
	}
	errors := models.ValidateStruct(input)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail validation", "errors": errors})
	}
	at, rt, userName, role, err := h.svc.Signin(input.Email, input.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "invalid credentials", "errors": err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    at,
		Path:     "/",
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    rt,
		Path:     "/",
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "login success",
		"data": SignInResponseDTO{
			AccessToken:  at,
			RefreshToken: rt,
			UserName:     userName,
			Role:         role,
		}})

}

// / RefreshToken godoc
//
//	@Summary		Refresh access token
//	@Description	Get new access token using refresh token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body	RefreshRequestDTO	true	"Refresh token request"
//	@Success		200		{object}	RefreshResponseDTO
//	@Failure		400		{object}	httputil.HttpError400
//	@Router			/api/auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var body RefreshResponseDTO

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  "Invalid body",
		})
	}
	at, rt, err := h.svc.Refresh(body.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"error":  err.Error(),
		})
	}
	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    "",
		Path:     "/",
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    rt,
		Path:     "/",
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Token refreshed successfully",
		"data": RefreshResponseDTO{
			AccessToken:  at,
			RefreshToken: rt,
		}})
}

// Logout	godoc
//
//	@Summary		Logout user
//	@Description	Logout user
//
//	@Tags			Auth
//	@Success		200
//	@Failure		400	{object}	httputil.HttpError400
//	@Failure		401	{object}	httputil.HttpError401
//	@Failure		500	{object}	httputil.HttpError500
//	@Failure		401	{object}	httputil.HttpError401
//	@Router			/api/auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	expired := time.Now().Add(-time.Hour * 24)
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		Expires:  expired,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "accessToken",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		Expires:  expired,
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}
