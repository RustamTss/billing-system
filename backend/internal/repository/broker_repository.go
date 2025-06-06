package repository

import (
	"billing-system/internal/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// brokerRepository реализация BrokerRepository
type brokerRepository struct {
	collection *mongo.Collection
}

// NewBrokerRepository создает новый BrokerRepository
func NewBrokerRepository(db *Database) BrokerRepository {
	return &brokerRepository{
		collection: db.GetCollection("brokers"),
	}
}

// Create создает нового брокера
func (r *brokerRepository) Create(ctx context.Context, broker *models.Broker) error {
	broker.ID = primitive.NewObjectID()
	broker.CreatedAt = time.Now()
	broker.UpdatedAt = time.Now()

	if broker.Status == "" {
		broker.Status = "active"
	}

	_, err := r.collection.InsertOne(ctx, broker)
	return err
}

// GetByID получает брокера по ID
func (r *brokerRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Broker, error) {
	var broker models.Broker
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&broker)
	if err != nil {
		return nil, err
	}
	return &broker, nil
}

// GetAll получает всех брокеров с пагинацией
func (r *brokerRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Broker, int64, error) {
	// Подсчет общего количества
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// Опции для пагинации
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var brokers []*models.Broker
	if err = cursor.All(ctx, &brokers); err != nil {
		return nil, 0, err
	}

	return brokers, total, nil
}

// Update обновляет брокера
func (r *brokerRepository) Update(ctx context.Context, id primitive.ObjectID, broker *models.Broker) error {
	broker.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"company_name":      broker.CompanyName,
			"contact_person":    broker.ContactPerson,
			"email":             broker.Email,
			"phone":             broker.Phone,
			"address":           broker.Address,
			"credit_limit":      broker.CreditLimit,
			"reliability_score": broker.ReliabilityScore,
			"status":            broker.Status,
			"notes":             broker.Notes,
			"updated_at":        broker.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete удаляет брокера
func (r *brokerRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Search поиск брокеров по запросу
func (r *brokerRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Broker, int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"company_name": bson.M{"$regex": query, "$options": "i"}},
			{"contact_person": bson.M{"$regex": query, "$options": "i"}},
			{"email": bson.M{"$regex": query, "$options": "i"}},
			{"phone": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	// Подсчет общего количества
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Опции для пагинации
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var brokers []*models.Broker
	if err = cursor.All(ctx, &brokers); err != nil {
		return nil, 0, err
	}

	return brokers, total, nil
}

// GetStats получает статистику по брокеру
func (r *brokerRepository) GetStats(ctx context.Context, brokerID primitive.ObjectID) (*models.BrokerStats, error) {
	// Эта функция будет реализована через агрегацию с другими коллекциями
	// Сейчас возвращаем базовую структуру
	stats := &models.BrokerStats{
		BrokerID:        brokerID,
		TotalDebt:       0,
		OverdueAmount:   0,
		PaidThisMonth:   0,
		InvoicesCount:   0,
		OverdueInvoices: 0,
		LastPayment:     nil,
	}

	return stats, nil
}
