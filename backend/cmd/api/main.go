package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/savegress/platform/backend/internal/config"
	"github.com/savegress/platform/backend/internal/handlers"
	appMiddleware "github.com/savegress/platform/backend/internal/middleware"
	"github.com/savegress/platform/backend/internal/repository"
	"github.com/savegress/platform/backend/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redis, err := repository.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Initialize services
	authService := services.NewAuthService(db, redis, cfg.JWTSecret)
	licenseService := services.NewLicenseService(db, cfg.LicensePrivateKey)
	billingService := services.NewBillingService(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	billingService.SetDB(db)
	userService := services.NewUserService(db)
	telemetryService := services.NewTelemetryService(db, redis)
	earlyAccessService := services.NewEarlyAccessService(db, cfg.AdminEmail, cfg.ResendAPIKey)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	licenseHandler := handlers.NewLicenseHandler(licenseService, authService)
	billingHandler := handlers.NewBillingHandler(billingService, licenseService, userService)
	userHandler := handlers.NewUserHandler(userService)
	telemetryHandler := handlers.NewTelemetryHandler(telemetryService, licenseService)
	healthHandler := handlers.NewHealthHandler(db, redis)
	earlyAccessHandler := handlers.NewEarlyAccessHandler(earlyAccessService, cfg.TurnstileSecretKey)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting
	r.Use(httprate.LimitByIP(100, time.Minute))

	// Health check (public)
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)
	r.Get("/health/detailed", healthHandler.Detailed)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
		})

		// License validation (used by CDC engines)
		r.Route("/license", func(r chi.Router) {
			r.Post("/validate", licenseHandler.Validate)
			r.Post("/activate", licenseHandler.Activate)
			r.Post("/deactivate", licenseHandler.Deactivate)
		})

		// Stripe webhooks (public but verified)
		r.Post("/webhooks/stripe", billingHandler.HandleWebhook)

		// Telemetry (from CDC engines)
		r.Post("/telemetry", telemetryHandler.Receive)

		// Early access form (landing page)
		r.Post("/early-access", earlyAccessHandler.Submit)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.Auth(authService))

			// User profile
			r.Route("/user", func(r chi.Router) {
				r.Get("/", userHandler.GetProfile)
				r.Put("/", userHandler.UpdateProfile)
				r.Put("/password", userHandler.ChangePassword)
			})

			// Licenses management
			r.Route("/licenses", func(r chi.Router) {
				r.Get("/", licenseHandler.List)
				r.Post("/", licenseHandler.Create)
				r.Get("/{id}", licenseHandler.Get)
				r.Delete("/{id}", licenseHandler.Revoke)
				r.Get("/{id}/activations", licenseHandler.GetActivations)
			})

			// Billing
			r.Route("/billing", func(r chi.Router) {
				r.Get("/subscription", billingHandler.GetSubscription)
				r.Post("/subscription", billingHandler.CreateSubscription)
				r.Put("/subscription", billingHandler.UpdateSubscription)
				r.Delete("/subscription", billingHandler.CancelSubscription)
				r.Get("/invoices", billingHandler.ListInvoices)
				r.Get("/payment-methods", billingHandler.ListPaymentMethods)
				r.Post("/payment-methods", billingHandler.AddPaymentMethod)
				r.Delete("/payment-methods/{id}", billingHandler.RemovePaymentMethod)
				r.Post("/portal-session", billingHandler.CreatePortalSession)
			})

			// Dashboard / Analytics
			r.Route("/dashboard", func(r chi.Router) {
				r.Get("/stats", telemetryHandler.GetStats)
				r.Get("/usage", telemetryHandler.GetUsage)
				r.Get("/instances", telemetryHandler.GetInstances)
			})

			// Downloads
			r.Route("/downloads", func(r chi.Router) {
				r.Get("/", handlers.ListDownloads)
				r.Get("/{product}/{version}", handlers.GetDownloadURL)
			})
		})

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(appMiddleware.Auth(authService))
			r.Use(appMiddleware.RequireAdmin)

			r.Get("/users", userHandler.ListUsers)
			r.Get("/users/{id}", userHandler.GetUser)
			r.Put("/users/{id}", userHandler.UpdateUser)
			r.Get("/licenses", licenseHandler.ListAll)
			r.Post("/licenses/generate", licenseHandler.AdminGenerate)
		})
	})

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
