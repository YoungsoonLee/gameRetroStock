package models

import "time"

type Video struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Game struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Platform     string    `json:"platform"`
	ReleaseDate  time.Time `json:"release_date"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"image_url"`
	YoutubeURL   string    `json:"youtube_url,omitempty"`
	CurrentPrice float64   `json:"current_price,omitempty"`
	Rating       float64   `json:"rating"`
	Videos       []Video   `json:"videos"`
	Screenshots  []string  `json:"screenshots"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GamePrice struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"game_id"`
	Price     float64   `json:"price"`
	Date      time.Time `json:"date"`
	Condition string    `json:"condition"` // new, used, etc.
	Source    string    `json:"source"`    // pricecharting, etc.
	CreatedAt time.Time `json:"created_at"`
}

type Platform struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Company   string    `json:"company"`
	CreatedAt time.Time `json:"created_at"`
}

type GameVideo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
