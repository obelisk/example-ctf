package main

import (
	"context"
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/obelisk/example-ctf/config"
	"github.com/obelisk/example-ctf/middleware"
	"github.com/obelisk/example-ctf/routes"
	"github.com/obelisk/example-ctf/services"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	log.SetLevel(log.InfoLevel)

	// Load configuration
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Build database connection string from config
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Hostname,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.SslMode,
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create dependency container
	container := services.NewContainer(db, &cfg)

	// Start Slack leaderboard updates if configured
	if container.SlackService != nil && cfg.Slack.LeaderboardInterval > 0 {
		container.SlackService.StartLeaderboardUpdates(context.Background())
		log.Printf("Slack leaderboard updates started with interval: %v", cfg.Slack.LeaderboardInterval)
	}

	// Create health check router (separate port)
	healthRouter := mux.NewRouter()
	healthRouter.HandleFunc("/health", routes.HealthCheckHandler(container)).Methods("GET")

	// Start health check server in a goroutine
	healthServerAddr := fmt.Sprintf("%s:%d", cfg.HealthCheck.Hostname, cfg.HealthCheck.Port)
	go func() {
		log.Printf("Health check server listening on %s", healthServerAddr)
		if err := http.ListenAndServe(healthServerAddr, healthRouter); err != nil {
			log.Fatalf("Health check server failed: %v", err)
		}
	}()

	// Create main router
	r := mux.NewRouter()

	// Add global middleware
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RateLimitMiddleware(container))
	r.Use(middleware.RequestSizeLimitMiddleware(container))
	r.Use(middleware.LoadAuthenticatedUser(container))
	r.Use(middleware.SecurityHeadersMiddleware)

	// Serve static files from frontend-simple directory
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./frontend-simple/"))))

	// Authenticated routes with user-based rate limiting
	authR := r.PathPrefix("/api").Subrouter()
	authR.Use(middleware.RequireAuthenticated(container))
	authR.Use(middleware.UserRateLimitMiddleware(container))

	authR.HandleFunc("/profile", routes.GetProfile(container)).Methods("GET")
	authR.HandleFunc("/alias", routes.SetAlias(container)).Methods("POST")
	authR.HandleFunc("/alias", routes.RemoveAlias(container)).Methods("DELETE")

	authR.HandleFunc("/challenges", routes.ListChallenges(container)).Methods("GET")
	authR.HandleFunc("/challenges/{id}", routes.GetChallenge(container)).Methods("GET")
	authR.HandleFunc("/challenges/{id}/submission", routes.SubmitChallenge(container)).Methods("POST")

	authR.HandleFunc("/adoble", routes.ListExamChallenges(container)).Methods("GET")
	authR.HandleFunc("/adoble/{id}", routes.GetExamChallenge(container)).Methods("GET")
	authR.HandleFunc("/adoble/{id}/submission", routes.SubmitExamChallenge(container)).Methods("POST")

	// Serve index.html for all other routes (SPA fallback)
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend-simple/index.html")
	})

	// Start main server with timeout configuration
	serverAddr := fmt.Sprintf("%s:%d", cfg.HTTP.Hostname, cfg.HTTP.Port)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  cfg.HTTP.Timeout,
		WriteTimeout: cfg.HTTP.Timeout,
		IdleTimeout:  cfg.HTTP.Timeout,
	}

	log.Printf("Main server listening on %s", serverAddr)
	log.Fatal(server.ListenAndServe())
}
