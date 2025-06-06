package services

import (
	"billing-system/internal/models"
	"billing-system/internal/repository"
	"context"
	"math"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// brokerService реализация BrokerService
type brokerService struct {
	brokerRepo repository.BrokerRepository
}

// NewBrokerService создает новый BrokerService
func NewBrokerService(brokerRepo repository.BrokerRepository) BrokerService {
	return &brokerService{
		brokerRepo: brokerRepo,
	}
}

// CreateBroker создает нового брокера
func (s *brokerService) CreateBroker(ctx context.Context, broker *models.Broker) error {
	// Валидация
	if err := s.validateBroker(broker); err != nil {
		return err
	}

	return s.brokerRepo.Create(ctx, broker)
}

// GetBroker получает брокера по ID
func (s *brokerService) GetBroker(ctx context.Context, id primitive.ObjectID) (*models.Broker, error) {
	return s.brokerRepo.GetByID(ctx, id)
}

// GetAllBrokers получает всех брокеров с пагинацией
func (s *brokerService) GetAllBrokers(ctx context.Context, page, limit int) ([]*models.Broker, *models.Pagination, error) {
	offset := (page - 1) * limit

	brokers, total, err := s.brokerRepo.GetAll(ctx, limit, offset)
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

	return brokers, pagination, nil
}

// UpdateBroker обновляет брокера
func (s *brokerService) UpdateBroker(ctx context.Context, id primitive.ObjectID, broker *models.Broker) error {
	// Проверяем, существует ли брокер
	existing, err := s.brokerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return &ValidationError{Message: "Broker not found"}
	}

	// Валидация
	if err := s.validateBroker(broker); err != nil {
		return err
	}

	return s.brokerRepo.Update(ctx, id, broker)
}

// DeleteBroker удаляет брокера
func (s *brokerService) DeleteBroker(ctx context.Context, id primitive.ObjectID) error {
	// Проверяем, существует ли брокер
	existing, err := s.brokerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return &ValidationError{Message: "Broker not found"}
	}

	// TODO: Проверить, есть ли связанные счета или грузы
	// Для безопасности можно запретить удаление брокера с активными счетами

	return s.brokerRepo.Delete(ctx, id)
}

// SearchBrokers поиск брокеров
func (s *brokerService) SearchBrokers(ctx context.Context, query string, page, limit int) ([]*models.Broker, *models.Pagination, error) {
	offset := (page - 1) * limit

	brokers, total, err := s.brokerRepo.Search(ctx, query, limit, offset)
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

	return brokers, pagination, nil
}

// GetBrokerStats получает статистику по брокеру
func (s *brokerService) GetBrokerStats(ctx context.Context, brokerID primitive.ObjectID) (*models.BrokerStats, error) {
	// Проверяем, существует ли брокер
	broker, err := s.brokerRepo.GetByID(ctx, brokerID)
	if err != nil {
		return nil, err
	}
	if broker == nil {
		return nil, &ValidationError{Message: "Broker not found"}
	}

	// Получаем статистику
	return s.brokerRepo.GetStats(ctx, brokerID)
}

// validateBroker валидирует данные брокера
func (s *brokerService) validateBroker(broker *models.Broker) error {
	if broker.CompanyName == "" {
		return &ValidationError{Message: "Company name is required"}
	}

	if broker.Email == "" {
		return &ValidationError{Message: "Email is required"}
	}

	// Простая валидация email
	if len(broker.Email) < 5 || !contains(broker.Email, "@") {
		return &ValidationError{Message: "Invalid email format"}
	}

	if broker.CreditLimit < 0 {
		return &ValidationError{Message: "Credit limit cannot be negative"}
	}

	if broker.ReliabilityScore < 0 || broker.ReliabilityScore > 10 {
		return &ValidationError{Message: "Reliability score must be between 0 and 10"}
	}

	return nil
}

// contains проверяет, содержит ли строка подстроку
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
