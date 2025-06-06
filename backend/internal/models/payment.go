package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentMethod методы оплаты
const (
	PaymentMethodWireTransfer = "wire_transfer"
	PaymentMethodCheck        = "check"
	PaymentMethodCash         = "cash"
	PaymentMethodCard         = "card"
	PaymentMethodCrypto       = "crypto"
)

// Payment представляет платеж
type Payment struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	InvoiceID       primitive.ObjectID `json:"invoice_id" bson:"invoice_id" validate:"required"`
	BrokerID        primitive.ObjectID `json:"broker_id" bson:"broker_id" validate:"required"`
	Amount          float64            `json:"amount" bson:"amount" validate:"required,gt=0"`
	Currency        string             `json:"currency" bson:"currency" validate:"required"`
	PaymentDate     time.Time          `json:"payment_date" bson:"payment_date" validate:"required"`
	PaymentMethod   string             `json:"payment_method" bson:"payment_method" validate:"required"`
	TransactionID   string             `json:"transaction_id" bson:"transaction_id"`
	ReferenceNumber string             `json:"reference_number" bson:"reference_number"`
	Notes           string             `json:"notes" bson:"notes"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	CreatedBy       string             `json:"created_by" bson:"created_by"`

	// Computed fields from JOINs (не сохраняются в БД)
	BrokerName    string `json:"broker_name" bson:"broker_name,omitempty"`
	InvoiceNumber string `json:"invoice_number" bson:"invoice_number,omitempty"`
}

// PaymentWithInvoice платеж с информацией о счете
type PaymentWithInvoice struct {
	Payment
	Invoice Invoice `json:"invoice" bson:"invoice"`
	Broker  Broker  `json:"broker" bson:"broker"`
}

// PaymentFilter фильтры для поиска платежей
type PaymentFilter struct {
	InvoiceID     primitive.ObjectID `json:"invoice_id"`
	BrokerID      primitive.ObjectID `json:"broker_id"`
	PaymentMethod string             `json:"payment_method"`
	Currency      string             `json:"currency"`
	DateFrom      *time.Time         `json:"date_from"`
	DateTo        *time.Time         `json:"date_to"`
	AmountFrom    *float64           `json:"amount_from"`
	AmountTo      *float64           `json:"amount_to"`
}
