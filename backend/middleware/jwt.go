package middleware

import (
	"backend/database"
	"backend/model"
	"backend/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func isTokenBlacklisted(tokenString string) bool {
	var blacklistedToken model.BlacklistedToken
	result := database.DB.Where("token = ?", tokenString).First(&blacklistedToken)
	return result.Error == nil
}

func JWTProtected(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid authorization format. Use Bearer <token>",
			})
		}
		tokenString := parts[1]

		if isTokenBlacklisted(tokenString) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Token has been revoked. Please login again.",
			})
		}

		var claims jwt.MapClaims
		var tokenRole string
		var err error

		for _, role := range allowedRoles {
			_, claims, err = utils.VerifyToken(tokenString, role)
			if err == nil {
				tokenRole = role
				break
			}
		}

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid or expired token",
			})
		}

		c.Locals("token", tokenString)
		c.Locals("user", claims)
		c.Locals("role", tokenRole)

		if userID, ok := claims["user_id"].(float64); ok {
			c.Locals("user_id", uint(userID))
		}

		return c.Next()
	}
}

func BlacklistCheckMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]

			if isTokenBlacklisted(tokenString) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": "Token has been revoked. Please login again.",
				})
			}
		}

		return c.Next()
	}
}

func AdminProtected() fiber.Handler {
	return JWTProtected("admin")
}

func AgentProtected() fiber.Handler {
	return JWTProtected("agent")
}

func UserProtected() fiber.Handler {
	return JWTProtected("user")
}

func AdminOrAgentProtected() fiber.Handler {
	return JWTProtected("admin", "agent")
}

func AllRolesProtected() fiber.Handler {
	return JWTProtected("admin", "agent", "user")
}
