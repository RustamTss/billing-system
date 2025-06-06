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

// paymentRepository реализация PaymentRepository
type paymentRepository struct {
	collection *mongo.Collection
}

// NewPaymentRepository создает новый PaymentRepository
func NewPaymentRepository(db *Database) PaymentRepository {
	return &paymentRepository{
		collection: db.GetCollection("payments"),
	}
}

// Create создает новый платеж
func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	payment.ID = primitive.NewObjectID()
	payment.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, payment)
	return err
}

// GetByID получает платеж по ID
func (r *paymentRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Payment, error) {
	var payment models.Payment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&payment)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetAll получает все платежи с фильтрацией и пагинацией
func (r *paymentRepository) GetAll(ctx context.Context, filter *models.PaymentFilter, limit, offset int) ([]*models.Payment, int64, error) {
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
		// JOIN с invoices для получения номера инвойса
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
						"Unknown Invoice",
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
			"$sort": bson.M{"payment_date": -1},
		},
	}

	// Подсчет общего количества
	countPipeline := make([]bson.M, 3)
	copy(countPipeline, pipeline[:3]) // Копируем только match и lookup'ы для подсчета
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

	var payments []*models.Payment
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// Update обновляет платеж
func (r *paymentRepository) Update(ctx context.Context, id primitive.ObjectID, payment *models.Payment) error {
	update := bson.M{
		"$set": bson.M{
			"invoice_id":       payment.InvoiceID,
			"broker_id":        payment.BrokerID,
			"amount":           payment.Amount,
			"currency":         payment.Currency,
			"payment_date":     payment.PaymentDate,
			"payment_method":   payment.PaymentMethod,
			"transaction_id":   payment.TransactionID,
			"reference_number": payment.ReferenceNumber,
			"notes":            payment.Notes,
			"created_by":       payment.CreatedBy,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete удаляет платеж
func (r *paymentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// GetByInvoice получает все платежи по счету
func (r *paymentRepository) GetByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Payment, error) {
	filter := bson.M{"invoice_id": invoiceID}

	opts := options.Find().SetSort(bson.M{"payment_date": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payments []*models.Payment
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, err
	}

	return payments, nil
}

// GetByBroker получает платежи по брокеру
func (r *paymentRepository) GetByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Payment, int64, error) {
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
		SetSort(bson.M{"payment_date": -1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var payments []*models.Payment
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}

// GetTotalPaidAmount получает общую сумму платежей по счету
func (r *paymentRepository) GetTotalPaidAmount(ctx context.Context, invoiceID primitive.ObjectID) (float64, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{"invoice_id": invoiceID},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$amount"},
			},
		},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		Total float64 `bson:"total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
	}

	return result.Total, nil
}

// buildFilter строит MongoDB фильтр из структуры фильтра
func (r *paymentRepository) buildFilter(filter *models.PaymentFilter) bson.M {
	mongoFilter := bson.M{}

	if filter == nil {
		return mongoFilter
	}

	if !filter.InvoiceID.IsZero() {
		mongoFilter["invoice_id"] = filter.InvoiceID
	}

	if !filter.BrokerID.IsZero() {
		mongoFilter["broker_id"] = filter.BrokerID
	}

	if filter.PaymentMethod != "" {
		mongoFilter["payment_method"] = filter.PaymentMethod
	}

	if filter.Currency != "" {
		mongoFilter["currency"] = filter.Currency
	}

	if filter.DateFrom != nil || filter.DateTo != nil {
		dateFilter := bson.M{}
		if filter.DateFrom != nil {
			dateFilter["$gte"] = *filter.DateFrom
		}
		if filter.DateTo != nil {
			dateFilter["$lte"] = *filter.DateTo
		}
		mongoFilter["payment_date"] = dateFilter
	}

	if filter.AmountFrom != nil || filter.AmountTo != nil {
		amountFilter := bson.M{}
		if filter.AmountFrom != nil {
			amountFilter["$gte"] = *filter.AmountFrom
		}
		if filter.AmountTo != nil {
			amountFilter["$lte"] = *filter.AmountTo
		}
		mongoFilter["amount"] = amountFilter
	}

	return mongoFilter
}
