package middleware

import (
	"billing-system/internal/models"
	"billing-system/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler глобальный обработчик ошибок
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Получаем код ошибки
	code := fiber.StatusInternalServerError

	// Проверяем тип ошибки
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// Проверяем кастомные ошибки валидации
	if validationErr, ok := err.(*services.ValidationError); ok {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Success: false,
			Error:   validationErr.Message,
		})
	}

	// Обрабатываем различные типы ошибок
	switch code {
	case fiber.StatusNotFound:
		return c.Status(code).JSON(models.APIResponse{
			Success: false,
			Error:   "Ресурс не найден",
		})
	case fiber.StatusBadRequest:
		return c.Status(code).JSON(models.APIResponse{
			Success: false,
			Error:   "Некорректный запрос",
		})
	case fiber.StatusUnauthorized:
		return c.Status(code).JSON(models.APIResponse{
			Success: false,
			Error:   "Необходима авторизация",
		})
	case fiber.StatusForbidden:
		return c.Status(code).JSON(models.APIResponse{
			Success: false,
			Error:   "Доступ запрещен",
		})
	default:
		// Логируем внутренние ошибки сервера
		// TODO: Добавить логирование
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Success: false,
			Error:   "Внутренняя ошибка сервера",
		})
	}
}
