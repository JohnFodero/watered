package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"watered/internal/auth"
	"watered/internal/handlers"
	"watered/internal/monitoring"
	"watered/internal/services"
	"watered/internal/storage"
)

func main() {
	// Load environment variables from .env files
	loadEnvFiles()

	// Initialize storage
	store := storage.NewMemoryStorage()
	defer store.Close()

	// Initialize services
	authService := auth.NewAuthService(store)
	plantService := services.NewPlantService(store)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService)
	plantHandlers := handlers.NewPlantHandlers(plantService, authService)
	adminHandlers := handlers.NewAdminHandler(store)

	// Initialize health monitoring
	healthMonitor := monitoring.NewHealthMonitor("1.0.0")
	healthMonitor.RegisterChecker(monitoring.NewDatabaseHealthChecker(store))
	healthMonitor.RegisterChecker(monitoring.NewMemoryHealthChecker(512.0)) // 512MB limit
	healthMonitor.RegisterChecker(monitoring.NewApplicationHealthChecker(store))

	// Parse templates
	templates, err := template.ParseGlob(filepath.Join("web", "templates", "*.html"))
	if err != nil {
		log.Printf("Warning: Could not parse templates: %v", err)
		templates = template.New("empty")
	}

	// Create router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// Health check endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"watered"}`))
	})

	// Comprehensive health monitoring endpoint
	r.Get("/health/detailed", healthMonitor.HTTPHandler())

	// Authentication routes
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", authHandlers.LoginHandler)
		r.Get("/callback", authHandlers.CallbackHandler)
		r.Post("/logout", authHandlers.LogoutHandler)
		r.Get("/status", authHandlers.StatusHandler)
		// Demo routes (only available in demo mode)
		r.HandleFunc("/demo-login", authHandlers.DemoLoginHandler)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/status", handlers.GetStatus)

		// Plant API routes
		r.Route("/plant", func(r chi.Router) {
			// Public plant endpoints (read-only)
			r.Get("/", plantHandlers.GetPlantHandler)
			r.Get("/status", plantHandlers.GetPlantStatusHandler)
			r.Get("/timer", plantHandlers.GetPlantTimerHandler)

			// Protected plant endpoints (require authentication)
			r.Group(func(r chi.Router) {
				r.Use(authService.AuthRequired)
				r.Post("/water", plantHandlers.WaterPlantHandler)
			})

			// Admin-only plant endpoints
			r.Group(func(r chi.Router) {
				r.Use(authService.AdminRequired)
				r.Put("/settings", plantHandlers.UpdatePlantSettingsHandler)
				r.Post("/reset", plantHandlers.ResetPlantHandler)
			})
		})
	})

	// Admin API routes
	r.Route("/admin", func(r chi.Router) {
		r.Use(authService.AdminRequired)

		// Configuration endpoints
		r.Get("/config", adminHandlers.GetConfigHandler)
		r.Put("/config/timeout", adminHandlers.UpdateTimeoutHandler)

		// User management endpoints
		r.Get("/users", adminHandlers.GetUsersHandler)
		r.Post("/users", adminHandlers.AddUserHandler)
		r.Delete("/users/{email}", adminHandlers.RemoveUserHandler)

		// History and statistics endpoints
		r.Get("/history", adminHandlers.GetHistoryHandler)
		r.Get("/stats", adminHandlers.GetStatsHandler)
	})

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Frontend routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// Check authentication and pass user data to template
		user, _ := authService.GetCurrentUser(r)
		templateData := map[string]interface{}{
			"User":          user,
			"Authenticated": user != nil,
		}

		if err := templates.ExecuteTemplate(w, "index.html", templateData); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			log.Printf("Template error: %v", err)
		}
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		// Redirect if already authenticated
		if authService.IsAuthenticated(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Check if demo mode is enabled (when GOOGLE_CLIENT_ID is not set or is demo value)
		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		demoMode := clientID == "" || clientID == "demo-client-id"

		templateData := map[string]interface{}{
			"DemoMode": demoMode,
		}

		if err := templates.ExecuteTemplate(w, "login.html", templateData); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			log.Printf("Template error: %v", err)
		}
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authService.AdminRequired)
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			user, _ := authService.GetCurrentUser(r)
			templateData := map[string]interface{}{
				"User": user,
			}

			if err := templates.ExecuteTemplate(w, "admin.html", templateData); err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
				log.Printf("Template error: %v", err)
			}
		})
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// loadEnvFiles loads environment variables from .env files in order of precedence
func loadEnvFiles() {
	// Environment files to load in order of precedence (last wins)
	envFiles := []string{
		".env.example", // Template with defaults (lowest priority)
		".env",         // Main environment file
		".env.local",   // Local overrides (highest priority)
	}

	// Also check for environment-specific files
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		envFiles = append(envFiles, ".env."+env)
	}

	loadedFiles := []string{}

	for _, file := range envFiles {
		if err := godotenv.Load(file); err == nil {
			loadedFiles = append(loadedFiles, file)
		}
		// Silently ignore missing files - they're optional
	}

	if len(loadedFiles) > 0 {
		log.Printf("Loaded environment variables from: %v", loadedFiles)
	}

	// Log current configuration status (without sensitive values)
	logConfigurationStatus()
}

// logConfigurationStatus logs the current configuration status
func logConfigurationStatus() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	sessionSecret := os.Getenv("SESSION_SECRET")
	allowedEmails := os.Getenv("ALLOWED_EMAILS")
	adminEmails := os.Getenv("ADMIN_EMAILS")
	environment := os.Getenv("ENVIRONMENT")

	log.Printf("Configuration Status:")
	log.Printf("  Environment: %s", getEnvOrDefault(environment, "development"))

	if clientID != "" && clientID != "demo-client-id" {
		log.Printf("  OAuth Mode: Production (Google OAuth enabled)")
		log.Printf("  Demo Login: Disabled")
	} else {
		log.Printf("  OAuth Mode: Demo (Google OAuth not configured)")
		log.Printf("  Demo Login: Available at /auth/demo-login")
	}

	if sessionSecret != "" && sessionSecret != "development-secret-change-in-production" {
		log.Printf("  Session Secret: Configured")
	} else {
		log.Printf("  Session Secret: Using development default")
	}

	if allowedEmails != "" {
		log.Printf("  Allowed Emails: Configured")
	} else {
		log.Printf("  Allowed Emails: Using demo defaults")
	}

	if adminEmails != "" {
		log.Printf("  Admin Emails: Configured")
	} else {
		log.Printf("  Admin Emails: Using demo defaults")
	}
}

// getEnvOrDefault returns the environment variable value or default if empty
func getEnvOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
