package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"billing-system/config"
	"billing-system/internal/handlers"
	"billing-system/internal/middleware"
	"billing-system/internal/repositories"
	"billing-system/internal/repository"
	"billing-system/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения")
	}

	// Загружаем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := repository.NewDatabase(cfg.Database.MongoURI, cfg.Database.DatabaseName)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer func() {
		if err := db.Disconnect(); err != nil {
			log.Printf("Ошибка отключения от базы данных: %v", err)
		}
	}()

	// Инициализируем репозитории
	repos := repository.NewRepositories(db)
	userRepo := repositories.NewUserRepository(db.DB)

	// Инициализируем сервисы
	emailService := services.NewEmailService(cfg.Email)
	authService := services.NewAuthService(userRepo)
	brokerService := services.NewBrokerService(repos.Broker)
	invoiceService := services.NewInvoiceService(repos.Invoice, repos.Payment, repos.Broker, emailService)
	paymentService := services.NewPaymentService(repos.Payment, repos.Invoice, repos.Broker, emailService)
	loadService := services.NewLoadService(repos.Load, repos.Broker, repos.Invoice)
	dashboardService := services.NewDashboardService(repos)

	// Устанавливаем взаимные зависимости
	// TODO: Реализовать правильную настройку зависимостей между сервисами

	// Создаем Fiber приложение
	app := fiber.New(fiber.Config{
		AppName:               cfg.App.Name,
		DisableStartupMessage: false,
		ErrorHandler:          middleware.ErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))

	// Инициализируем middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Инициализируем handlers
	h := handlers.NewHandlers(
		brokerService,
		invoiceService,
		paymentService,
		loadService,
		dashboardService,
	)
	authHandlers := handlers.NewAuthHandlers(authService)

	// Настраиваем маршруты
	setupRoutes(app, h, authHandlers, authMiddleware)

	// Запуск сервера в отдельной горутине
	go func() {
		addr := cfg.Server.Host + ":" + cfg.Server.Port
		log.Printf("Сервер запущен на %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Завершение работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Ошибка при завершении работы сервера: %v", err)
	}

	log.Println("Сервер остановлен")
}

// setupRoutes настраивает маршруты API
func setupRoutes(app *fiber.App, h *handlers.Handlers, authHandlers *handlers.AuthHandlers, authMiddleware *middleware.AuthMiddleware) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"timestamp": time.Now(),
		})
	})

	// API v1
	api := app.Group("/api/v1")

	// Auth routes (публичные)
	auth := api.Group("/auth")
	auth.Post("/login", authHandlers.Login)
	auth.Post("/register", authHandlers.Register)
	auth.Get("/profile", authMiddleware.RequireAuth(), authHandlers.GetProfile)
	auth.Get("/validate", authMiddleware.RequireAuth(), authHandlers.ValidateToken)

	// Защищенные маршруты
	protected := api.Group("/", authMiddleware.RequireAuth())

	// Dashboard routes
	dashboard := protected.Group("dashboard")
	dashboard.Get("/metrics", h.GetDashboardMetrics)

	// Brokers routes
	brokers := protected.Group("brokers")
	brokers.Get("/", h.GetBrokers)
	brokers.Post("/", h.CreateBroker)
	brokers.Get("/search", h.SearchBrokers)
	brokers.Get("/:id", h.GetBroker)
	brokers.Put("/:id", h.UpdateBroker)
	brokers.Delete("/:id", h.DeleteBroker)
	brokers.Get("/:id/stats", h.GetBrokerStats)
	brokers.Get("/:id/invoices", h.GetBrokerInvoices)
	brokers.Get("/:id/payments", h.GetBrokerPayments)
	brokers.Get("/:id/loads/unbilled", h.GetBrokerUnbilledLoads)

	// Invoices routes
	invoices := protected.Group("invoices")
	invoices.Get("/", h.GetInvoices)
	invoices.Post("/", h.CreateInvoice)
	invoices.Get("/overdue", h.GetOverdueInvoices)
	invoices.Get("/status/:status", h.GetInvoicesByStatus)
	invoices.Get("/:id", h.GetInvoice)
	invoices.Put("/:id", h.UpdateInvoice)
	invoices.Delete("/:id", h.DeleteInvoice)
	invoices.Get("/:id/payments", h.GetInvoicePayments)

	// Payments routes
	payments := protected.Group("payments")
	payments.Get("/", h.GetPayments)
	payments.Post("/", h.CreatePayment)
	payments.Get("/:id", h.GetPayment)
	payments.Put("/:id", h.UpdatePayment)
	payments.Delete("/:id", h.DeletePayment)

	// Loads routes
	loads := protected.Group("loads")
	loads.Get("/", h.GetLoads)
	loads.Post("/", h.CreateLoad)
	loads.Get("/:id", h.GetLoad)
	loads.Put("/:id", h.UpdateLoad)
	loads.Delete("/:id", h.DeleteLoad)
	loads.Put("/:id/status", h.UpdateLoadStatus)

	// Export routes (только для admin)
	exports := protected.Group("export", authMiddleware.RequireRole("admin"))
	exports.Post("/invoices", h.ExportInvoices)
	exports.Post("/payments", h.ExportPayments)
	exports.Post("/brokers", h.ExportBrokers)

	// Administrative routes (только для admin)
	admin := protected.Group("admin", authMiddleware.RequireRole("admin"))
	admin.Post("/send-overdue-notifications", h.SendOverdueNotifications)
}
