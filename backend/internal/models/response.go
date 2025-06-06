package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// APIResponse общий формат ответа API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse ответ с пагинацией
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination информация о пагинации
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// DashboardMetrics метрики для дашборда
type DashboardMetrics struct {
	TotalDebt        float64           `json:"total_debt"`
	OverdueAmount    float64           `json:"overdue_amount"`
	PaidThisMonth    float64           `json:"paid_this_month"`
	PaidLastMonth    float64           `json:"paid_last_month"`
	TotalInvoices    int               `json:"total_invoices"`
	OverdueInvoices  int               `json:"overdue_invoices"`
	PendingInvoices  int               `json:"pending_invoices"`
	ActiveBrokers    int               `json:"active_brokers"`
	TotalLoads       int               `json:"total_loads"`
	CompletedLoads   int               `json:"completed_loads"`
	TopDebtors       []TopDebtor       `json:"top_debtors"`
	PaymentsByDay    []PaymentByDay    `json:"payments_by_day"`
	InvoicesByStatus []InvoiceByStatus `json:"invoices_by_status"`
	RevenueByMonth   []RevenueByMonth  `json:"revenue_by_month"`
}

// TopDebtor топ должники
type TopDebtor struct {
	BrokerID      primitive.ObjectID `json:"broker_id"`
	CompanyName   string             `json:"company_name"`
	TotalDebt     float64            `json:"total_debt"`
	OverdueAmount float64            `json:"overdue_amount"`
	Currency      string             `json:"currency"`
}

// PaymentByDay платежи по дням
type PaymentByDay struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
	Count  int       `json:"count"`
}

// InvoiceByStatus счета по статусам
type InvoiceByStatus struct {
	Status string  `json:"status"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

// RevenueByMonth доходы по месяцам
type RevenueByMonth struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
	Count   int     `json:"count"`
}

// SearchRequest запрос поиска
type SearchRequest struct {
	Query   string                 `json:"query"`
	Filters map[string]interface{} `json:"filters"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
	Sort    map[string]int         `json:"sort"` // field: 1 for asc, -1 for desc
}

// ExportRequest запрос экспорта
type ExportRequest struct {
	Format   string                 `json:"format"` // excel, pdf
	Type     string                 `json:"type"`   // invoices, payments, brokers
	Filters  map[string]interface{} `json:"filters"`
	DateFrom *time.Time             `json:"date_from"`
	DateTo   *time.Time             `json:"date_to"`
}
