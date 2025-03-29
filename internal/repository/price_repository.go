package repository

import (
	"database/sql"
	"log"

	"github.com/YoungsoonLee/gameRetroStock/internal/models"
)

type PriceRepository struct {
	db *sql.DB
}

func NewPriceRepository(db *sql.DB) *PriceRepository {
	return &PriceRepository{
		db: db,
	}
}

// CreatePriceTable creates the price table if it doesn't exist
func (r *PriceRepository) CreatePriceTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS prices (
        id SERIAL PRIMARY KEY,
        game_id BIGINT NOT NULL,
        price DECIMAL(10, 2) NOT NULL,
        condition VARCHAR(50),
        source VARCHAR(50),
        listing_url TEXT,
        title TEXT,
        collected_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
    `

	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("Error creating price table: %v", err)
		return err
	}

	return nil
}

// SavePrices saves a list of price data to the database
func (r *PriceRepository) SavePrices(prices []models.PriceData) error {
	// For now, just log the prices
	log.Printf("Saving %d prices to database", len(prices))
	return nil
}

// GetPriceHistory returns price history for a game from the database
func (r *PriceRepository) GetPriceHistory(gameID int64, days int) ([]models.DailyPrice, error) {
	// For simplicity, we'll return nil as the actual implementation would
	// be handled by the PriceService with mock data for now
	return nil, nil
}

// GetPriceStats returns price statistics for a game from the database
func (r *PriceRepository) GetPriceStats(gameID int64, days int) (models.PriceStats, error) {
	// For simplicity, we'll return an empty struct
	return models.PriceStats{}, nil
}
