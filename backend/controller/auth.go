package controller

import (
	"backend/database"
	"backend/model"
	"backend/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type AuthResponse struct {
	Token     string     `json:"token"`
	User      model.User `json:"user"`
	ExpiresIn int64      `json:"expires_in"`
}

func Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	if req.Role == "" {
		req.Role = "user"
	}

	validRoles := map[string]bool{"admin": true, "agent": true, "user": true}
	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid role. Must be admin, agent, or user",
		})
	}

	user := model.User{
		Email:        req.Email,
		PasswordHash: utils.GeneratePassword(req.Password),
		FullName:     req.FullName,
		Role:         model.Role(req.Role),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create user: " + err.Error(),
		})
	}

	user.PasswordHash = ""

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data":    user,
	})
}

func Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	isLimited, attempts := utils.IsRateLimited(req.Email, 5, 15*time.Minute)
	if isLimited {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error":   true,
			"message": "Too many failed attempts. Please try again later.",
		})
	}
	_ = attempts

	var user model.User
	err := database.DB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.IncrementFailedLogin(req.Email)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid email or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Database error",
		})
	}

	if !utils.ComparePassword(user.PasswordHash, req.Password) {
		utils.IncrementFailedLogin(req.Email)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid email or password",
		})
	}
	utils.ResetFailedLogin(req.Email)

	tokenDetails, err := utils.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to generate token",
		})
	}

	refreshToken := utils.GenerateRefreshToken()
	utils.StoreRefreshToken(user.ID, refreshToken, 7*24*time.Hour)

	user.PasswordHash = ""

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"access_token":  tokenDetails.Token,
			"refresh_token": refreshToken,
			"expires_in":    tokenDetails.ExpiresAt.Unix(),
			"user":          user,
		},
	})
}

func CreateUserByAdmin(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	validRoles := map[string]bool{"admin": true, "agent": true, "user": true}
	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid role. Must be admin, agent, or user",
		})
	}

	user := model.User{
		Email:        req.Email,
		PasswordHash: utils.GeneratePassword(req.Password),
		FullName:     req.FullName,
		Role:         model.Role(req.Role),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create user: " + err.Error(),
		})
	}

	user.PasswordHash = ""

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data":    user,
	})
}

func Logout(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "No token provided",
		})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid token format",
		})
	}
	tokenString := parts[1]

	utils.BlacklistToken(tokenString, 24*time.Hour)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Successfully logged out",
	})
}

func RefreshToken(c fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Refresh token endpoint",
	})
}
