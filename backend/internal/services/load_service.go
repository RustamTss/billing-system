package services

import (
	"billing-system/internal/models"
	"billing-system/internal/repository"
	"context"
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// loadService реализация LoadService
type loadService struct {
	loadRepo    repository.LoadRepository
	brokerRepo  repository.BrokerRepository
	invoiceRepo repository.InvoiceRepository
}

// NewLoadService создает новый LoadService
func NewLoadService(
	loadRepo repository.LoadRepository,
	brokerRepo repository.BrokerRepository,
	invoiceRepo repository.InvoiceRepository,
) LoadService {
	return &loadService{
		loadRepo:    loadRepo,
		brokerRepo:  brokerRepo,
		invoiceRepo: invoiceRepo,
	}
}

// CreateLoad создает новый груз
func (s *loadService) CreateLoad(ctx context.Context, load *models.Load) error {
	// Валидация
	if err := s.validateLoad(ctx, load); err != nil {
		return err
	}

	return s.loadRepo.Create(ctx, load)
}

// GetLoad получает груз по ID
func (s *loadService) GetLoad(ctx context.Context, id primitive.ObjectID) (*models.Load, error) {
	return s.loadRepo.GetByID(ctx, id)
}

// GetAllLoads получает все грузы с фильтрацией и пагинацией
func (s *loadService) GetAllLoads(ctx context.Context, filter *models.LoadFilter, page, limit int) ([]*models.Load, *models.Pagination, error) {
	offset := (page - 1) * limit

	loads, total, err := s.loadRepo.GetAll(ctx, filter, limit, offset)
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

	return loads, pagination, nil
}

// UpdateLoad обновляет груз
func (s *loadService) UpdateLoad(ctx context.Context, id primitive.ObjectID, load *models.Load) error {
	// Проверяем, существует ли груз
	existing, err := s.loadRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return &ValidationError{Message: "Load not found"}
	}

	// Валидация
	if err := s.validateLoad(ctx, load); err != nil {
		return err
	}

	return s.loadRepo.Update(ctx, id, load)
}

// DeleteLoad удаляет груз
func (s *loadService) DeleteLoad(ctx context.Context, id primitive.ObjectID) error {
	// Проверяем, существует ли груз
	existing, err := s.loadRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return &ValidationError{Message: "Load not found"}
	}

	return s.loadRepo.Delete(ctx, id)
}

// GetLoadsByBroker получает грузы по брокеру
func (s *loadService) GetLoadsByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Load, *models.Pagination, error) {
	offset := (page - 1) * limit

	loads, total, err := s.loadRepo.GetByBroker(ctx, brokerID, limit, offset)
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

	return loads, pagination, nil
}

// GetLoadsByInvoice получает грузы по счету
func (s *loadService) GetLoadsByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Load, error) {
	return s.loadRepo.GetByInvoice(ctx, invoiceID)
}

// UpdateLoadStatus обновляет статус груза
func (s *loadService) UpdateLoadStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	// Проверяем, существует ли груз
	existing, err := s.loadRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return &ValidationError{Message: "Load not found"}
	}

	// Валидация статуса
	if !isValidLoadStatus(status) {
		return &ValidationError{Message: "Invalid load status"}
	}

	return s.loadRepo.UpdateStatus(ctx, id, status)
}

// GetUnbilledLoadsByBroker получает неоплаченные грузы брокера
func (s *loadService) GetUnbilledLoadsByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Load, *models.Pagination, error) {
	offset := (page - 1) * limit

	// Получаем грузы брокера которые:
	// 1. Не привязаны к инвойсу (invoice_id пустой)
	// 2. Привязаны к неоплаченному инвойсу (статус pending, partial, overdue)
	loads, total, err := s.loadRepo.GetUnbilledByBroker(ctx, brokerID, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	totalPages := (int(total) + limit - 1) / limit
	pagination := &models.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      int64(total),
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	return loads, pagination, nil
}

// validateLoad валидирует данные груза
func (s *loadService) validateLoad(ctx context.Context, load *models.Load) error {
	if load.BrokerID.IsZero() {
		return &ValidationError{Message: "Broker ID is required"}
	}

	// Проверяем, существует ли брокер
	broker, err := s.brokerRepo.GetByID(ctx, load.BrokerID)
	if err != nil {
		return err
	}
	if broker == nil {
		return &ValidationError{Message: "Broker not found"}
	}

	// Проверяем, существует ли счет (если указан)
	if !load.InvoiceID.IsZero() {
		invoice, err := s.invoiceRepo.GetByID(ctx, load.InvoiceID)
		if err != nil {
			return err
		}
		if invoice == nil {
			return &ValidationError{Message: "Invoice not found"}
		}

		// Проверяем, что брокер счета совпадает с брокером груза
		if invoice.BrokerID != load.BrokerID {
			return &ValidationError{Message: "Load broker must match invoice broker"}
		}
	}

	if load.Cost <= 0 {
		return &ValidationError{Message: "Cost must be greater than zero"}
	}

	if load.Currency == "" {
		return &ValidationError{Message: "Currency is required"}
	}

	if load.Currency != models.CurrencyUSD &&
		load.Currency != models.CurrencyEUR &&
		load.Currency != models.CurrencyRUB {
		return &ValidationError{Message: "Unsupported currency"}
	}

	// Валидация маршрута
	if load.Route.Origin.Address == "" {
		return &ValidationError{Message: "Origin address is required"}
	}

	if load.Route.Origin.City == "" {
		return &ValidationError{Message: "Origin city is required"}
	}

	if load.Route.Origin.State == "" {
		return &ValidationError{Message: "Origin state is required"}
	}

	if load.Route.Destination.Address == "" {
		return &ValidationError{Message: "Destination address is required"}
	}

	if load.Route.Destination.City == "" {
		return &ValidationError{Message: "Destination city is required"}
	}

	if load.Route.Destination.State == "" {
		return &ValidationError{Message: "Destination state is required"}
	}

	// Валидация статуса
	if load.Status != "" && !isValidLoadStatus(load.Status) {
		return &ValidationError{Message: "Invalid load status"}
	}

	return nil
}

// isValidLoadStatus проверяет валидность статуса груза
func isValidLoadStatus(status string) bool {
	validStatuses := []string{
		models.LoadStatusPlanned,
		models.LoadStatusInTransit,
		models.LoadStatusDelivered,
		models.LoadStatusCanceled,
	}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}

	return false
}
