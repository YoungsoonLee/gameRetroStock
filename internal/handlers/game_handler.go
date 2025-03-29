package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/YoungsoonLee/gameRetroStock/internal/services"
	"github.com/gorilla/mux"
)

type GameHandler struct {
	igdbService  *services.IGDBService
	priceService *services.PriceService
}

func NewGameHandler(igdbService *services.IGDBService, priceService *services.PriceService) *GameHandler {
	return &GameHandler{
		igdbService:  igdbService,
		priceService: priceService,
	}
}

// GetIndex handles the main page request
func (h *GameHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/templates/index.html")
}

// GetGames handles requests for getting a list of games
func (h *GameHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 100

	games, total, err := h.igdbService.GetTopGames(page, limit)
	if err != nil {
		http.Error(w, "Failed to fetch games", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": games,
		"total": total,
	})
}

// SearchGames handles game search requests
func (h *GameHandler) SearchGames(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	platform := r.URL.Query().Get("platform")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 100

	games, total, err := h.igdbService.SearchGames(query, platform, page, limit)
	if err != nil {
		http.Error(w, "Failed to search games", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": games,
		"total": total,
	})
}

// GetGameDetails handles requests for getting detailed game information
func (h *GameHandler) GetGameDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	id, err := strconv.ParseInt(gameID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	// Get game details from IGDB
	game, err := h.igdbService.GetGameByID(id)
	if err != nil {
		http.Error(w, "Failed to fetch game details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

// GetGamePriceHistory handles requests for getting game price history
func (h *GameHandler) GetGamePriceHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	id, err := strconv.ParseInt(gameID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	// Get days parameter, default to 30 if not provided
	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// Get condition parameter
	condition := r.URL.Query().Get("condition")
	if condition == "" {
		condition = "all"
	}

	// Get price history
	history, err := h.priceService.GetPriceHistory(id, days)
	if err != nil {
		http.Error(w, "Failed to get price history", http.StatusInternalServerError)
		return
	}

	// Get price stats
	stats, err := h.priceService.GetPriceStats(id)
	if err != nil {
		http.Error(w, "Failed to get price stats", http.StatusInternalServerError)
		return
	}

	// Return combined response
	response := map[string]interface{}{
		"history": history,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetGameDetail handles the game detail page request
func (h *GameHandler) GetGameDetail(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/templates/game_detail.html")
}

func (h *GameHandler) GetGamePrices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameID := vars["id"]
	id, err := strconv.ParseInt(gameID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}
	condition := r.URL.Query().Get("condition")
	if condition == "" {
		condition = "all"
	}

	// Get price history from price service
	priceHistory, err := h.priceService.GetPriceHistory(id, days)
	if err != nil {
		http.Error(w, "Failed to fetch price history", http.StatusInternalServerError)
		return
	}

	// Price history should already include stats
	// No need to call GetPriceStats separately

	response := map[string]interface{}{
		"history": priceHistory.History,
		"stats":   priceHistory.Stats,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
