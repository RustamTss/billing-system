package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Broker представляет брокера/компанию-клиента
type Broker struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CompanyName      string             `json:"company_name" bson:"company_name" validate:"required"`
	ContactPerson    string             `json:"contact_person" bson:"contact_person"`
	Email            string             `json:"email" bson:"email" validate:"required,email"`
	Phone            string             `json:"phone" bson:"phone"`
	Address          Address            `json:"address" bson:"address"`
	CreditLimit      float64            `json:"credit_limit" bson:"credit_limit"`
	ReliabilityScore int                `json:"reliability_score" bson:"reliability_score"` // 1-10
	Status           string             `json:"status" bson:"status"`                       // active, inactive, suspended
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
	Notes            string             `json:"notes" bson:"notes"`
}

// Address структура для адреса
type Address struct {
	Street  string `json:"street" bson:"street"`
	City    string `json:"city" bson:"city"`
	State   string `json:"state" bson:"state"`
	Country string `json:"country" bson:"country"`
	ZipCode string `json:"zip_code" bson:"zip_code"`
}

// BrokerStats статистика по брокеру
type BrokerStats struct {
	BrokerID        primitive.ObjectID `json:"broker_id" bson:"broker_id"`
	TotalDebt       float64            `json:"total_debt" bson:"total_debt"`
	OverdueAmount   float64            `json:"overdue_amount" bson:"overdue_amount"`
	PaidThisMonth   float64            `json:"paid_this_month" bson:"paid_this_month"`
	InvoicesCount   int                `json:"invoices_count" bson:"invoices_count"`
	OverdueInvoices int                `json:"overdue_invoices" bson:"overdue_invoices"`
	LastPayment     *time.Time         `json:"last_payment" bson:"last_payment"`
}
