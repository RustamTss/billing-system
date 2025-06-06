package handlers

import (
	"billing-system/internal/models"
	"billing-system/internal/services"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Handlers основная структура handlers
type Handlers struct {
	brokerService    services.BrokerService
	invoiceService   services.InvoiceService
	paymentService   services.PaymentService
	loadService      services.LoadService
	dashboardService services.DashboardService
}

// NewHandlers создает новый экземпляр handlers
func NewHandlers(
	brokerService services.BrokerService,
	invoiceService services.InvoiceService,
	paymentService services.PaymentService,
	loadService services.LoadService,
	dashboardService services.DashboardService,
) *Handlers {
	return &Handlers{
		brokerService:    brokerService,
		invoiceService:   invoiceService,
		paymentService:   paymentService,
		loadService:      loadService,
		dashboardService: dashboardService,
	}
}

// Dashboard handlers

// GetDashboardMetrics получает метрики дашборда
func (h *Handlers) GetDashboardMetrics(c *fiber.Ctx) error {
	metrics, err := h.dashboardService.GetDashboardMetrics(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    metrics,
	})
}

// Broker handlers

// GetBrokers получает список брокеров
func (h *Handlers) GetBrokers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	brokers, pagination, err := h.brokerService.GetAllBrokers(c.Context(), page, limit)
	if err != nil {
		return err
	}

	return c.JSON(models.PaginatedResponse{
		Success:    true,
		Data:       brokers,
		Pagination: *pagination,
	})
}

// CreateBroker создает нового брокера
func (h *Handlers) CreateBroker(c *fiber.Ctx) error {
	var broker models.Broker
	if err := c.BodyParser(&broker); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.brokerService.CreateBroker(c.Context(), &broker); err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create broker",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broker created successfully",
		"data":    broker,
	})
}

// GetBroker получает брокера по ID
func (h *Handlers) GetBroker(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid broker ID",
		})
	}

	broker, err := h.brokerService.GetBroker(c.Context(), objectID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Broker not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    broker,
	})
}

// UpdateBroker обновляет брокера
func (h *Handlers) UpdateBroker(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid broker ID",
		})
	}

	var broker models.Broker
	if err := c.BodyParser(&broker); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.brokerService.UpdateBroker(c.Context(), objectID, &broker); err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update broker",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broker updated successfully",
	})
}

// DeleteBroker удаляет брокера
func (h *Handlers) DeleteBroker(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid broker ID",
		})
	}

	if err := h.brokerService.DeleteBroker(c.Context(), objectID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete broker",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broker deleted successfully",
	})
}

// SearchBrokers поиск брокеров
func (h *Handlers) GetAllBrokers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	brokers, pagination, err := h.brokerService.GetAllBrokers(c.Context(), page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch brokers",
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"data":       brokers,
		"pagination": pagination,
	})
}

func (h *Handlers) SearchBrokers(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Search query is required",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	brokers, pagination, err := h.brokerService.SearchBrokers(c.Context(), query, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to search brokers",
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"data":       brokers,
		"pagination": pagination,
	})
}

// GetBrokerStats получает статистику брокера
func (h *Handlers) GetBrokerStats(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Некорректный ID")
	}

	stats, err := h.brokerService.GetBrokerStats(c.Context(), objectID)
	if err != nil {
		return err
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// GetBrokerInvoices получает счета брокера
func (h *Handlers) GetBrokerInvoices(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Некорректный ID")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	invoices, pagination, err := h.invoiceService.GetInvoicesByBroker(c.Context(), objectID, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(models.PaginatedResponse{
		Success:    true,
		Data:       invoices,
		Pagination: *pagination,
	})
}

// GetBrokerPayments получает платежи брокера
func (h *Handlers) GetBrokerPayments(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Некорректный ID")
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	payments, pagination, err := h.paymentService.GetPaymentsByBroker(c.Context(), objectID, page, limit)
	if err != nil {
		return err
	}

	return c.JSON(models.PaginatedResponse{
		Success:    true,
		Data:       payments,
		Pagination: *pagination,
	})
}

// GetBrokerUnbilledLoads получает неоплаченные грузы брокера
func (h *Handlers) GetBrokerUnbilledLoads(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid broker ID",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)

	loads, pagination, err := h.loadService.GetUnbilledLoadsByBroker(c.Context(), objectID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch unbilled loads",
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"data":       loads,
		"pagination": pagination,
	})
}

// Placeholder handlers для остальных endpoints
// TODO: Реализовать полные handlers для Invoice, Payment, Load

// GetInvoices получает список счетов
func (h *Handlers) GetInvoices(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Создаем фильтр
	filter := &models.InvoiceFilter{}
	if status := c.Query("status"); status != "" {
		filter.Status = strings.Split(status, ",")
	}

	invoices, pagination, err := h.invoiceService.GetAllInvoices(c.Context(), filter, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch invoices",
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"data":       invoices,
		"pagination": pagination,
	})
}

// CreateInvoice создает новый счет
func (h *Handlers) CreateInvoice(c *fiber.Ctx) error {
	var invoice models.Invoice
	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	err := h.invoiceService.CreateInvoice(c.Context(), &invoice)
	if err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create invoice",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Invoice created successfully",
		"data":    invoice,
	})
}

// GetInvoice получает счет по ID
func (h *Handlers) GetInvoice(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid invoice ID",
		})
	}
	invoice, err := h.invoiceService.GetInvoice(c.Context(), objectID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Invoice not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    invoice,
	})
}

// UpdateInvoice обновляет счет
func (h *Handlers) UpdateInvoice(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid invoice ID",
		})
	}
	var invoice models.Invoice
	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.invoiceService.UpdateInvoice(c.Context(), objectID, &invoice); err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update invoice",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Invoice updated successfully",
	})
}

// DeleteInvoice удаляет счет
func (h *Handlers) DeleteInvoice(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid invoice ID",
		})
	}

	if err := h.invoiceService.DeleteInvoice(c.Context(), objectID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete invoice",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Invoice deleted successfully",
	})
}

// GetOverdueInvoices получает просроченные счета
func (h *Handlers) GetOverdueInvoices(c *fiber.Ctx) error {
	// TODO: Реализовать
	return c.JSON(models.APIResponse{
		Success: true,
		Data:    []interface{}{},
	})
}

// GetInvoicesByStatus получает счета по статусу
func (h *Handlers) GetInvoicesByStatus(c *fiber.Ctx) error {
	// TODO: Реализовать
	return c.JSON(models.APIResponse{
		Success: true,
		Data:    []interface{}{},
	})
}

// GetInvoicePayments получает платежи по счету
func (h *Handlers) GetInvoicePayments(c *fiber.Ctx) error {
	// TODO: Реализовать
	return c.JSON(models.APIResponse{
		Success: true,
		Data:    []interface{}{},
	})
}

// GetPayments получает список платежей
func (h *Handlers) GetPayments(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Создаем фильтр
	filter := &models.PaymentFilter{}
	if method := c.Query("method"); method != "" {
		filter.PaymentMethod = method
	}

	payments, pagination, err := h.paymentService.GetAllPayments(c.Context(), filter, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch payments",
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"data":       payments,
		"pagination": pagination,
	})
}

// CreatePayment создает новый платеж
func (h *Handlers) CreatePayment(c *fiber.Ctx) error {
	var payment models.Payment
	if err := c.BodyParser(&payment); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	err := h.paymentService.CreatePayment(c.Context(), &payment)
	if err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create payment",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment created successfully",
		"data":    payment,
	})
}

// GetPayment получает платеж по ID
func (h *Handlers) GetPayment(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid payment ID",
		})
	}

	payment, err := h.paymentService.GetPayment(c.Context(), objectID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Payment not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    payment,
	})
}

// UpdatePayment обновляет платеж
func (h *Handlers) UpdatePayment(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid payment ID",
		})
	}

	var payment models.Payment
	if err := c.BodyParser(&payment); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.paymentService.UpdatePayment(c.Context(), objectID, &payment); err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update payment",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment updated successfully",
	})
}

// DeletePayment удаляет платеж
func (h *Handlers) DeletePayment(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid payment ID",
		})
	}

	if err := h.paymentService.DeletePayment(c.Context(), objectID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete payment",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment deleted successfully",
	})
}

// GetLoads получает список грузов
func (h *Handlers) GetLoads(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Создаем фильтр
	filter := &models.LoadFilter{}
	if brokerID := c.Query("broker_id"); brokerID != "" {
		if objectID, err := primitive.ObjectIDFromHex(brokerID); err == nil {
			filter.BrokerID = objectID
		}
	}
	if status := c.Query("status"); status != "" {
		filter.Status = strings.Split(status, ",")
	}

	loads, pagination, err := h.loadService.GetAllLoads(c.Context(), filter, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch loads",
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"data":       loads,
		"pagination": pagination,
	})
}

// CreateLoad создает новый груз
func (h *Handlers) CreateLoad(c *fiber.Ctx) error {
	var load models.Load
	if err := c.BodyParser(&load); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	err := h.loadService.CreateLoad(c.Context(), &load)
	if err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create load",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Load created successfully",
		"data":    load,
	})
}

// GetLoad получает груз по ID
func (h *Handlers) GetLoad(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid load ID",
		})
	}

	load, err := h.loadService.GetLoad(c.Context(), objectID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "Load not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    load,
	})
}

// UpdateLoad обновляет груз
func (h *Handlers) UpdateLoad(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid load ID",
		})
	}

	var load models.Load
	if err := c.BodyParser(&load); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.loadService.UpdateLoad(c.Context(), objectID, &load); err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update load",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Load updated successfully",
	})
}

// DeleteLoad удаляет груз
func (h *Handlers) DeleteLoad(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid load ID",
		})
	}

	if err := h.loadService.DeleteLoad(c.Context(), objectID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete load",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Load deleted successfully",
	})
}

// UpdateLoadStatus обновляет статус груза
func (h *Handlers) UpdateLoadStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid load ID",
		})
	}

	var statusRequest struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&statusRequest); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if err := h.loadService.UpdateLoadStatus(c.Context(), objectID, statusRequest.Status); err != nil {
		if validationErr, ok := err.(*services.ValidationError); ok {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"error":   validationErr.Message,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update load status",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Load status updated successfully",
	})
}

// ExportInvoices экспорт счетов
func (h *Handlers) ExportInvoices(c *fiber.Ctx) error {
	// TODO: Реализовать
	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Функция в разработке",
	})
}

// ExportPayments экспорт платежей
func (h *Handlers) ExportPayments(c *fiber.Ctx) error {
	// TODO: Реализовать
	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Функция в разработке",
	})
}

// ExportBrokers экспорт брокеров
func (h *Handlers) ExportBrokers(c *fiber.Ctx) error {
	// TODO: Реализовать
	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Функция в разработке",
	})
}

// SendOverdueNotifications отправляет уведомления о просроченных счетах
func (h *Handlers) SendOverdueNotifications(c *fiber.Ctx) error {
	if err := h.invoiceService.SendOverdueNotifications(c.Context()); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to send notifications",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Notifications sent successfully",
	})
}
