# Этап сборки
FROM node:18 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Увеличиваем лимит памяти для Node.js
ENV NODE_OPTIONS="--max-old-space-size=2048"

# Копируем package.json и package-lock.json
COPY package*.json ./

# Очищаем npm кэш и устанавливаем зависимости
RUN npm cache clean --force && \
    rm -rf node_modules package-lock.json && \
    npm install --production=false

# Копируем исходный код
COPY . .

# Изменяем baseURL в API конфигурации для продакшена
RUN sed -i 's|http://165.232.113.23:8081/api/v1|/api/v1|g' src/services/api.js

# Собираем приложение с увеличенной памятью
RUN NODE_OPTIONS="--max-old-space-size=2048" npm run build

# Этап запуска с nginx
FROM nginx:alpine

# Удаляем дефолтную страницу nginx
RUN rm -rf /usr/share/nginx/html/*

# Копируем собранное приложение
COPY --from=builder /app/dist /usr/share/nginx/html

# Копируем конфигурацию nginx
COPY nginx.conf /etc/nginx/nginx.conf

# Открываем порт
EXPOSE 80

# Healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:80 || exit 1

# Запускаем nginx
CMD ["nginx", "-g", "daemon off;"] 