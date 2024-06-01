package auth

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type AuthHandler struct {
	svc AuthServiceInterface
}

var (
	hdlInstance *AuthHandler
	hdlOnce     sync.Once
)

func NewAuthHandler(svc AuthServiceInterface) *AuthHandler {
	log.Println(util.Red + "AuthHandler constructor is called" + util.Reset)
	hdlOnce.Do(func() {
		hdlInstance = &AuthHandler{
			svc: svc,
		}
	})
	return hdlInstance
}

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
	at, rt, err := h.svc.Signin(input.Email, input.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail signin", "errors": err.Error()})
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
		}})

}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {

	tokenString := c.Cookies("refreshToken")
	rt, err := h.svc.Refresh(tokenString)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "fail",
			"errors": err.Error(),
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
		"message": "login success",
		"data": RefreshResponseDTO{
			RefreshToken: rt,
		}})
}

func(h *AuthHandler) Logout(c *fiber.Ctx) error {
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
