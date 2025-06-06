// Скрипт инициализации MongoDB
// Создает базу данных и первого админ пользователя

// Переключаемся на базу billing_system
db = db.getSiblingDB('billing_system');

// Создаем коллекции с индексами
db.createCollection('users');
db.createCollection('brokers');
db.createCollection('invoices');
db.createCollection('payments');
db.createCollection('loads');

// Создаем индексы для пользователей
db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "email": 1 }, { unique: true });

// Создаем индексы для остальных коллекций
db.brokers.createIndex({ "company_name": 1 });
db.brokers.createIndex({ "email": 1 });
db.invoices.createIndex({ "broker_id": 1 });
db.invoices.createIndex({ "status": 1 });
db.invoices.createIndex({ "due_date": 1 });
db.payments.createIndex({ "invoice_id": 1 });
db.payments.createIndex({ "broker_id": 1 });
db.loads.createIndex({ "broker_id": 1 });
db.loads.createIndex({ "status": 1 });

print('MongoDB инициализирован успешно!');
print('Создана база данных billing_system с коллекциями и индексами');
print('Первый администратор будет создан через API после запуска приложения'); 