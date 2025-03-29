package services

import (
	"database/sql"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/YoungsoonLee/gameRetroStock/internal/models"
	"github.com/YoungsoonLee/gameRetroStock/internal/repository"
)

type PriceService struct {
	ebayService *EbayService
	repository  *repository.PriceRepository
	db          *sql.DB
}

func NewPriceService(ebayService *EbayService, repository *repository.PriceRepository, db *sql.DB) *PriceService {
	return &PriceService{
		ebayService: ebayService,
		repository:  repository,
		db:          db,
	}
}

// NewPriceServiceWithoutDB creates a PriceService that works without a database connection
func NewPriceServiceWithoutDB() *PriceService {
	return &PriceService{
		// No database connection or dependencies
	}
}

func (s *PriceService) UpdateGamePrices(gameID int64, gameName string, platform string) error {
	// Fetch prices from eBay
	prices, err := s.ebayService.SearchGamePrices(gameName, platform)
	if err != nil {
		log.Printf("Error fetching prices from eBay: %v", err)
		return err
	}

	// Set game ID for all prices
	for i := range prices {
		prices[i].GameID = gameID
	}

	// Save to database
	return s.repository.SavePrices(prices)
}

// GetPriceHistory returns the price history for a game
func (s *PriceService) GetPriceHistory(gameID int64, days int) (*models.PriceHistory, error) {
	// Handle invalid game IDs (0 or negative)
	if gameID <= 0 {
		return &models.PriceHistory{
			History: []models.DailyPrice{},
			Stats: models.PriceStats{
				Lowest:  0,
				Highest: 0,
				Average: 0,
				Current: 0,
				Trend:   0,
			},
		}, nil
	}

	// Set a seed based on the gameID to ensure consistent results for the same game
	// Use a fixed salt to help with debugging
	seed := time.Now().UnixNano()%100000 + (gameID * 1000)
	rand.Seed(seed)

	// Generate mock daily prices
	dailyPrices := make([]models.DailyPrice, days)
	basePrice := 1 + (float64(gameID%10) * 5.0) // Vary base price by game ID

	// Start from 'days' ago and go to today
	now := time.Now()

	for i := 0; i < days; i++ {
		dayOffset := days - i - 1
		date := now.AddDate(0, 0, -dayOffset)

		// Generate random price variations based on the day
		noiseScale := 0.05 // 5% noise
		noise := (rand.Float64()*2.0 - 1.0) * basePrice * noiseScale

		// Add some trend
		trend := float64(i) / float64(days) * 2.0 // +/- 2% total trend
		if gameID%2 == 0 {
			trend = -trend // Half the games trend down
		}

		price := basePrice*(1.0+(trend/100.0)) + noise
		price = math.Round(price*100) / 100 // Round to 2 decimal places

		dailyPrices[i] = models.DailyPrice{
			Date:  date,
			Price: price,
		}
	}

	// Calculate stats
	var lowest, highest, total float64
	if len(dailyPrices) > 0 {
		highest = dailyPrices[0].Price
		lowest = dailyPrices[0].Price
		total = dailyPrices[0].Price

		for i := 1; i < len(dailyPrices); i++ {
			price := dailyPrices[i].Price
			if price < lowest {
				lowest = price
			}
			if price > highest {
				highest = price
			}
			total += price
		}
	} else {
		// Return empty stats if no data
		return &models.PriceHistory{
			History: []models.DailyPrice{},
			Stats: models.PriceStats{
				Lowest:  0,
				Highest: 0,
				Average: 0,
				Current: 0,
				Trend:   0,
			},
		}, nil
	}

	average := total / float64(len(dailyPrices))
	current := dailyPrices[len(dailyPrices)-1].Price

	// Calculate trend (percentage change over the period)
	oldestPrice := dailyPrices[0].Price
	newestPrice := dailyPrices[len(dailyPrices)-1].Price
	trend := ((newestPrice - oldestPrice) / oldestPrice) * 100
	trend = math.Round(trend*10) / 10 // Round to 1 decimal place

	// Log for debugging
	log.Printf("Generated price history for game %d: %d items, current: %.2f, trend: %.1f%%",
		gameID, len(dailyPrices), current, trend)

	return &models.PriceHistory{
		History: dailyPrices,
		Stats: models.PriceStats{
			Lowest:  math.Round(lowest*100) / 100,
			Highest: math.Round(highest*100) / 100,
			Average: math.Round(average*100) / 100,
			Current: math.Round(current*100) / 100,
			Trend:   trend,
		},
	}, nil
}

// StartPriceUpdateScheduler starts a goroutine that periodically updates prices
func (s *PriceService) StartPriceUpdateScheduler() {
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			// TODO: Implement fetching all games and updating their prices
			log.Println("Running scheduled price update...")
		}
	}()
}

// GetPriceHistoryOld returns price history for a game within the specified time range
func (s *PriceService) GetPriceHistoryOld(gameID int64, days int) ([]models.DailyPrice, error) {
	// TODO: Implement database query
	// For now, return mock data
	basePrice := 59.99
	mockData := make([]models.DailyPrice, 0)

	// Calculate the start date based on the requested days
	now := time.Now()
	startDate := now.AddDate(0, 0, -days)

	// Generate daily price points
	currentDate := startDate
	for currentDate.Before(now) || currentDate.Equal(now) {
		// Add some random variation to the price (-5% to +5%)
		variation := (math.Sin(float64(currentDate.Day())) * 5.0)
		currentPrice := basePrice * (1 + variation/100)
		currentPrice = math.Round(currentPrice*100) / 100 // Round to 2 decimal places

		pricePoint := models.DailyPrice{
			Date:  currentDate,
			Price: currentPrice,
		}

		mockData = append(mockData, pricePoint)
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Sort data by date in ascending order
	sort.Slice(mockData, func(i, j int) bool {
		return mockData[i].Date.Before(mockData[j].Date)
	})

	return mockData, nil
}

// GetPriceStats returns price statistics for a game
func (s *PriceService) GetPriceStats(gameID int64) (*models.PriceStats, error) {
	priceHistory, err := s.GetPriceHistory(gameID, 30)
	if err != nil {
		return nil, err
	}
	return &priceHistory.Stats, nil
}
