package services

import (
	"billing-system/config"
	"billing-system/internal/models"
	"context"
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

// emailService реализация EmailService
type emailService struct {
	config config.EmailConfig
}

// NewEmailService создает новый EmailService
func NewEmailService(config config.EmailConfig) EmailService {
	return &emailService{
		config: config,
	}
}

// SendOverdueNotification отправляет уведомление о просроченных счетах
func (s *emailService) SendOverdueNotification(ctx context.Context, broker *models.Broker, invoices []*models.Invoice) error {
	if !s.isConfigured() {
		return nil // Пропускаем отправку если email не настроен
	}

	subject := fmt.Sprintf("Уведомление о просроченных счетах - %s", broker.CompanyName)

	// Формируем тело письма
	body := s.buildOverdueEmailBody(broker, invoices)

	return s.sendEmail(broker.Email, subject, body)
}

// SendInvoiceCreated отправляет уведомление о создании счета
func (s *emailService) SendInvoiceCreated(ctx context.Context, broker *models.Broker, invoice *models.Invoice) error {
	if !s.isConfigured() {
		return nil
	}

	subject := fmt.Sprintf("Новый счет %s - %s", invoice.InvoiceNumber, broker.CompanyName)

	body := s.buildInvoiceCreatedEmailBody(broker, invoice)

	return s.sendEmail(broker.Email, subject, body)
}

// SendPaymentReceived отправляет уведомление о получении платежа
func (s *emailService) SendPaymentReceived(ctx context.Context, broker *models.Broker, payment *models.Payment, invoice *models.Invoice) error {
	if !s.isConfigured() {
		return nil
	}

	subject := fmt.Sprintf("Платеж получен для счета %s - %s", invoice.InvoiceNumber, broker.CompanyName)

	body := s.buildPaymentReceivedEmailBody(broker, payment, invoice)

	return s.sendEmail(broker.Email, subject, body)
}

// sendEmail отправляет email
func (s *emailService) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.config.SMTPHost, s.config.SMTPPort, s.config.SMTPUsername, s.config.SMTPPassword)

	return d.DialAndSend(m)
}

// isConfigured проверяет, настроен ли email
func (s *emailService) isConfigured() bool {
	return s.config.SMTPHost != "" &&
		s.config.SMTPUsername != "" &&
		s.config.SMTPPassword != ""
}

// buildOverdueEmailBody формирует тело письма для просроченных счетов
func (s *emailService) buildOverdueEmailBody(broker *models.Broker, invoices []*models.Invoice) string {
	var totalAmount float64
	var invoicesList strings.Builder

	for _, invoice := range invoices {
		totalAmount += invoice.RemainingAmount
		invoicesList.WriteString(fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%s %.2f</td>
				<td>%s</td>
			</tr>
		`, invoice.InvoiceNumber,
			getCurrencySymbol(invoice.Currency),
			invoice.RemainingAmount,
			invoice.DueDate.Format("02.01.2006")))
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Просроченные счета</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5;">
	<div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
		<h1 style="color: #ff4d4f; border-bottom: 2px solid #ff4d4f; padding-bottom: 10px;">
			⚠️ Уведомление о просроченных счетах
		</h1>
		
		<p>Уважаемые коллеги из <strong>%s</strong>!</p>
		
		<p>Обращаем ваше внимание на то, что у вас имеются просроченные счета на общую сумму <strong style="color: #ff4d4f;">%.2f</strong>.</p>
		
		<h3>Детали просроченных счетов:</h3>
		<table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
			<thead>
				<tr style="background-color: #f0f0f0;">
					<th style="border: 1px solid #ddd; padding: 12px; text-align: left;">Номер счета</th>
					<th style="border: 1px solid #ddd; padding: 12px; text-align: left;">Сумма к доплате</th>
					<th style="border: 1px solid #ddd; padding: 12px; text-align: left;">Срок оплаты</th>
				</tr>
			</thead>
			<tbody>
				%s
			</tbody>
		</table>
		
		<p style="color: #666;">Просим вас произвести оплату в кратчайшие сроки. При возникновении вопросов обращайтесь к нашим менеджерам.</p>
		
		<div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; color: #888; font-size: 12px;">
			<p>С уважением,<br>Команда биллинг-системы</p>
			<p>Это автоматическое уведомление, пожалуйста, не отвечайте на это письмо.</p>
		</div>
	</div>
</body>
</html>
	`, broker.CompanyName, totalAmount, invoicesList.String())
}

// buildInvoiceCreatedEmailBody формирует тело письма для нового счета
func (s *emailService) buildInvoiceCreatedEmailBody(broker *models.Broker, invoice *models.Invoice) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Новый счет</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5;">
	<div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
		<h1 style="color: #1890ff; border-bottom: 2px solid #1890ff; padding-bottom: 10px;">
			📄 Новый счет выставлен
		</h1>
		
		<p>Уважаемые коллеги из <strong>%s</strong>!</p>
		
		<p>Для вас выставлен новый счет:</p>
		
		<div style="background-color: #f8f9fa; padding: 20px; border-radius: 4px; margin: 20px 0;">
			<table style="width: 100%%;">
				<tr>
					<td><strong>Номер счета:</strong></td>
					<td>%s</td>
				</tr>
				<tr>
					<td><strong>Сумма:</strong></td>
					<td style="color: #1890ff; font-size: 18px; font-weight: bold;">%s %.2f</td>
				</tr>
				<tr>
					<td><strong>Срок оплаты:</strong></td>
					<td>%s</td>
				</tr>
				<tr>
					<td><strong>Описание:</strong></td>
					<td>%s</td>
				</tr>
			</table>
		</div>
		
		<p style="color: #666;">Просим произвести оплату до указанного срока. Спасибо за сотрудничество!</p>
		
		<div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; color: #888; font-size: 12px;">
			<p>С уважением,<br>Команда биллинг-системы</p>
		</div>
	</div>
</body>
</html>
	`, broker.CompanyName,
		invoice.InvoiceNumber,
		getCurrencySymbol(invoice.Currency),
		invoice.Amount,
		invoice.DueDate.Format("02.01.2006"),
		invoice.Description)
}

// buildPaymentReceivedEmailBody формирует тело письма для полученного платежа
func (s *emailService) buildPaymentReceivedEmailBody(broker *models.Broker, payment *models.Payment, invoice *models.Invoice) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Платеж получен</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5;">
	<div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
		<h1 style="color: #52c41a; border-bottom: 2px solid #52c41a; padding-bottom: 10px;">
			✅ Платеж получен
		</h1>
		
		<p>Уважаемые коллеги из <strong>%s</strong>!</p>
		
		<p>Мы получили ваш платеж по счету <strong>%s</strong>.</p>
		
		<div style="background-color: #f6ffed; padding: 20px; border-radius: 4px; margin: 20px 0; border-left: 4px solid #52c41a;">
			<table style="width: 100%%;">
				<tr>
					<td><strong>Сумма платежа:</strong></td>
					<td style="color: #52c41a; font-size: 18px; font-weight: bold;">%s %.2f</td>
				</tr>
				<tr>
					<td><strong>Дата платежа:</strong></td>
					<td>%s</td>
				</tr>
				<tr>
					<td><strong>Способ оплаты:</strong></td>
					<td>%s</td>
				</tr>
				<tr>
					<td><strong>Номер транзакции:</strong></td>
					<td>%s</td>
				</tr>
			</table>
		</div>
		
		<p style="color: #666;">Спасибо за своевременную оплату! Ваш платеж обработан и учтен в системе.</p>
		
		<div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; color: #888; font-size: 12px;">
			<p>С уважением,<br>Команда биллинг-системы</p>
		</div>
	</div>
</body>
</html>
	`, broker.CompanyName,
		invoice.InvoiceNumber,
		getCurrencySymbol(payment.Currency),
		payment.Amount,
		payment.PaymentDate.Format("02.01.2006 15:04"),
		getPaymentMethodName(payment.PaymentMethod),
		payment.TransactionID)
}

// getCurrencySymbol возвращает символ валюты
func getCurrencySymbol(currency string) string {
	switch currency {
	case models.CurrencyUSD:
		return "$"
	case models.CurrencyEUR:
		return "€"
	case models.CurrencyRUB:
		return "₽"
	default:
		return currency
	}
}

// getPaymentMethodName возвращает читаемое название метода оплаты
func getPaymentMethodName(method string) string {
	switch method {
	case models.PaymentMethodWireTransfer:
		return "Банковский перевод"
	case models.PaymentMethodCheck:
		return "Чек"
	case models.PaymentMethodCash:
		return "Наличные"
	case models.PaymentMethodCard:
		return "Банковская карта"
	case models.PaymentMethodCrypto:
		return "Криптовалюта"
	default:
		return method
	}
}
