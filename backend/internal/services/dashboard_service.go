package services

import (
	"billing-system/internal/models"
	"billing-system/internal/repository"
	"context"
	"time"
)

// dashboardService реализация DashboardService
type dashboardService struct {
	repos *repository.Repositories
}

// NewDashboardService создает новый DashboardService
func NewDashboardService(repos *repository.Repositories) DashboardService {
	return &dashboardService{
		repos: repos,
	}
}

// GetDashboardMetrics получает все метрики для дашборда
func (s *dashboardService) GetDashboardMetrics(ctx context.Context) (*models.DashboardMetrics, error) {
	metrics := &models.DashboardMetrics{}

	// Получаем базовые метрики параллельно
	totalDebt, err := s.calculateTotalDebt(ctx)
	if err != nil {
		return nil, err
	}
	metrics.TotalDebt = totalDebt

	overdueAmount, err := s.calculateOverdueAmount(ctx)
	if err != nil {
		return nil, err
	}
	metrics.OverdueAmount = overdueAmount

	paidThisMonth, err := s.calculatePaidThisMonth(ctx)
	if err != nil {
		return nil, err
	}
	metrics.PaidThisMonth = paidThisMonth

	paidLastMonth, err := s.calculatePaidLastMonth(ctx)
	if err != nil {
		return nil, err
	}
	metrics.PaidLastMonth = paidLastMonth

	// Подсчет счетов
	invoiceCounts, err := s.calculateInvoiceCounts(ctx)
	if err != nil {
		return nil, err
	}
	metrics.TotalInvoices = invoiceCounts.Total
	metrics.OverdueInvoices = invoiceCounts.Overdue
	metrics.PendingInvoices = invoiceCounts.Pending

	// Подсчет брокеров
	activeBrokers, err := s.calculateActiveBrokers(ctx)
	if err != nil {
		return nil, err
	}
	metrics.ActiveBrokers = activeBrokers

	// Подсчет грузов
	loadCounts, err := s.calculateLoadCounts(ctx)
	if err != nil {
		return nil, err
	}
	metrics.TotalLoads = loadCounts.Total
	metrics.CompletedLoads = loadCounts.Completed

	// Получаем топ должников
	topDebtors, err := s.GetTopDebtors(ctx, 10)
	if err != nil {
		return nil, err
	}
	metrics.TopDebtors = topDebtors

	// Получаем платежи по дням (последние 30 дней)
	paymentsByDay, err := s.GetPaymentsByPeriod(ctx, 30)
	if err != nil {
		return nil, err
	}
	metrics.PaymentsByDay = paymentsByDay

	// Получаем счета по статусам
	invoicesByStatus, err := s.GetInvoicesByStatus(ctx)
	if err != nil {
		return nil, err
	}
	metrics.InvoicesByStatus = invoicesByStatus

	// Получаем доходы по месяцам (последние 12 месяцев)
	revenueByMonth, err := s.GetRevenueByMonth(ctx, 12)
	if err != nil {
		return nil, err
	}
	metrics.RevenueByMonth = revenueByMonth

	return metrics, nil
}

// GetTopDebtors получает топ должников
func (s *dashboardService) GetTopDebtors(ctx context.Context, limit int) ([]models.TopDebtor, error) {
	// Используем агрегацию MongoDB для подсчета задолженности по брокерам
	// Это упрощенная версия - в реальности нужно джоинить коллекции

	// Пока возвращаем пустой массив
	// TODO: Реализовать агрегацию с джоинами между brokers и invoices
	return []models.TopDebtor{}, nil
}

// GetPaymentsByPeriod получает платежи за период
func (s *dashboardService) GetPaymentsByPeriod(ctx context.Context, days int) ([]models.PaymentByDay, error) {
	// Пока возвращаем пустой массив
	// TODO: Реализовать агрегацию платежей по дням
	return []models.PaymentByDay{}, nil
}

// GetInvoicesByStatus получает счета по статусам
func (s *dashboardService) GetInvoicesByStatus(ctx context.Context) ([]models.InvoiceByStatus, error) {
	// Пока возвращаем пустой массив
	// TODO: Реализовать агрегацию счетов по статусам
	return []models.InvoiceByStatus{}, nil
}

// GetRevenueByMonth получает доходы по месяцам
func (s *dashboardService) GetRevenueByMonth(ctx context.Context, months int) ([]models.RevenueByMonth, error) {
	// Пока возвращаем пустой массив
	// TODO: Реализовать агрегацию доходов по месяцам
	return []models.RevenueByMonth{}, nil
}

// calculateTotalDebt вычисляет общую задолженность
func (s *dashboardService) calculateTotalDebt(ctx context.Context) (float64, error) {
	// Получаем все неоплаченные инвойсы (pending, partial, overdue)
	filter := &models.InvoiceFilter{
		Status: []string{"pending", "partial", "overdue"},
	}

	invoices, _, err := s.repos.Invoice.GetAll(ctx, filter, 1000, 0) // Получаем все неоплаченные инвойсы
	if err != nil {
		return 0, err
	}

	var totalDebt float64
	for _, invoice := range invoices {
		remainingAmount := invoice.Amount - invoice.PaidAmount
		if remainingAmount > 0 {
			totalDebt += remainingAmount
		}
	}

	return totalDebt, nil
}

// calculateOverdueAmount вычисляет сумму просроченных платежей
func (s *dashboardService) calculateOverdueAmount(ctx context.Context) (float64, error) {
	// Получаем все просроченные инвойсы
	overdueInvoices, _, err := s.repos.Invoice.GetOverdue(ctx, 1000, 0)
	if err != nil {
		return 0, err
	}

	var overdueAmount float64
	for _, invoice := range overdueInvoices {
		remainingAmount := invoice.Amount - invoice.PaidAmount
		if remainingAmount > 0 {
			overdueAmount += remainingAmount
		}
	}

	return overdueAmount, nil
}

// calculatePaidThisMonth вычисляет сумму оплаченного в этом месяце
func (s *dashboardService) calculatePaidThisMonth(ctx context.Context) (float64, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	filter := &models.PaymentFilter{
		DateFrom: &startOfMonth,
		DateTo:   &endOfMonth,
	}

	payments, _, err := s.repos.Payment.GetAll(ctx, filter, 1000, 0) // Получаем все платежи
	if err != nil {
		return 0, err
	}

	var total float64
	for _, payment := range payments {
		total += payment.Amount
	}

	return total, nil
}

// calculatePaidLastMonth вычисляет сумму оплаченного в прошлом месяце
func (s *dashboardService) calculatePaidLastMonth(ctx context.Context) (float64, error) {
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	endOfLastMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Add(-time.Nanosecond)

	filter := &models.PaymentFilter{
		DateFrom: &startOfLastMonth,
		DateTo:   &endOfLastMonth,
	}

	payments, _, err := s.repos.Payment.GetAll(ctx, filter, 1000, 0) // Получаем все платежи
	if err != nil {
		return 0, err
	}

	var total float64
	for _, payment := range payments {
		total += payment.Amount
	}

	return total, nil
}

// InvoiceCounts структура для подсчета счетов
type InvoiceCounts struct {
	Total   int
	Overdue int
	Pending int
}

// calculateInvoiceCounts вычисляет количество счетов по категориям
func (s *dashboardService) calculateInvoiceCounts(ctx context.Context) (*InvoiceCounts, error) {
	// Общее количество инвойсов
	_, total, err := s.repos.Invoice.GetAll(ctx, nil, 1, 0)
	if err != nil {
		return nil, err
	}

	// Просроченные инвойсы
	_, overdueTotal, err := s.repos.Invoice.GetOverdue(ctx, 1, 0)
	if err != nil {
		return nil, err
	}

	// Pending инвойсы
	_, pendingTotal, err := s.repos.Invoice.GetByStatus(ctx, "pending", 1, 0)
	if err != nil {
		return nil, err
	}

	return &InvoiceCounts{
		Total:   int(total),
		Overdue: int(overdueTotal),
		Pending: int(pendingTotal),
	}, nil
}

// calculateActiveBrokers вычисляет количество активных брокеров
func (s *dashboardService) calculateActiveBrokers(ctx context.Context) (int, error) {
	_, total, err := s.repos.Broker.GetAll(ctx, 1, 0)
	if err != nil {
		return 0, err
	}
	return int(total), nil
}

// LoadCounts структура для подсчета грузов
type LoadCounts struct {
	Total     int
	Completed int
}

// calculateLoadCounts вычисляет количество грузов
func (s *dashboardService) calculateLoadCounts(ctx context.Context) (*LoadCounts, error) {
	// Общее количество грузов
	_, total, err := s.repos.Load.GetAll(ctx, nil, 1, 0)
	if err != nil {
		return nil, err
	}

	// Завершенные грузы (статус delivered)
	filter := &models.LoadFilter{
		Status: []string{"delivered"},
	}
	_, completedTotal, err := s.repos.Load.GetAll(ctx, filter, 1, 0)
	if err != nil {
		return nil, err
	}

	return &LoadCounts{
		Total:     int(total),
		Completed: int(completedTotal),
	}, nil
}
