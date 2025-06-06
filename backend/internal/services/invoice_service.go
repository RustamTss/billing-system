package services

import (
	"billing-system/internal/models"
	"billing-system/internal/repository"
	"context"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// invoiceService реализация InvoiceService
type invoiceService struct {
	invoiceRepo  repository.InvoiceRepository
	paymentRepo  repository.PaymentRepository
	brokerRepo   repository.BrokerRepository
	emailService EmailService
}

// NewInvoiceService создает новый InvoiceService
func NewInvoiceService(
	invoiceRepo repository.InvoiceRepository,
	paymentRepo repository.PaymentRepository,
	brokerRepo repository.BrokerRepository,
	emailService EmailService,
) InvoiceService {
	return &invoiceService{
		invoiceRepo:  invoiceRepo,
		paymentRepo:  paymentRepo,
		brokerRepo:   brokerRepo,
		emailService: emailService,
	}
}

// CreateInvoice создает новый счет
func (s *invoiceService) CreateInvoice(ctx context.Context, invoice *models.Invoice) error {
	// Валидация
	if err := s.validateInvoice(invoice); err != nil {
		return err
	}

	// Устанавливаем статус по умолчанию
	if invoice.Status == "" {
		invoice.Status = models.InvoiceStatusPending
	}

	// Создаем счет
	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return err
	}

	// Отправляем уведомление брокеру
	broker, err := s.brokerRepo.GetByID(ctx, invoice.BrokerID)
	if err == nil && s.emailService != nil {
		go s.emailService.SendInvoiceCreated(context.Background(), broker, invoice)
	}

	return nil
}

// GetInvoice получает счет по ID
func (s *invoiceService) GetInvoice(ctx context.Context, id primitive.ObjectID) (*models.Invoice, error) {
	return s.invoiceRepo.GetByID(ctx, id)
}

// GetAllInvoices получает все счета с фильтрацией и пагинацией
func (s *invoiceService) GetAllInvoices(ctx context.Context, filter *models.InvoiceFilter, page, limit int) ([]*models.Invoice, *models.Pagination, error) {
	offset := (page - 1) * limit

	invoices, total, err := s.invoiceRepo.GetAll(ctx, filter, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		HasNext:    int64(page*limit) < total,
		HasPrev:    page > 1,
	}

	return invoices, pagination, nil
}

// UpdateInvoice обновляет счет
func (s *invoiceService) UpdateInvoice(ctx context.Context, id primitive.ObjectID, invoice *models.Invoice) error {
	// Валидация
	if err := s.validateInvoice(invoice); err != nil {
		return err
	}

	return s.invoiceRepo.Update(ctx, id, invoice)
}

// DeleteInvoice удаляет счет
func (s *invoiceService) DeleteInvoice(ctx context.Context, id primitive.ObjectID) error {
	// Проверяем, есть ли связанные платежи
	payments, err := s.paymentRepo.GetByInvoice(ctx, id)
	if err != nil {
		return err
	}

	if len(payments) > 0 {
		return &ValidationError{Message: "Cannot delete invoice with existing payments"}
	}

	return s.invoiceRepo.Delete(ctx, id)
}

// GetInvoicesByStatus получает счета по статусу
func (s *invoiceService) GetInvoicesByStatus(ctx context.Context, status string, page, limit int) ([]*models.Invoice, *models.Pagination, error) {
	offset := (page - 1) * limit

	invoices, total, err := s.invoiceRepo.GetByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		HasNext:    int64(page*limit) < total,
		HasPrev:    page > 1,
	}

	return invoices, pagination, nil
}

// GetInvoicesByBroker получает счета по брокеру
func (s *invoiceService) GetInvoicesByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Invoice, *models.Pagination, error) {
	offset := (page - 1) * limit

	invoices, total, err := s.invoiceRepo.GetByBroker(ctx, brokerID, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		HasNext:    int64(page*limit) < total,
		HasPrev:    page > 1,
	}

	return invoices, pagination, nil
}

// GetOverdueInvoices получает просроченные счета
func (s *invoiceService) GetOverdueInvoices(ctx context.Context, page, limit int) ([]*models.Invoice, *models.Pagination, error) {
	offset := (page - 1) * limit

	invoices, total, err := s.invoiceRepo.GetOverdue(ctx, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		HasNext:    int64(page*limit) < total,
		HasPrev:    page > 1,
	}

	return invoices, pagination, nil
}

// UpdateInvoiceStatus обновляет статус счета на основе платежей
func (s *invoiceService) UpdateInvoiceStatus(ctx context.Context, invoiceID primitive.ObjectID) error {
	// Получаем счет
	invoice, err := s.invoiceRepo.GetByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Получаем общую сумму платежей
	totalPaid, err := s.paymentRepo.GetTotalPaidAmount(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Определяем новый статус
	var newStatus string
	if totalPaid >= invoice.Amount {
		newStatus = models.InvoiceStatusPaid
	} else if totalPaid > 0 {
		newStatus = models.InvoiceStatusPartial
	} else if time.Now().After(invoice.DueDate) {
		newStatus = models.InvoiceStatusOverdue
	} else {
		newStatus = models.InvoiceStatusPending
	}

	// Обновляем статус и оплаченную сумму
	return s.invoiceRepo.UpdateStatus(ctx, invoiceID, newStatus, totalPaid)
}

// SendOverdueNotifications отправляет уведомления о просроченных счетах
func (s *invoiceService) SendOverdueNotifications(ctx context.Context) error {
	// Получаем просроченные счета
	overdueInvoices, _, err := s.invoiceRepo.GetOverdue(ctx, 1000, 0) // Максимум 1000 за раз
	if err != nil {
		return err
	}

	// Группируем по брокерам
	brokerInvoices := make(map[primitive.ObjectID][]*models.Invoice)
	for _, invoice := range overdueInvoices {
		brokerInvoices[invoice.BrokerID] = append(brokerInvoices[invoice.BrokerID], invoice)
	}

	// Отправляем уведомления каждому брокеру
	for brokerID, invoices := range brokerInvoices {
		broker, err := s.brokerRepo.GetByID(ctx, brokerID)
		if err != nil {
			continue // Пропускаем если брокер не найден
		}

		if s.emailService != nil {
			go s.emailService.SendOverdueNotification(context.Background(), broker, invoices)
		}
	}

	return nil
}

// validateInvoice валидирует данные счета
func (s *invoiceService) validateInvoice(invoice *models.Invoice) error {
	if invoice.Amount <= 0 {
		return &ValidationError{Message: "Invoice amount must be greater than zero"}
	}

	if invoice.Currency == "" {
		return &ValidationError{Message: "Currency is required"}
	}

	if invoice.Currency != models.CurrencyUSD &&
		invoice.Currency != models.CurrencyEUR &&
		invoice.Currency != models.CurrencyRUB {
		return &ValidationError{Message: "Unsupported currency"}
	}

	if invoice.DueDate.IsZero() {
		return &ValidationError{Message: "Due date is required"}
	}

	if invoice.BrokerID.IsZero() {
		return &ValidationError{Message: "Broker ID is required"}
	}

	return nil
}

// ValidationError ошибка валидации
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
