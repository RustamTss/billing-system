package services

import (
	"billing-system/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BrokerService интерфейс для работы с брокерами
type BrokerService interface {
	CreateBroker(ctx context.Context, broker *models.Broker) error
	GetBroker(ctx context.Context, id primitive.ObjectID) (*models.Broker, error)
	GetAllBrokers(ctx context.Context, page, limit int) ([]*models.Broker, *models.Pagination, error)
	UpdateBroker(ctx context.Context, id primitive.ObjectID, broker *models.Broker) error
	DeleteBroker(ctx context.Context, id primitive.ObjectID) error
	SearchBrokers(ctx context.Context, query string, page, limit int) ([]*models.Broker, *models.Pagination, error)
	GetBrokerStats(ctx context.Context, brokerID primitive.ObjectID) (*models.BrokerStats, error)
}

// InvoiceService интерфейс для работы со счетами
type InvoiceService interface {
	CreateInvoice(ctx context.Context, invoice *models.Invoice) error
	GetInvoice(ctx context.Context, id primitive.ObjectID) (*models.Invoice, error)
	GetAllInvoices(ctx context.Context, filter *models.InvoiceFilter, page, limit int) ([]*models.Invoice, *models.Pagination, error)
	UpdateInvoice(ctx context.Context, id primitive.ObjectID, invoice *models.Invoice) error
	DeleteInvoice(ctx context.Context, id primitive.ObjectID) error
	GetInvoicesByStatus(ctx context.Context, status string, page, limit int) ([]*models.Invoice, *models.Pagination, error)
	GetInvoicesByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Invoice, *models.Pagination, error)
	GetOverdueInvoices(ctx context.Context, page, limit int) ([]*models.Invoice, *models.Pagination, error)
	UpdateInvoiceStatus(ctx context.Context, invoiceID primitive.ObjectID) error
	SendOverdueNotifications(ctx context.Context) error
}

// PaymentService интерфейс для работы с платежами
type PaymentService interface {
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPayment(ctx context.Context, id primitive.ObjectID) (*models.Payment, error)
	GetAllPayments(ctx context.Context, filter *models.PaymentFilter, page, limit int) ([]*models.Payment, *models.Pagination, error)
	UpdatePayment(ctx context.Context, id primitive.ObjectID, payment *models.Payment) error
	DeletePayment(ctx context.Context, id primitive.ObjectID) error
	GetPaymentsByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Payment, error)
	GetPaymentsByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Payment, *models.Pagination, error)
}

// LoadService интерфейс для работы с грузами
type LoadService interface {
	CreateLoad(ctx context.Context, load *models.Load) error
	GetLoad(ctx context.Context, id primitive.ObjectID) (*models.Load, error)
	GetAllLoads(ctx context.Context, filter *models.LoadFilter, page, limit int) ([]*models.Load, *models.Pagination, error)
	UpdateLoad(ctx context.Context, id primitive.ObjectID, load *models.Load) error
	DeleteLoad(ctx context.Context, id primitive.ObjectID) error
	GetLoadsByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Load, *models.Pagination, error)
	GetLoadsByInvoice(ctx context.Context, invoiceID primitive.ObjectID) ([]*models.Load, error)
	UpdateLoadStatus(ctx context.Context, id primitive.ObjectID, status string) error
	GetUnbilledLoadsByBroker(ctx context.Context, brokerID primitive.ObjectID, page, limit int) ([]*models.Load, *models.Pagination, error)
}

// DashboardService интерфейс для дашборда
type DashboardService interface {
	GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error)
	GetTopDebtors(ctx context.Context, limit int) ([]models.TopDebtor, error)
	GetPaymentsByPeriod(ctx context.Context, days int) ([]models.PaymentByDay, error)
	GetInvoicesByStatus(ctx context.Context) ([]models.InvoiceByStatus, error)
	GetRevenueByMonth(ctx context.Context, months int) ([]models.RevenueByMonth, error)
}

// EmailService интерфейс для отправки email
type EmailService interface {
	SendOverdueNotification(ctx context.Context, broker *models.Broker, invoices []*models.Invoice) error
	SendInvoiceCreated(ctx context.Context, broker *models.Broker, invoice *models.Invoice) error
	SendPaymentReceived(ctx context.Context, broker *models.Broker, payment *models.Payment, invoice *models.Invoice) error
}

// ExportService интерфейс для экспорта данных
type ExportService interface {
	ExportInvoices(ctx context.Context, format string, filter *models.InvoiceFilter) ([]byte, error)
	ExportPayments(ctx context.Context, format string, filter *models.PaymentFilter) ([]byte, error)
	ExportBrokers(ctx context.Context, format string) ([]byte, error)
}
