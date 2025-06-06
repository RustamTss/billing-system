#!/bin/bash

# Скрипт деплоя Billing System на продакшен сервер
# Использование: ./deploy.sh [environment]

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

# Проверяем параметры
ENVIRONMENT=${1:-production}
log "Запуск деплоя в режиме: $ENVIRONMENT"

# Проверяем что Docker установлен
if ! command -v docker &> /dev/null; then
    error "Docker не установлен! Установите Docker и попробуйте снова."
fi

if ! command -v docker-compose &> /dev/null; then
    error "Docker Compose не установлен! Установите Docker Compose и попробуйте снова."
fi

# Проверяем что .env файл существует
if [ ! -f .env ]; then
    error "Файл .env не найден! Создайте .env файл с необходимыми переменными."
fi

log "Загружаем переменные окружения..."
source .env

# Проверяем обязательные переменные
required_vars=("MONGO_ROOT_USERNAME" "MONGO_ROOT_PASSWORD" "JWT_SECRET")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        error "Переменная окружения $var не установлена в .env файле"
    fi
done

log "Проверяем доступность Docker..."
if ! docker info &> /dev/null; then
    error "Docker daemon не запущен или недоступен"
fi

# Останавливаем старые контейнеры
log "Останавливаем старые контейнеры..."
if [ "$ENVIRONMENT" == "production" ]; then
    docker-compose -f docker-compose.prod.yml down || warning "Не удалось остановить контейнеры (возможно они не запущены)"
else
    docker-compose down || warning "Не удалось остановить контейнеры (возможно они не запущены)"
fi

# Создаем необходимые директории
log "Создаем необходимые директории..."
mkdir -p nginx/logs
mkdir -p ssl

# Устанавливаем права доступа
chmod +x deploy.sh

# Собираем и запускаем контейнеры
log "Собираем и запускаем контейнеры..."
if [ "$ENVIRONMENT" == "production" ]; then
    log "Запуск в продакшен режиме..."
    docker-compose -f docker-compose.prod.yml up --build -d
else
    log "Запуск в режиме разработки..."
    docker-compose up --build -d
fi

# Ждем запуска сервисов
log "Ожидаем запуска сервисов..."
sleep 30

# Проверяем статус контейнеров
log "Проверяем статус контейнеров..."
if [ "$ENVIRONMENT" == "production" ]; then
    docker-compose -f docker-compose.prod.yml ps
else
    docker-compose ps
fi

# Проверяем health checks
log "Проверяем доступность сервисов..."

# Проверяем backend
max_attempts=30
attempt=1
while [ $attempt -le $max_attempts ]; do
    if curl -f -s http://localhost:8081/health > /dev/null 2>&1; then
        log "✅ Backend сервис доступен"
        break
    fi
    
    if [ $attempt -eq $max_attempts ]; then
        error "❌ Backend сервис недоступен после $max_attempts попыток"
    fi
    
    log "Попытка $attempt/$max_attempts: ожидаем запуска backend..."
    sleep 5
    ((attempt++))
done

# Проверяем frontend
if curl -f -s http://localhost/ > /dev/null 2>&1; then
    log "✅ Frontend сервис доступен"
else
    warning "⚠️  Frontend сервис может быть еще недоступен"
fi

# Очищаем старые образы
log "Очищаем старые Docker образы..."
docker image prune -f

# Показываем использование ресурсов
log "Использование ресурсов:"
docker stats --no-stream

log "🎉 Деплой завершен успешно!"
log "Backend доступен по адресу: http://localhost:8081"
log "Frontend доступен по адресу: http://localhost"

# Показываем полезные команды
echo ""
echo "Полезные команды для управления:"
echo "  - Просмотр логов: docker-compose logs -f"
echo "  - Остановка: docker-compose down"
echo "  - Перезапуск: docker-compose restart"
echo "  - Статус: docker-compose ps" 