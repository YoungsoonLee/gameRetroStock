package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/YoungsoonLee/gameRetroStock/internal/config"
	"github.com/YoungsoonLee/gameRetroStock/internal/models"
	"github.com/YoungsoonLee/gameRetroStock/internal/repository"
	"github.com/YoungsoonLee/gameRetroStock/internal/services"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Initialize variables
	var db *sql.DB
	var priceRepo *repository.PriceRepository
	var ebayService *services.EbayService
	var priceService *services.PriceService

	// Try to initialize database
	db, err = sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Fatal: Error connecting to database:", err)
	} else {
		defer db.Close()

		log.Println("Database URL:", cfg.DatabaseURL)

		// Test connection
		if err = db.Ping(); err != nil {
			log.Fatal("Fatal: Error pinging database:", err)
		} else {
			log.Println("Successfully connected to database")

			// Initialize repositories
			priceRepo = repository.NewPriceRepository(db)

			// Initialize services that need database
			ebayService = services.NewEbayService(cfg)
			priceService = services.NewPriceService(ebayService, priceRepo, db)
		}
	}

	// Initialize services that don't require database
	if ebayService == nil {
		ebayService = services.NewEbayService(cfg)
	}

	if priceService == nil {
		// Create a mock price service without database
		priceService = services.NewPriceServiceWithoutDB()
	}

	// Initialize services
	serviceContainer := &services.Services{
		IGDB:  services.NewIGDBService(cfg),
		Price: priceService,
	}

	// Initialize router
	r := mux.NewRouter()
	setupRoutes(r, serviceContainer)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	log.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func setupRoutes(r *mux.Router, services *services.Services) {
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// API routes
	r.HandleFunc("/api/games/top", func(w http.ResponseWriter, r *http.Request) {
		page := 1
		limit := 100

		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		games, total, err := services.IGDB.GetTopGames(page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": games,
			"total": total,
		})
	}).Methods("GET")

	r.HandleFunc("/api/games/search", func(w http.ResponseWriter, r *http.Request) {
		page := 1
		limit := 100
		query := r.URL.Query().Get("q")
		platform := r.URL.Query().Get("platform")

		if pageStr := r.URL.Query().Get("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		games, total, err := services.IGDB.SearchGames(query, platform, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": games,
			"total": total,
		})
	}).Methods("GET")

	// Price history endpoint
	r.HandleFunc("/api/games/{id}/prices", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		gameIDStr := vars["id"]

		// Check if the ID is "undefined" literally
		if gameIDStr == "undefined" || gameIDStr == "" {
			http.Error(w, "Invalid game ID", http.StatusBadRequest)
			return
		}

		gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing game ID %s: %v", gameIDStr, err)
			http.Error(w, "Invalid game ID format", http.StatusBadRequest)
			return
		}

		days := 30 // Default to 30 days
		if daysStr := r.URL.Query().Get("days"); daysStr != "" {
			if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
				days = d
			}
		}

		history, err := services.Price.GetPriceHistory(gameID, days)
		if err != nil {
			log.Printf("Error fetching price history for game %d: %v", gameID, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Ensure we're returning valid JSON even when history is empty
		if history == nil {
			history = &models.PriceHistory{
				History: []models.DailyPrice{},
				Stats: models.PriceStats{
					Lowest:  0,
					Highest: 0,
					Average: 0,
					Current: 0,
					Trend:   0,
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if err := json.NewEncoder(w).Encode(history); err != nil {
			log.Printf("Error encoding price history response for game %d: %v", gameID, err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	// Game detail API endpoint
	r.HandleFunc("/api/games/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		gameIDStr := vars["id"]

		// Check if the ID is "undefined" literally
		if gameIDStr == "undefined" || gameIDStr == "" {
			http.Error(w, "Invalid game ID", http.StatusBadRequest)
			return
		}

		gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing game ID %s: %v", gameIDStr, err)
			http.Error(w, "Invalid game ID format", http.StatusBadRequest)
			return
		}

		game, err := services.IGDB.GetGameByID(gameID)
		if err != nil {
			log.Printf("Error fetching game details for game %d: %v", gameID, err)
			http.Error(w, "Failed to fetch game details", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(game); err != nil {
			log.Printf("Error encoding game details response for game %d: %v", gameID, err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	// Game detail routes
	r.HandleFunc("/games/{id}", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("static/templates/game_detail.html"))
		tmpl.Execute(w, nil)
	}).Methods("GET")

	// 플랫폼 목록 API 엔드포인트
	r.HandleFunc("/api/platforms", func(w http.ResponseWriter, r *http.Request) {
		platforms, err := services.IGDB.GetPlatformsList()
		if err != nil {
			log.Printf("Error fetching platforms: %v", err)
			http.Error(w, "Failed to fetch platforms", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(platforms); err != nil {
			log.Printf("Error encoding platforms response: %v", err)
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	}).Methods("GET")

	// Main page route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("static/templates/index.html"))
		tmpl.Execute(w, nil)
	})
}
