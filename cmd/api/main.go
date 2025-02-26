// cmd/api/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"github.com/dangerclosesec/supra/internal/auth"
	"github.com/dangerclosesec/supra/internal/config"
	"github.com/dangerclosesec/supra/internal/email"
	"github.com/dangerclosesec/supra/internal/handler"
	"github.com/dangerclosesec/supra/internal/middleware"
	"github.com/dangerclosesec/supra/internal/repository"
	"github.com/dangerclosesec/supra/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "startup error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := setupDatabase(cfg)
	if err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	factorRepo := repository.NewUserFactorRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Initialize auth services
	passwordHasher := auth.NewPasswordHasher()
	tokenManager := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.ExpiryPeriod)

	// Initialize email service
	emailService, err := email.NewEmailService(cfg, email.ProviderSendgrid)
	if err != nil {
		log.Fatalf("Failed to initialize email service: %v", err)
	}

	// Initialize cache service
	config := service.CacheConfig{
		TTL:         5 * time.Minute,
		CleanupFreq: 1 * time.Minute,
	}
	cacheService := service.NewCacheService(config)
	defer cacheService.Close()

	// Initialize factor service
	userFactorService := service.NewUserFactorService(factorRepo)

	// Initialize user service
	userService := service.NewUserService(
		userRepo,
		factorRepo,
		orgRepo,
		passwordHasher,
		tokenManager,
		emailService,
		userFactorService,
		cacheService,
		cfg,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userService, cacheService)
	userFactorHandler := handler.NewUserFactorHandler(userFactorService)

	// Create router
	r := chi.NewRouter()

	// Basic middleware stack
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(loggingMiddleware(logger))
	r.Use(recoveryMiddleware(logger))
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Get("/signup/verify", authHandler.VerifyHandler)

			r.Group(func(r chi.Router) {
				r.Use(chimw.AllowContentType("application/json"))

				// Auth routes
				r.Get("/signup", authHandler.SignupHandler)
				r.Post("/signup", authHandler.SignupHandler)
				r.Post("/login", authHandler.LoginHandler)
				// r.Post("/verify/resend", authHandler.ResendVerificationHandler)
			})

		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(chimw.AllowContentType("application/json"))
			r.Use(middleware.AuthMiddleware(tokenManager))

			// User factor routes
			r.Route("/factors", func(r chi.Router) {
				r.Get("/", userFactorHandler.ListFactors)
				r.Post("/", userFactorHandler.CreateFactor)
				r.Post("/{id}/verify", userFactorHandler.VerifyFactor)
				r.Delete("/{id}", userFactorHandler.RemoveFactor)
			})
		})
	})

	// Create server
	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           r,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Server error channel
	serverErrors := make(chan error, 1)

	// Start server
	go func() {
		logger.Info("server starting", "port", cfg.Server.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	// Shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	// Wait for shutdown or error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Info("shutdown started", "signal", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			// If shutdown times out, forcefully close
			srv.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func setupDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
		cfg.Database.SearchPath,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.Info("request completed",
					"method", r.Method,
					"path", r.URL.Path,
					"duration", time.Since(start),
					"status", ww.Status(),
					"size", ww.BytesWritten(),
					"requestID", chimw.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func recoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					err := errors.New("panic recovered")
					logger.Error("panic recovered",
						"error", err,
						"panic", rvr,
						"stack", string(debug.Stack()),
						"requestID", chimw.GetReqID(r.Context()),
					)

					// http.Error(w, map[string], http.StatusInternalServerError)
					w.Write([]byte("{\"error\":\"error encountered\"}"))
					return
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
