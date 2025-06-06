package repository

import (
	"billing-system/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BrokerRepository интерфейс для работы с брокерами
type BrokerRepository interface {
	Create(ctx context.Context, broker *models.Broker) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Broker, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Broker, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, broker *models.Broker) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Search(ctx context.Context, query string, limit, offset int) ([]*models.Broker, int64, error)
	GetStats(ctx context.Context, brokerID primitive.ObjectID) (*models.BrokerStats, error)
}

// InvoiceRepository интерфейс для работы со счетами
type InvoiceRepository interface {
	Create(ctx context.Context, invoice *models.Invoice) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Invoice, error)
	GetAll(ctx context.Context, filter *models.InvoiceFilter, limit, offset int) ([]*models.Invoice, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, invoice *models.Invoice) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Invoice, int64, error)
	GetByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Invoice, int64, error)
	GetOverdue(ctx context.Context, limit, offset int) ([]*models.Invoice, int64, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string, paidAmount float64) error
	GenerateInvoiceNumber(ctx context.Context) (string, error)
}

// PaymentRepository интерфейс для работы с платежами
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Payment, error)
	GetAll(ctx context.Context, filter *models.PaymentFilter, limit, offset int) ([]*models.Payment, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, payment *models.Payment) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Payment, error)
	GetByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Payment, int64, error)
	GetTotalPaidAmount(ctx context.Context, invoiceID primitive.ObjectID) (float64, error)
}

// LoadRepository интерфейс для работы с грузами
type LoadRepository interface {
	Create(ctx context.Context, load *models.Load) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Load, error)
	GetAll(ctx context.Context, filter *models.LoadFilter, limit, offset int) ([]*models.Load, int64, error)
	Update(ctx context.Context, id primitive.ObjectID, load *models.Load) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	GetByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Load, int64, error)
	GetByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Load, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status string) error
	GenerateLoadNumber(ctx context.Context) (string, error)
	GetUnbilledByBroker(ctx context.Context, brokerID primitive.ObjectID, limit, offset int) ([]*models.Load, int64, error)
}
