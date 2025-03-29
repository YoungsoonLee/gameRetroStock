package services

import (
	"math/rand"
	"time"

	"github.com/YoungsoonLee/gameRetroStock/internal/config"
	"github.com/YoungsoonLee/gameRetroStock/internal/models"
)

type EbayService struct {
	config *config.Config
}

func NewEbayService(config *config.Config) *EbayService {
	return &EbayService{
		config: config,
	}
}

// SearchGamePrices returns a list of price data for a game from eBay
func (s *EbayService) SearchGamePrices(gameName string, platform string) ([]models.PriceData, error) {
	// For now, return mock data
	rand.Seed(time.Now().UnixNano())

	mockPrices := make([]models.PriceData, 10)
	basePrice := 1.0

	for i := 0; i < 10; i++ {
		// Generate random price variations
		variation := (rand.Float64() * 10) - 5 // Random value between -5 and 5
		price := basePrice + variation

		condition := "Used"
		if rand.Intn(5) == 0 {
			condition = "New"
			price += 10.0 // New items are more expensive
		}

		mockPrices[i] = models.PriceData{
			ID:          int64(i + 1),
			Price:       price,
			Condition:   condition,
			Source:      "eBay",
			ListingURL:  "https://www.ebay.com/itm/example",
			Title:       gameName + " for " + platform,
			CollectedAt: time.Now().AddDate(0, 0, -rand.Intn(30)),
		}
	}

	return mockPrices, nil
}

// GetCurrentPrice gets the current price for a game
func (s *EbayService) GetCurrentPrice(gameID int64) (float64, error) {
	// Mock implementation - in real life, this would query eBay API
	// Let the price vary by game ID for now
	basePrice := 1.0
	seed := time.Now().Add(time.Hour * time.Duration(gameID)).UnixNano()
	rand.Seed(seed)
	variance := rand.Float64() * 10.0

	return basePrice + variance, nil
}

// GetPriceHistory returns the price history for a game from eBay
func (s *EbayService) GetPriceHistory(gameID int64) (*models.PriceHistory, error) {
	// For now, return mock data
	dailyPrices := make([]models.DailyPrice, 30)
	basePrice := 1.0

	for i := 0; i < 30; i++ {
		// Generate random price variations
		variation := rand.Float64()*4 - 2 // Random value between -2 and 2
		price := basePrice + variation

		dailyPrices[i] = models.DailyPrice{
			Date:  time.Now().AddDate(0, 0, -i),
			Price: price,
		}
	}

	// Calculate stats
	var lowest, highest, total float64
	highest = dailyPrices[0].Price
	lowest = dailyPrices[0].Price
	total = 0

	for _, price := range dailyPrices {
		if price.Price < lowest {
			lowest = price.Price
		}
		if price.Price > highest {
			highest = price.Price
		}
		total += price.Price
	}

	average := total / float64(len(dailyPrices))
	current := dailyPrices[0].Price

	// Calculate trend (percentage change over the period)
	oldestPrice := dailyPrices[len(dailyPrices)-1].Price
	newestPrice := dailyPrices[0].Price
	trend := ((newestPrice - oldestPrice) / oldestPrice) * 100

	return &models.PriceHistory{
		History: dailyPrices,
		Stats: models.PriceStats{
			Lowest:  lowest,
			Highest: highest,
			Average: average,
			Current: current,
			Trend:   trend,
		},
	}, nil
}
