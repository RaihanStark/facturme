// Package main provides the entry point for the Worklio API server.
// It initializes the database connection, sets up the Echo web server,
// and configures routes for authentication endpoints.
//
// @title Worklio API
// @version 1.0
// @description A basic authentication API built with Go, Echo v4, sqlc, and PostgreSQL
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@worklio.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"database/sql"
	"log"
	"time"
	"worklio-api/internal/db"
	"worklio-api/internal/email"
	"worklio-api/internal/handlers"
	appMiddleware "worklio-api/internal/middleware"
	"worklio-api/internal/services"
	"worklio-api/pkg/config"

	_ "worklio-api/docs"

	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	database, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Test database connection
	if err := database.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize queries
	queries := db.New(database)

	// Initialize email service
	var emailService *email.Service
	if cfg.SMTPHost != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != "" {
		var err error
		emailService, err = email.NewService(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SenderEmail,
			cfg.SenderName,
			cfg.AppURL,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize email service: %v", err)
			log.Println("Email sending will be disabled. Verification tokens will be logged to console.")
		} else {
			log.Println("Email service initialized successfully with SMTP")
		}
	} else {
		log.Println("SMTP credentials not configured. Email sending disabled. Verification tokens will be logged to console.")
	}

	// Initialize exchange rate service
	exchangeRateService := services.NewExchangeRateService(queries)

	// Initialize and start cron scheduler for exchange rates
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal("Failed to create scheduler:", err)
	}

	// Schedule exchange rate updates daily at 2 AM
	_, err = scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(2, 0, 0))),
		gocron.NewTask(func() {
			log.Println("Starting scheduled exchange rate update...")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			if err := exchangeRateService.UpdateAllRates(ctx); err != nil {
				log.Printf("Error updating exchange rates: %v", err)
			} else {
				log.Println("Exchange rates updated successfully")
			}
		}),
	)
	if err != nil {
		log.Fatal("Failed to schedule exchange rate job:", err)
	}

	// Start the scheduler
	scheduler.Start()
	log.Println("Exchange rate scheduler started (runs daily at 2 AM)")

	// Run initial update on startup
	go func() {
		log.Println("Running initial exchange rate update...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := exchangeRateService.UpdateAllRates(ctx); err != nil {
			log.Printf("Warning: Initial exchange rate update failed: %v", err)
		} else {
			log.Println("Initial exchange rates loaded successfully")
		}
	}()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(queries, cfg.JWTSecret, emailService)
	clientHandler := handlers.NewClientHandler(queries)
	timeEntryHandler := handlers.NewTimeEntryHandler(queries, exchangeRateService)
	invoiceHandler := handlers.NewInvoiceHandler(queries)
	demoHandler := handlers.NewDemoHandler(queries)
	currencyHandler := handlers.NewCurrencyHandler(exchangeRateService)
	statsHandler := handlers.NewStatsHandler(queries, exchangeRateService)

	// Routes
	api := e.Group("/api")

	// Public routes
	api.GET("/supported-currencies", currencyHandler.GetSupportedCurrencies)
	api.GET("/convert-currency", currencyHandler.ConvertCurrency)

	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(appMiddleware.JWTAuth(cfg.JWTSecret))
	{
		// User routes
		protected.GET("/users/me", authHandler.GetCurrentUser)
		protected.POST("/users/complete-onboarding", authHandler.CompleteOnboarding)
		protected.POST("/users/complete-tour", authHandler.CompleteTour)
		protected.POST("/users/change-password", authHandler.ChangePassword)
		protected.POST("/users/currency", authHandler.UpdateCurrency)

		// Auth routes (protected)
		protected.POST("/auth/resend-verification", authHandler.ResendVerificationEmail)

		// Client routes
		protected.POST("/clients", clientHandler.CreateClient)
		protected.GET("/clients", clientHandler.GetClients)
		protected.GET("/clients/:id", clientHandler.GetClient)
		protected.PUT("/clients/:id", clientHandler.UpdateClient)
		protected.DELETE("/clients/:id", clientHandler.DeleteClient)

		// Time entry routes
		protected.POST("/time-entries", timeEntryHandler.CreateTimeEntry)
		protected.GET("/time-entries", timeEntryHandler.GetTimeEntries)
		protected.GET("/time-entries/stats", timeEntryHandler.GetTimeEntriesStats)
		protected.GET("/time-entries/heatmap", timeEntryHandler.GetHeatmap)
		protected.GET("/time-entries/:id", timeEntryHandler.GetTimeEntry)
		protected.PUT("/time-entries/:id", timeEntryHandler.UpdateTimeEntry)
		protected.DELETE("/time-entries/:id", timeEntryHandler.DeleteTimeEntry)

		// Invoice routes
		protected.POST("/invoices", invoiceHandler.CreateInvoice)
		protected.GET("/invoices", invoiceHandler.GetInvoices)
		protected.GET("/invoices/available-time-entries", invoiceHandler.GetAvailableTimeEntries)
		protected.GET("/invoices/:id", invoiceHandler.GetInvoice)
		protected.GET("/invoices/:id/pdf", invoiceHandler.DownloadInvoicePDF)
		protected.PUT("/invoices/:id", invoiceHandler.UpdateInvoice)
		protected.PATCH("/invoices/:id/status", invoiceHandler.UpdateInvoiceStatus)
		protected.DELETE("/invoices/:id", invoiceHandler.DeleteInvoice)

		// Demo routes
		protected.POST("/demo/generate", demoHandler.GenerateDemoData)
		protected.DELETE("/demo", demoHandler.DeleteDemoData)

		// Stats routes
		protected.GET("/stats/dashboard", statsHandler.GetDashboardStats)
		protected.GET("/stats/recent-time-entries", statsHandler.GetRecentTimeEntries)
		protected.GET("/stats/recent-invoices", statsHandler.GetRecentInvoices)
		protected.GET("/stats/invoices", statsHandler.GetInvoiceStats)
	}

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
