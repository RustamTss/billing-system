package repository

import (
	"billing-system/internal/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// loadRepository реализация LoadRepository
type loadRepository struct {
	collection *mongo.Collection
}

// NewLoadRepository создает новый LoadRepository
func NewLoadRepository(db *Database) LoadRepository {
	return &loadRepository{
		collection: db.GetCollection("loads"),
	}
}

// Create создает новый груз
func (r *loadRepository) Create(ctx context.Context, load *models.Load) error {
	load.ID = primitive.NewObjectID()
	load.CreatedAt = time.Now()
	load.UpdatedAt = time.Now()

	if load.Status == "" {
		load.Status = models.LoadStatusPlanned
	}

	// Генерируем номер груза если не указан
	if load.LoadNumber == "" {
		number, err := r.GenerateLoadNumber(ctx)
		if err != nil {
			return err
		}
		load.LoadNumber = number
	}

	_, err := r.collection.InsertOne(ctx, load)
	return err
}

// GetByID получает груз по ID
func (r *loadRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Load, error) {
	var load models.Load
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&load)
	if err != nil {
		return nil, err
	}
	return &load, nil
}

// GetAll получает все грузы с фильтрацией и пагинацией
func (r *loadRepository) GetAll(ctx context.Context, filter *models.LoadFilter, limit, offset int) ([]*models.Load, int64, error) {
	// Используем aggregation pipeline для JOIN с brokers и invoices
	pipeline := []bson.M{
		// Первичная фильтрация
		{
			"$match": r.buildFilter(filter),
		},
		// JOIN с brokers для получения имени брокера
		{
			"$lookup": bson.M{
				"from": "brokers",
				"let":  bson.M{"broker_id": "$broker_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$eq": []interface{}{"$_id", "$$broker_id"},
							},
						},
					},
				},
				"as": "broker",
			},
		},
		// JOIN с invoices для получения информации об инвойсе
		{
			"$lookup": bson.M{
				"from": "invoices",
				"let":  bson.M{"invoice_id": "$invoice_id"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{
								"$eq": []interface{}{"$_id", "$$invoice_id"},
							},
						},
					},
				},
				"as": "invoice",
			},
		},
		// Добавляем поля broker_name и invoice_number
		{
			"$addFields": bson.M{
				"broker_name": bson.M{
					"$ifNull": []interface{}{
						bson.M{"$arrayElemAt": []interface{}{"$broker.company_name", 0}},
						"Unknown Broker",
					},
				},
				"invoice_number": bson.M{
					"$ifNull": []interface{}{
						bson.M{"$arrayElemAt": []interface{}{"$invoice.invoice_number", 0}},
						nil,
					},
				},
			},
		},
		// Удаляем временные поля
		{
			"$project": bson.M{
				"broker":  0,
				"invoice": 0,
			},
		},
		// Сортировка
		{
			"$sort": bson.M{"created_at": -1},
		},
	}

	// Подсчет общего количества
	countPipeline := make([]bson.M, 2)
	copy(countPipeline, pipeline[:2]) // Копируем только match и lookup для подсчета
	countPipeline = append(countPipeline, bson.M{"$count": "total"})

	countCursor, err := r.collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer countCursor.Close(ctx)

	var countResult []bson.M
	if err = countCursor.All(ctx, &countResult); err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if len(countResult) > 0 {
		if count, ok := countResult[0]["total"].(int32); ok {
			total = int64(count)
		} else if count, ok := countResult[0]["total"].(int64); ok {
			total = count
		}
	}

	// Добавляем пагинацию к основному pipeline
	dataPipeline := make([]bson.M, len(pipeline))
	copy(dataPipeline, pipeline)
	dataPipeline = append(dataPipeline, bson.M{"$skip": offset})
	dataPipeline = append(dataPipeline, bson.M{"$limit": limit})

	cursor, err := r.collection.Aggregate(ctx, dataPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var loads []*models.Load
	if err = cursor.All(ctx, &loads); err != nil {
		return nil, 0, err
	}

	return loads, total, nil
}

// Update обновляет груз
func (r *loadRepository) Update(ctx context.Context, id primitive.ObjectID, load *models.Load) error {
	load.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"broker_id":     load.BrokerID,
			"invoice_id":    load.InvoiceID,
			"route":         load.Route,
			"pickup_date":   load.PickupDate,
			"delivery_date": load.DeliveryDate,
			"cost":          load.Cost,
			"currency":      load.Currency,
			"status":        load.Status,
			"weight":        load.Weight,
			"distance":      load.Distance,
			"equipment":     load.Equipment,
			"driver_info":   load.DriverInfo,
			"notes":         load.Notes,
			"updated_at":    load.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete удаляет груз
func (r *loadRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// GetByBroker получает грузы по брокеру
func (r *loadRepository) GetByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Load, int64, error) {
	filter := bson.M{"broker_id": brokerID}

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

	var loads []*models.Load
	if err = cursor.All(ctx, &loads); err != nil {
		return nil, 0, err
	}

	return loads, total, nil
}

// GetByInvoice получает грузы по счету
func (r *loadRepository) GetByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Load, error) {
	filter := bson.M{"invoice_id": invoiceID}

	opts := options.Find().SetSort(bson.M{"pickup_date": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var loads []*models.Load
	if err = cursor.All(ctx, &loads); err != nil {
		return nil, err
	}

	return loads, nil
}

// UpdateStatus обновляет статус груза
func (r *loadRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// GenerateLoadNumber генерирует номер груза
func (r *loadRepository) GenerateLoadNumber(ctx context.Context) (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("LD-%d%02d%02d-", now.Year(), now.Month(), now.Day())

	// Ищем последний номер за текущий день
	filter := bson.M{
		"load_number": bson.M{
			"$regex": fmt.Sprintf("^%s", prefix),
		},
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(1)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return "", err
	}
	defer cursor.Close(ctx)

	var lastLoad models.Load
	if cursor.Next(ctx) {
		if err := cursor.Decode(&lastLoad); err != nil {
			return "", err
		}
	}

	// Извлекаем номер и увеличиваем на 1
	nextNumber := 1
	if lastLoad.LoadNumber != "" {
		// Парсим номер из строки вида "LD-20240115-001"
		var num int
		fmt.Sscanf(lastLoad.LoadNumber, prefix+"%d", &num)
		nextNumber = num + 1
	}

	return fmt.Sprintf("%s%03d", prefix, nextNumber), nil
}

// GetUnbilledByBroker получает грузы брокера, которые не оплачены или не привязаны к инвойсу
func (r *loadRepository) GetUnbilledByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Load, int64, error) {
	// Aggregation pipeline для поиска грузов, которые:
	// 1. Принадлежат указанному брокеру
	// 2. Либо не имеют invoice_id (равен null или ObjectID("000000000000000000000000"))
	// 3. Либо имеют invoice_id, но статус инвойса pending, partial, overdue
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"broker_id": brokerID,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "invoices",
				"localField":   "invoice_id",
				"foreignField": "_id",
				"as":           "invoice",
			},
		},
		{
			"$match": bson.M{
				"$or": []bson.M{
					// Грузы без инвойса
					{"invoice_id": bson.M{"$in": []interface{}{nil, primitive.NilObjectID}}},
					// Грузы с неоплаченными инвойсами
					{"invoice.status": bson.M{"$in": []string{"pending", "partial", "overdue"}}},
				},
			},
		},
		{
			"$project": bson.M{
				"invoice": 0, // Убираем invoice из результата
			},
		},
		{
			"$sort": bson.M{"created_at": -1},
		},
	}

	// Подсчет общего количества
	countPipeline := append(pipeline, bson.M{"$count": "total"})
	countCursor, err := r.collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}

	var countResult []bson.M
	if err = countCursor.All(ctx, &countResult); err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if len(countResult) > 0 {
		if count, ok := countResult[0]["total"].(int32); ok {
			total = int64(count)
		}
	}

	// Добавляем пагинацию
	dataPipeline := make([]bson.M, len(pipeline))
	copy(dataPipeline, pipeline)
	dataPipeline = append(dataPipeline, bson.M{"$skip": offset})
	dataPipeline = append(dataPipeline, bson.M{"$limit": limit})

	cursor, err := r.collection.Aggregate(ctx, dataPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var loads []*models.Load
	if err = cursor.All(ctx, &loads); err != nil {
		return nil, 0, err
	}

	return loads, total, nil
}

// buildFilter строит MongoDB фильтр из структуры фильтра
func (r *loadRepository) buildFilter(filter *models.LoadFilter) bson.M {
	mongoFilter := bson.M{}

	if filter == nil {
		return mongoFilter
	}

	if !filter.BrokerID.IsZero() {
		mongoFilter["broker_id"] = filter.BrokerID
	}

	if !filter.InvoiceID.IsZero() {
		mongoFilter["invoice_id"] = filter.InvoiceID
	}

	if len(filter.Status) > 0 {
		mongoFilter["status"] = bson.M{"$in": filter.Status}
	}

	if filter.DateFrom != nil || filter.DateTo != nil {
		dateFilter := bson.M{}
		if filter.DateFrom != nil {
			dateFilter["$gte"] = *filter.DateFrom
		}
		if filter.DateTo != nil {
			dateFilter["$lte"] = *filter.DateTo
		}
		mongoFilter["pickup_date"] = dateFilter
	}

	if filter.OriginState != "" {
		mongoFilter["route.origin.state"] = filter.OriginState
	}

	if filter.DestState != "" {
		mongoFilter["route.destination.state"] = filter.DestState
	}

	return mongoFilter
}
