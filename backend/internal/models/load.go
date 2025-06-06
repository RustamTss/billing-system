package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LoadStatus статусы груза
const (
	LoadStatusPlanned   = "planned"
	LoadStatusInTransit = "in_transit"
	LoadStatusDelivered = "delivered"
	LoadStatusCanceled  = "canceled"
)

// Load представляет груз/рейс
type Load struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	LoadNumber   string             `json:"load_number" bson:"load_number" validate:"required"`
	BrokerID     primitive.ObjectID `json:"broker_id" bson:"broker_id" validate:"required"`
	InvoiceID    primitive.ObjectID `json:"invoice_id" bson:"invoice_id"`
	Route        Route              `json:"route" bson:"route" validate:"required"`
	PickupDate   time.Time          `json:"pickup_date" bson:"pickup_date"`
	DeliveryDate time.Time          `json:"delivery_date" bson:"delivery_date"`
	Cost         float64            `json:"cost" bson:"cost" validate:"required,gt=0"`
	Currency     string             `json:"currency" bson:"currency" validate:"required"`
	Status       string             `json:"status" bson:"status"`
	Weight       float64            `json:"weight" bson:"weight"`
	Distance     float64            `json:"distance" bson:"distance"`   // в милях
	Equipment    string             `json:"equipment" bson:"equipment"` // тип трейлера
	DriverInfo   DriverInfo         `json:"driver_info" bson:"driver_info"`
	Notes        string             `json:"notes" bson:"notes"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`

	// Computed fields from JOINs (не сохраняются в БД)
	BrokerName    string `json:"broker_name" bson:"broker_name,omitempty"`
	InvoiceNumber string `json:"invoice_number" bson:"invoice_number,omitempty"`
}

// Route маршрут
type Route struct {
	Origin      Location `json:"origin" bson:"origin" validate:"required"`
	Destination Location `json:"destination" bson:"destination" validate:"required"`
}

// Location локация
type Location struct {
	Address   string  `json:"address" bson:"address" validate:"required"`
	City      string  `json:"city" bson:"city" validate:"required"`
	State     string  `json:"state" bson:"state" validate:"required"`
	ZipCode   string  `json:"zip_code" bson:"zip_code"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

// DriverInfo информация о водителе
type DriverInfo struct {
	Name          string `json:"name" bson:"name"`
	Phone         string `json:"phone" bson:"phone"`
	TruckNumber   string `json:"truck_number" bson:"truck_number"`
	TrailerNumber string `json:"trailer_number" bson:"trailer_number"`
}

// LoadWithDetails груз с детальной информацией
type LoadWithDetails struct {
	Load
	Broker  Broker   `json:"broker" bson:"broker"`
	Invoice *Invoice `json:"invoice" bson:"invoice"`
}

// LoadFilter фильтры для поиска грузов
type LoadFilter struct {
	BrokerID    primitive.ObjectID `json:"broker_id"`
	InvoiceID   primitive.ObjectID `json:"invoice_id"`
	Status      []string           `json:"status"`
	DateFrom    *time.Time         `json:"date_from"`
	DateTo      *time.Time         `json:"date_to"`
	OriginState string             `json:"origin_state"`
	DestState   string             `json:"dest_state"`
}
