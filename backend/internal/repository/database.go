package repository

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database представляет подключение к базе данных
type Database struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// NewDatabase создает новое подключение к MongoDB
func NewDatabase(mongoURI, dbName string) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Опции подключения
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Подключение к MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Проверка подключения
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("Connected to MongoDB: %s", dbName)

	return &Database{
		Client: client,
		DB:     client.Database(dbName),
	}, nil
}

// Disconnect закрывает подключение к базе данных
func (d *Database) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.Client.Disconnect(ctx)
}

// GetCollection возвращает коллекцию по имени
func (d *Database) GetCollection(name string) *mongo.Collection {
	return d.DB.Collection(name)
}

// Repositories структура для всех репозиториев
type Repositories struct {
	Broker  BrokerRepository
	Invoice InvoiceRepository
	Payment PaymentRepository
	Load    LoadRepository
}

// NewRepositories создает новые репозитории
func NewRepositories(db *Database) *Repositories {
	return &Repositories{
		Broker:  NewBrokerRepository(db),
		Invoice: NewInvoiceRepository(db),
		Payment: NewPaymentRepository(db),
		Load:    NewLoadRepository(db),
	}
}
