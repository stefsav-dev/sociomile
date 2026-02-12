package controller

import (
	"backend/database"
	"backend/model"

	"github.com/gofiber/fiber/v3"
)

func GetProfile(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Unauthorized - User ID not found",
		})
	}

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "User not found",
		})
	}

	user.PasswordHash = ""

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

func GetAllUsers(c fiber.Ctx) error {
	role, ok := c.Locals("role").(string)
	if !ok || role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied. Admin only.",
		})
	}

	var users []model.User

	if err := database.DB.Select("id, email, full_name, role, created_at, updated_at").Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch users",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    users,
		"total":   len(users),
	})
}
