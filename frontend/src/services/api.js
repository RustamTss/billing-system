import axios from 'axios'

// Базовая настройка API клиента
const api = axios.create({
  baseURL: 'http://165.232.113.23:8081/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Интерсептор для запросов
api.interceptors.request.use(
  (config) => {
    // Добавляем токен авторизации если есть
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Интерсептор для ответов
api.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      // Обработка неавторизованного доступа
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API для авторизации
export const authApi = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  getProfile: () => api.get('/auth/profile'),
  validateToken: () => api.get('/auth/validate'),
}

// API для дашборда
export const dashboardApi = {
  getMetrics: () => api.get('/dashboard/metrics'),
}

// API для брокеров
export const brokersApi = {
  getAll: (params) => api.get('/brokers', { params }),
  getById: (id) => api.get(`/brokers/${id}`),
  create: (data) => api.post('/brokers', data),
  update: (id, data) => api.put(`/brokers/${id}`, data),
  delete: (id) => api.delete(`/brokers/${id}`),
  search: (query, params) => api.get(`/brokers/search?query=${query}`, { params }),
  getStats: (id) => api.get(`/brokers/${id}/stats`),
  getInvoices: (id, params) => api.get(`/brokers/${id}/invoices`, { params }),
  getPayments: (id, params) => api.get(`/brokers/${id}/payments`, { params }),
}

// API для счетов
export const invoicesApi = {
  getAll: (params) => api.get('/invoices', { params }),
  getById: (id) => api.get(`/invoices/${id}`),
  create: (data) => api.post('/invoices', data),
  update: (id, data) => api.put(`/invoices/${id}`, data),
  delete: (id) => api.delete(`/invoices/${id}`),
  getByStatus: (status, params) => api.get(`/invoices/status/${status}`, { params }),
  getOverdue: (params) => api.get('/invoices/overdue', { params }),
  getPayments: (id) => api.get(`/invoices/${id}/payments`),
}

// API для платежей
export const paymentsApi = {
  getAll: (params) => api.get('/payments', { params }),
  getById: (id) => api.get(`/payments/${id}`),
  create: (data) => api.post('/payments', data),
  update: (id, data) => api.put(`/payments/${id}`, data),
  delete: (id) => api.delete(`/payments/${id}`),
}

// API для грузов
export const loadsApi = {
  getAll: (params) => api.get('/loads', { params }),
  getById: (id) => api.get(`/loads/${id}`),
  create: (data) => api.post('/loads', data),
  update: (id, data) => api.put(`/loads/${id}`, data),
  delete: (id) => api.delete(`/loads/${id}`),
  updateStatus: (id, status) => api.put(`/loads/${id}/status`, { status }),
  getUnbilledByBroker: (brokerId, params) => api.get(`/brokers/${brokerId}/loads/unbilled`, { params }),
}

// API для экспорта
export const exportApi = {
  invoices: (data) => api.post('/export/invoices', data, { responseType: 'blob' }),
  payments: (data) => api.post('/export/payments', data, { responseType: 'blob' }),
  brokers: (data) => api.post('/export/brokers', data, { responseType: 'blob' }),
}

// Административные API
export const adminApi = {
  sendOverdueNotifications: () => api.post('/admin/send-overdue-notifications'),
}

export default api 