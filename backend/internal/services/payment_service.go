package services

import (
	"billing-system/internal/models"
	"billing-system/internal/repository"
	"context"
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// paymentService реализация PaymentService
type paymentService struct {
	paymentRepo    repository.PaymentRepository
	invoiceRepo    repository.InvoiceRepository
	brokerRepo     repository.BrokerRepository
	emailService   EmailService
	invoiceService InvoiceService
}

// NewPaymentService создает новый PaymentService
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	invoiceRepo repository.InvoiceRepository,
	brokerRepo repository.BrokerRepository,
	emailService EmailService,
) PaymentService {
	return &paymentService{
		paymentRepo:  paymentRepo,
		invoiceRepo:  invoiceRepo,
		brokerRepo:   brokerRepo,
		emailService: emailService,
	}
}

// SetInvoiceService устанавливает ссылку на InvoiceService для избежания циклических зависимостей
func (s *paymentService) SetInvoiceService(invoiceService InvoiceService) {
	s.invoiceService = invoiceService
}

// CreatePayment создает новый платеж
func (s *paymentService) CreatePayment(ctx context.Context, payment *models.Payment) error {
	// Валидация
	if err := s.validatePayment(ctx, payment); err != nil {
		return err
	}

	// Создаем платеж
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return err
	}

	// Обновляем статус счета
	if s.invoiceService != nil {
		if err := s.invoiceService.UpdateInvoiceStatus(ctx, payment.InvoiceID); err != nil {
			// Логируем ошибку, но не прерываем процесс
			// TODO: добавить логирование
		}
	}

	// Отправляем уведомление
	go s.sendPaymentNotification(payment)

	return nil
}

// GetPayment получает платеж по ID
func (s *paymentService) GetPayment(ctx context.Context, id primitive.ObjectID) (*models.Payment, error) {
	return s.paymentRepo.GetByID(ctx, id)
}

// GetAllPayments получает все платежи с фильтрацией и пагинацией
func (s *paymentService) GetAllPayments(ctx context.Context, filter *models.PaymentFilter, page, limit int) ([]*models.Payment, *models.Pagination, error) {
	offset := (page - 1) * limit

	payments, total, err := s.paymentRepo.GetAll(ctx, filter, limit, offset)
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

	return payments, pagination, nil
}

// UpdatePayment обновляет платеж
func (s *paymentService) UpdatePayment(ctx context.Context, id primitive.ObjectID, payment *models.Payment) error {
	// Получаем существующий платеж
	existingPayment, err := s.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Валидация
	if err := s.validatePayment(ctx, payment); err != nil {
		return err
	}

	// Обновляем платеж
	if err := s.paymentRepo.Update(ctx, id, payment); err != nil {
		return err
	}

	// Обновляем статус счета для старого и нового счета (если изменился)
	if s.invoiceService != nil {
		if err := s.invoiceService.UpdateInvoiceStatus(ctx, existingPayment.InvoiceID); err != nil {
			// Логируем ошибку
		}

		if existingPayment.InvoiceID != payment.InvoiceID {
			if err := s.invoiceService.UpdateInvoiceStatus(ctx, payment.InvoiceID); err != nil {
				// Логируем ошибку
			}
		}
	}

	return nil
}

// DeletePayment удаляет платеж
func (s *paymentService) DeletePayment(ctx context.Context, id primitive.ObjectID) error {
	// Получаем платеж для получения invoice_id
	payment, err := s.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Удаляем платеж
	if err := s.paymentRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Обновляем статус счета
	if s.invoiceService != nil {
		if err := s.invoiceService.UpdateInvoiceStatus(ctx, payment.InvoiceID); err != nil {
			// Логируем ошибку
		}
	}

	return nil
}

// GetPaymentsByInvoice получает все платежи по счету
func (s *paymentService) GetPaymentsByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Payment, error) {
	return s.paymentRepo.GetByInvoice(ctx, invoiceID)
}

// GetPaymentsByBroker получает платежи по брокеру
func (s *paymentService) GetPaymentsByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Payment, *models.Pagination, error) {
	offset := (page - 1) * limit

	payments, total, err := s.paymentRepo.GetByBroker(ctx, brokerID, limit, offset)
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

	return payments, pagination, nil
}

// validatePayment валидирует данные платежа
func (s *paymentService) validatePayment(ctx context.Context, payment *models.Payment) error {
	if payment.Amount <= 0 {
		return &ValidationError{Message: "Payment amount must be greater than zero"}
	}

	if payment.Currency == "" {
		return &ValidationError{Message: "Currency is required"}
	}

	if payment.Currency != models.CurrencyUSD &&
		payment.Currency != models.CurrencyEUR &&
		payment.Currency != models.CurrencyRUB {
		return &ValidationError{Message: "Unsupported currency"}
	}

	if payment.PaymentMethod == "" {
		return &ValidationError{Message: "Payment method is required"}
	}

	if payment.InvoiceID.IsZero() {
		return &ValidationError{Message: "Invoice ID is required"}
	}

	if payment.BrokerID.IsZero() {
		return &ValidationError{Message: "Broker ID is required"}
	}

	// Проверяем существование счета
	invoice, err := s.invoiceRepo.GetByID(ctx, payment.InvoiceID)
	if err != nil || invoice == nil {
		return &ValidationError{Message: "Invoice not found"}
	}

	// Проверяем соответствие валют
	if invoice.Currency != payment.Currency {
		return &ValidationError{Message: "Payment currency must match invoice currency"}
	}

	// Проверяем соответствие брокера
	if invoice.BrokerID != payment.BrokerID {
		return &ValidationError{Message: "Payment broker must match invoice broker"}
	}

	// Проверяем, что сумма платежа не превышает оставшуюся к доплате
	totalPaid, err := s.paymentRepo.GetTotalPaidAmount(ctx, payment.InvoiceID)
	if err != nil {
		return err
	}

	remainingAmount := invoice.Amount - totalPaid
	if payment.Amount > remainingAmount {
		return &ValidationError{Message: "Payment amount exceeds remaining amount due"}
	}

	return nil
}

// sendPaymentNotification отправляет уведомление о платеже
func (s *paymentService) sendPaymentNotification(payment *models.Payment) {
	if s.emailService == nil {
		return
	}

	ctx := context.Background()

	// Получаем брокера и счет
	broker, err := s.brokerRepo.GetByID(ctx, payment.BrokerID)
	if err != nil {
		return
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, payment.InvoiceID)
	if err != nil {
		return
	}

	// Отправляем уведомление
	s.emailService.SendPaymentReceived(ctx, broker, payment, invoice)
}
