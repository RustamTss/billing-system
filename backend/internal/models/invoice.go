package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// InvoiceStatus статусы счетов
const (
	InvoiceStatusPending  = "pending"
	InvoiceStatusPartial  = "partial"
	InvoiceStatusPaid     = "paid"
	InvoiceStatusOverdue  = "overdue"
	InvoiceStatusCanceled = "canceled"
)

// Currency валюты
const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyRUB = "RUB"
)

// Invoice представляет счет
type Invoice struct {
	ID            primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	InvoiceNumber string               `json:"invoice_number" bson:"invoice_number" validate:"required"`
	BrokerID      primitive.ObjectID   `json:"broker_id" bson:"broker_id" validate:"required"`
	Amount        float64              `json:"amount" bson:"amount" validate:"required,gt=0"`
	PaidAmount    float64              `json:"paid_amount" bson:"paid_amount"`
	Currency      string               `json:"currency" bson:"currency" validate:"required"`
	Status        string               `json:"status" bson:"status"`
	CreatedAt     time.Time            `json:"created_at" bson:"created_at"`
	DueDate       time.Time            `json:"due_date" bson:"due_date" validate:"required"`
	PaidAt        *time.Time           `json:"paid_at" bson:"paid_at"`
	Description   string               `json:"description" bson:"description"`
	LoadIDs       []primitive.ObjectID `json:"load_ids" bson:"load_ids"`
	Notes         string               `json:"notes" bson:"notes"`

	// Calculated fields
	IsOverdue       bool    `json:"is_overdue" bson:"-"`
	RemainingAmount float64 `json:"remaining_amount" bson:"-"`

	// Computed fields from JOINs (не сохраняются в БД)
	BrokerName string `json:"broker_name" bson:"broker_name,omitempty"`
}

// InvoiceWithBroker счет с информацией о брокере
type InvoiceWithBroker struct {
	Invoice
	Broker Broker `json:"broker" bson:"broker"`
}

// InvoiceFilter фильтры для поиска счетов
type InvoiceFilter struct {
	Status     []string           `json:"status"`
	BrokerID   primitive.ObjectID `json:"broker_id"`
	Currency   string             `json:"currency"`
	DateFrom   *time.Time         `json:"date_from"`
	DateTo     *time.Time         `json:"date_to"`
	AmountFrom *float64           `json:"amount_from"`
	AmountTo   *float64           `json:"amount_to"`
	IsOverdue  *bool              `json:"is_overdue"`
}
