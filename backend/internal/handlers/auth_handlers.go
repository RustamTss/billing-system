package handlers

import (
	"billing-system/internal/models"
	"billing-system/internal/services"

	"github.com/gofiber/fiber/v2"
)

type AuthHandlers struct {
	authService services.AuthService
}

func NewAuthHandlers(authService services.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// Login аутентификация пользователя
func (h *AuthHandlers) Login(c *fiber.Ctx) error {
	var loginReq models.LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Неверный формат запроса",
		})
	}

	authResponse, err := h.authService.Login(c.Context(), &loginReq)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    authResponse,
	})
}

// Register регистрация нового пользователя
func (h *AuthHandlers) Register(c *fiber.Ctx) error {
	var registerReq models.RegisterRequest
	if err := c.BodyParser(&registerReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Неверный формат запроса",
		})
	}

	authResponse, err := h.authService.Register(c.Context(), &registerReq)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    authResponse,
	})
}

// GetProfile получение профиля текущего пользователя
func (h *AuthHandlers) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Пользователь не найден",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// ValidateToken проверка валидности токена
func (h *AuthHandlers) ValidateToken(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	username := c.Locals("username").(string)
	role := c.Locals("role").(string)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":  userID,
			"username": username,
			"role":     role,
		},
	})
}
