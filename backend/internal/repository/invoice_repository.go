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

// invoiceRepository реализация InvoiceRepository
type invoiceRepository struct {
	collection *mongo.Collection
}

// NewInvoiceRepository создает новый InvoiceRepository
func NewInvoiceRepository(db *Database) InvoiceRepository {
	return &invoiceRepository{
		collection: db.GetCollection("invoices"),
	}
}

// Create создает новый счет
func (r *invoiceRepository) Create(ctx context.Context, invoice *models.Invoice) error {
	invoice.ID = primitive.NewObjectID()
	invoice.CreatedAt = time.Now()
	invoice.PaidAmount = 0

	if invoice.Status == "" {
		invoice.Status = models.InvoiceStatusPending
	}

	// Генерируем номер счета если не указан
	if invoice.InvoiceNumber == "" {
		number, err := r.GenerateInvoiceNumber(ctx)
		if err != nil {
			return err
		}
		invoice.InvoiceNumber = number
	}

	_, err := r.collection.InsertOne(ctx, invoice)
	return err
}

// GetByID получает счет по ID
func (r *invoiceRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&invoice)
	if err != nil {
		return nil, err
	}

	// Вычисляем поля
	r.calculateFields(&invoice)

	return &invoice, nil
}

// GetAll получает все счета с фильтрацией и пагинацией
func (r *invoiceRepository) GetAll(ctx context.Context, filter *models.InvoiceFilter, limit, offset int) ([]*models.Invoice, int64, error) {
	// Используем aggregation pipeline для JOIN с brokers
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
		// Добавляем поле broker_name
		{
			"$addFields": bson.M{
				"broker_name": bson.M{
					"$ifNull": []interface{}{
						bson.M{"$arrayElemAt": []interface{}{"$broker.company_name", 0}},
						"Unknown Broker",
					},
				},
			},
		},
		// Удаляем временное поле
		{
			"$project": bson.M{
				"broker": 0,
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

	var invoices []*models.Invoice
	if err = cursor.All(ctx, &invoices); err != nil {
		return nil, 0, err
	}

	// Вычисляем поля для каждого счета
	for _, invoice := range invoices {
		r.calculateFields(invoice)
	}

	return invoices, total, nil
}

// Update обновляет счет
func (r *invoiceRepository) Update(ctx context.Context, id primitive.ObjectID, invoice *models.Invoice) error {
	update := bson.M{
		"$set": bson.M{
			"broker_id":   invoice.BrokerID,
			"amount":      invoice.Amount,
			"currency":    invoice.Currency,
			"due_date":    invoice.DueDate,
			"description": invoice.Description,
			"load_ids":    invoice.LoadIDs,
			"notes":       invoice.Notes,
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete удаляет счет
func (r *invoiceRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// GetByStatus получает счета по статусу
func (r *invoiceRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Invoice, int64, error) {
	filter := bson.M{"status": status}

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

	var invoices []*models.Invoice
	if err = cursor.All(ctx, &invoices); err != nil {
		return nil, 0, err
	}

	// Вычисляем поля для каждого счета
	for _, invoice := range invoices {
		r.calculateFields(invoice)
	}

	return invoices, total, nil
}

// GetByBroker получает счета по брокеру
func (r *invoiceRepository) GetByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Invoice, int64, error) {
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

	var invoices []*models.Invoice
	if err = cursor.All(ctx, &invoices); err != nil {
		return nil, 0, err
	}

	// Вычисляем поля для каждого счета
	for _, invoice := range invoices {
		r.calculateFields(invoice)
	}

	return invoices, total, nil
}

// GetOverdue получает просроченные счета
func (r *invoiceRepository) GetOverdue(ctx context.Context, limit, offset int) ([]*models.Invoice, int64, error) {
	now := time.Now()
	filter := bson.M{
		"due_date": bson.M{"$lt": now},
		"status":   bson.M{"$nin": []string{models.InvoiceStatusPaid, models.InvoiceStatusCanceled}},
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
		SetSort(bson.M{"due_date": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var invoices []*models.Invoice
	if err = cursor.All(ctx, &invoices); err != nil {
		return nil, 0, err
	}

	// Вычисляем поля для каждого счета
	for _, invoice := range invoices {
		r.calculateFields(invoice)
	}

	return invoices, total, nil
}

// UpdateStatus обновляет статус счета и сумму к оплате
func (r *invoiceRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status string, paidAmount float64) error {
	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"paid_amount": paidAmount,
		},
	}

	// Если статус "оплачен", ставим дату оплаты
	if status == models.InvoiceStatusPaid {
		update["$set"].(bson.M)["paid_at"] = time.Now()
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// GenerateInvoiceNumber генерирует номер счета
func (r *invoiceRepository) GenerateInvoiceNumber(ctx context.Context) (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("INV-%d%02d-", now.Year(), now.Month())

	// Ищем последний номер за текущий месяц
	filter := bson.M{
		"invoice_number": bson.M{
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

	var lastInvoice models.Invoice
	if cursor.Next(ctx) {
		if err := cursor.Decode(&lastInvoice); err != nil {
			return "", err
		}
	}

	// Извлекаем номер и увеличиваем на 1
	nextNumber := 1
	if lastInvoice.InvoiceNumber != "" {
		// Парсим номер из строки вида "INV-2024-01-0001"
		var num int
		fmt.Sscanf(lastInvoice.InvoiceNumber, prefix+"%d", &num)
		nextNumber = num + 1
	}

	return fmt.Sprintf("%s%04d", prefix, nextNumber), nil
}

// buildFilter строит MongoDB фильтр из структуры фильтра
func (r *invoiceRepository) buildFilter(filter *models.InvoiceFilter) bson.M {
	mongoFilter := bson.M{}

	if filter == nil {
		return mongoFilter
	}

	if len(filter.Status) > 0 {
		mongoFilter["status"] = bson.M{"$in": filter.Status}
	}

	if !filter.BrokerID.IsZero() {
		mongoFilter["broker_id"] = filter.BrokerID
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
		mongoFilter["created_at"] = dateFilter
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

	if filter.IsOverdue != nil && *filter.IsOverdue {
		mongoFilter["due_date"] = bson.M{"$lt": time.Now()}
		mongoFilter["status"] = bson.M{"$nin": []string{models.InvoiceStatusPaid, models.InvoiceStatusCanceled}}
	}

	return mongoFilter
}

// calculateFields вычисляет дополнительные поля
func (r *invoiceRepository) calculateFields(invoice *models.Invoice) {
	// Проверяем, просрочен ли счет
	invoice.IsOverdue = time.Now().After(invoice.DueDate) &&
		invoice.Status != models.InvoiceStatusPaid &&
		invoice.Status != models.InvoiceStatusCanceled

	// Вычисляем оставшуюся сумму
	invoice.RemainingAmount = invoice.Amount - invoice.PaidAmount
}
