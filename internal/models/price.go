package models

import "time"

type PriceData struct {
	ID          int64     `json:"id"`
	GameID      int64     `json:"game_id"`
	Price       float64   `json:"price"`
	Condition   string    `json:"condition"`
	Source      string    `json:"source"`
	ListingURL  string    `json:"listing_url"`
	Title       string    `json:"title"`
	CollectedAt time.Time `json:"collected_at"`
}

type DailyPrice struct {
	Date  time.Time `json:"date"`
	Price float64   `json:"price"`
}

type PriceStats struct {
	Lowest  float64 `json:"lowest"`
	Highest float64 `json:"highest"`
	Average float64 `json:"average"`
	Current float64 `json:"current"`
	Trend   float64 `json:"trend"` // Percentage change
}

type PriceHistory struct {
	History []DailyPrice `json:"history"`
	Stats   PriceStats   `json:"stats"`
}
