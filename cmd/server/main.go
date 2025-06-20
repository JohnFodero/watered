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

	"watered/internal/auth"
	"watered/internal/handlers"
	"watered/internal/storage"
)

func main() {
	// Initialize storage
	store := storage.NewMemoryStorage()
	defer store.Close()

	// Initialize auth service
	authService := auth.NewAuthService(store)
	authHandlers := handlers.NewAuthHandlers(authService)

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

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"watered"}`))
	})

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
		// Protected API routes
		r.Group(func(r chi.Router) {
			r.Use(authService.AuthRequired)
			// Future plant management endpoints will go here
		})
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
		
		if err := templates.ExecuteTemplate(w, "login.html", nil); err != nil {
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