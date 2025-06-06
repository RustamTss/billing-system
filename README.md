# Billing Management System

Система управления счетами и платежами для транспортных компаний.

## 🚀 Функционал

- ✅ **Авторизация** с JWT токенами и ролями (admin/user)
- ✅ **Брокеры** - управление клиентами
- ✅ **Счета** - создание и отслеживание счетов
- ✅ **Платежи** - учет платежей и долгов
- ✅ **Грузы** - выбор грузов для счетов
- ✅ **Dashboard** - аналитика и метрики

## 🛠 Технологии

### Backend
- **Go** с Fiber фреймворком
- **MongoDB** для хранения данных
- **JWT** для авторизации
- **Docker** для контейнеризации

### Frontend
- **React** 18 с TypeScript
- **Ant Design** UI библиотека
- **React Query** для работы с API
- **React Router** для навигации

## 📋 Требования

- **Docker** и **Docker Compose**
- **Git** для клонирования
- **2GB RAM** минимум для сервера

## 🔧 Локальная разработка

### 1. Клонирование проекта
```bash
git clone https://github.com/YOUR_USERNAME/billing-system.git
cd billing-system
```

### 2. Запуск с Docker
```bash
# Сборка и запуск всех сервисов
docker-compose up --build

# Или в фоне
docker-compose up -d --build
```

### 3. Первый вход
- Откройте http://localhost
- Создайте первого админа через регистрацию
- Или используйте API:

```bash
curl -X POST http://localhost/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@billing.com", 
    "password": "admin123",
    "role": "admin"
  }'
```

## 🚀 Развертывание на сервере

### 1. Подготовка сервера

```bash
# Обновляем систему
sudo apt update && sudo apt upgrade -y

# Устанавливаем Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Устанавливаем Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Перезагружаемся
sudo reboot
```

### 2. Клонирование на сервер

```bash
git clone https://github.com/YOUR_USERNAME/billing-system.git
cd billing-system
```

### 3. Настройка переменных окружения

Создайте файл `.env`:

```bash
# Database
MONGO_ROOT_USERNAME=admin
MONGO_ROOT_PASSWORD=SuperSecurePassword123!
DATABASE_NAME=billing_system

# App
JWT_SECRET=super-secret-jwt-key-256-bit-long
APP_VERSION=latest
DOCKER_REGISTRY=your-registry.com

# Email (опционально)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@yourdomain.com
```

### 4. Запуск продакшен версии

```bash
# Запуск
docker-compose -f docker-compose.prod.yml up -d

# Проверка статуса
docker-compose -f docker-compose.prod.yml ps

# Логи
docker-compose -f docker-compose.prod.yml logs -f
```

## 🌐 Настройка домена и SSL

### 1. Настройка DNS
Добавьте A-запись:
```
yourdomain.com → IP_ВАШЕГО_СЕРВЕРА
```

### 2. SSL сертификат (Let's Encrypt)

```bash
# Устанавливаем Certbot
sudo apt install certbot

# Получаем сертификат
sudo certbot certonly --standalone -d yourdomain.com

# Копируем сертификаты
sudo mkdir -p ./ssl
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./ssl/
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./ssl/
sudo chown -R $USER:$USER ./ssl
```

## 🔒 Безопасность

### Изменить пароли по умолчанию:
- ✅ MongoDB root пароль
- ✅ JWT секретный ключ  
- ✅ Первый админ пароль

### Настроить файрвол:
```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable
```

## 📊 Мониторинг

### Проверка работы сервисов:
```bash
# Статус контейнеров
docker ps

# Логи
docker logs billing_backend_prod
docker logs billing_frontend_prod
docker logs billing_mongodb_prod

# Использование ресурсов
docker stats
```

### Health checks:
- **Backend:** http://yourdomain.com/api/health
- **Frontend:** http://yourdomain.com

## 🔄 Обновление

```bash
# Получаем последние изменения
git pull origin main

# Пересобираем и перезапускаем
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up --build -d

# Удаляем старые образы
docker image prune -f
```

## 📱 API Документация

### Авторизация
- `POST /api/auth/register` - Регистрация
- `POST /api/auth/login` - Вход
- `GET /api/auth/profile` - Профиль пользователя

### Основные endpoints
- `GET /api/brokers` - Список брокеров
- `GET /api/invoices` - Список счетов
- `GET /api/payments` - Список платежей
- `GET /api/loads` - Список грузов
- `GET /api/dashboard/metrics` - Метрики дашборда

## ❌ Устранение проблем

### Контейнер не запускается:
```bash
docker logs CONTAINER_NAME
```

### База данных недоступна:
```bash
docker exec -it billing_mongodb_prod mongo
```

### Проблемы с сетью:
```bash
docker network ls
docker network inspect billing_billing_network
```

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи контейнеров
2. Убедитесь что все порты открыты
3. Проверьте переменные окружения
4. Проверьте свободное место на диске

## 📄 Лицензия

MIT License - используйте свободно для коммерческих проектов. # Production deployment ready
