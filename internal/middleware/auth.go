package middleware

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
)


func Protected() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			Key: []byte(util.SecreteKey),
		},
		ErrorHandler: jwtError,
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{
				"status":  "error",
				"message": "Missing or malformed JWT",
				"data":    nil,
			})
	}
	return c.Status(fiber.StatusUnauthorized).
		JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
			"data":    nil,
		})
}

func Authorize(c *fiber.Ctx) error {

	tokenString := c.Cookies("refresh_token")
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(util.SecreteKey), nil
	})

	if err != nil {
		c.JSON(err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		role := claims["role"].(string)
		if role == "ADMIN" {
			c.Next()
		}

	}
	return nil

}