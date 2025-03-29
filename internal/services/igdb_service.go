package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/YoungsoonLee/gameRetroStock/internal/config"
	"github.com/YoungsoonLee/gameRetroStock/internal/models"
)

type IGDBService struct {
	config         *config.Config
	client         *http.Client
	authToken      string
	tokenExpiry    time.Time
	platformMap    map[string]int64 // 플랫폼 이름 -> ID 매핑
	platformMapMux sync.RWMutex     // 플랫폼 맵 동시성 제어
}

func NewIGDBService(config *config.Config) *IGDBService {
	svc := &IGDBService{
		config:      config,
		client:      &http.Client{Timeout: 10 * time.Second},
		platformMap: make(map[string]int64),
	}

	// 서비스 시작 시 플랫폼 목록 초기화
	go svc.initPlatforms()

	return svc
}

// initPlatforms IGDB API에서 플랫폼 목록을 가져와 매핑 테이블 구성
func (s *IGDBService) initPlatforms() {
	// 토큰이 없으면 먼저 가져오기
	if err := s.ensureValidToken(); err != nil {
		log.Printf("Failed to initialize platforms: %v", err)
		return
	}

	// IGDB에서 모든 플랫폼 가져오기
	query := `
		fields id, name, abbreviation;
		limit 500;
	`

	platforms, err := s.getPlatforms(query)
	if err != nil {
		log.Printf("Failed to fetch platforms: %v", err)
		return
	}

	// 플랫폼 맵 구성
	s.platformMapMux.Lock()
	defer s.platformMapMux.Unlock()

	for _, p := range platforms {
		s.platformMap[strings.ToLower(p.Name)] = p.ID
		if p.Abbreviation != "" {
			s.platformMap[strings.ToLower(p.Abbreviation)] = p.ID
		}
	}

	log.Printf("Initialized %d platforms", len(platforms))

	for _, p := range platforms {
		log.Printf("Platform: %+v", p)
	}
}

// Platform IGDB의 플랫폼 구조체
type Platform struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

// getPlatforms IGDB에서 플랫폼 목록 가져오기
func (s *IGDBService) getPlatforms(query string) ([]Platform, error) {
	url := "https://api.igdb.com/v4/platforms"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", s.config.IGDBClientID)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error querying IGDB platforms (status %d): %s", resp.StatusCode, string(body))
	}

	var platforms []Platform
	if err := json.NewDecoder(resp.Body).Decode(&platforms); err != nil {
		return nil, err
	}

	return platforms, nil
}

// getPlatformID 플랫폼 이름으로 ID 찾기
func (s *IGDBService) getPlatformID(platformName string) (int64, bool) {
	if platformName == "" || platformName == "All Platforms" {
		return 0, false
	}

	s.platformMapMux.RLock()
	defer s.platformMapMux.RUnlock()

	// 정확히 일치하는 이름으로 검색
	if id, ok := s.platformMap[strings.ToLower(platformName)]; ok {
		return id, true
	}

	// 키워드 부분 일치 검색
	for name, id := range s.platformMap {
		if strings.Contains(name, strings.ToLower(platformName)) {
			return id, true
		}
	}

	return 0, false
}

// GetPlatformsList 사용자 인터페이스에 표시할 플랫폼 목록 반환
func (s *IGDBService) GetPlatformsList() ([]Platform, error) {
	// 주요 콘솔 플랫폼 정적 목록 반환
	platforms := []Platform{
		{ID: 167, Name: "PlayStation 5", Abbreviation: "PS5"},
		{ID: 48, Name: "PlayStation 4", Abbreviation: "PS4"},
		{ID: 9, Name: "PlayStation 3", Abbreviation: "PS3"},
		{ID: 8, Name: "PlayStation 2", Abbreviation: "PS2"},
		{ID: 7, Name: "PlayStation", Abbreviation: "PSX"},
		{ID: 38, Name: "PlayStation Portable", Abbreviation: "PSP"},
		{ID: 46, Name: "PlayStation Vita", Abbreviation: "PSV"},
		{ID: 169, Name: "Xbox Series X|S", Abbreviation: "XSX"},
		{ID: 49, Name: "Xbox One", Abbreviation: "XONE"},
		{ID: 12, Name: "Xbox 360", Abbreviation: "X360"},
		{ID: 11, Name: "Xbox", Abbreviation: "XBOX"},
		{ID: 130, Name: "Nintendo Switch", Abbreviation: "Switch"},
		{ID: 41, Name: "Wii U", Abbreviation: "WiiU"},
		{ID: 5, Name: "Wii", Abbreviation: "Wii"},
		{ID: 37, Name: "Nintendo 3DS", Abbreviation: "3DS"},
		{ID: 4, Name: "Nintendo 64", Abbreviation: "N64"},
		{ID: 21, Name: "Nintendo GameCube", Abbreviation: "GC"},
		{ID: 18, Name: "Nintendo Entertainment System", Abbreviation: "NES"},
		{ID: 19, Name: "Super Nintendo Entertainment System", Abbreviation: "SNES"},
		{ID: 32, Name: "Sega Saturn", Abbreviation: "Saturn"},
		{ID: 35, Name: "Sega Game Gear", Abbreviation: "GG"},
		{ID: 64, Name: "Sega Master System", Abbreviation: "SMS"},
		{ID: 29, Name: "Sega Mega Drive/Genesis", Abbreviation: "GEN"},
		{ID: 23, Name: "Sega Dreamcast", Abbreviation: "DC"},
		{ID: 22, Name: "Game Boy Advance", Abbreviation: "GBA"},
		{ID: 24, Name: "Game Boy Color", Abbreviation: "GBC"},
		{ID: 26, Name: "Game Boy", Abbreviation: "GB"},
		{ID: 34, Name: "Atari 7800", Abbreviation: "7800"},
		{ID: 60, Name: "Atari 2600", Abbreviation: "2600"},
		{ID: 62, Name: "Atari Jaguar", Abbreviation: "Jaguar"},
		{ID: 63, Name: "Atari ST/STE", Abbreviation: "Atari-ST"},
	}

	log.Printf("Returning %d static console platforms", len(platforms))
	return platforms, nil
}

// ensureValidToken ensures a valid Twitch API token is available
func (s *IGDBService) ensureValidToken() error {
	// Check if token is still valid
	if s.authToken != "" && time.Now().Before(s.tokenExpiry) {
		return nil
	}

	// Get new token
	url := "https://id.twitch.tv/oauth2/token"
	reqBody := fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials",
		s.config.IGDBClientID, s.config.IGDBClientSecret)

	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error getting auth token (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	s.authToken = tokenResp.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return nil
}

// GetTopGames returns the top-rated games
func (s *IGDBService) GetTopGames(page, limit int) ([]*models.Game, int, error) {
	// Fetch real data if token is available, otherwise use mock data
	if err := s.ensureValidToken(); err != nil {
		log.Printf("Warning: using mock data due to token error: %v", err)
		return s.getMockGames(page, limit)
	}

	offset := (page - 1) * limit

	// Get list of console platforms names to filter by
	consolePlatforms := []string{
		"PlayStation 5", "PlayStation 4", "PlayStation 3", "PlayStation 2", "PlayStation",
		"PlayStation Portable", "PlayStation Vita", "Xbox Series X|S", "Xbox One", "Xbox 360",
		"Xbox", "Nintendo Switch", "Wii U", "Wii", "Nintendo 3DS", "Nintendo 64",
		"Nintendo GameCube", "Nintendo Entertainment System", "Super Nintendo Entertainment System",
		"Sega Saturn", "Sega Game Gear", "Sega Master System", "Sega Mega Drive", "Genesis",
		"Sega Dreamcast", "Game Boy Advance", "Game Boy Color", "Game Boy",
		"Atari 7800", "Atari 2600", "Atari Jaguar", "Atari ST/STE",
	}

	// Build a search condition to match any console platform
	platformConditions := make([]string, len(consolePlatforms))
	for i, platformName := range consolePlatforms {
		platformConditions[i] = fmt.Sprintf(`platforms.name ~ *"%s"*`, platformName)
	}
	platformClause := strings.Join(platformConditions, " | ")

	// Construct query to get games with high ratings from console platforms only
	query := fmt.Sprintf(`
		fields name, summary, first_release_date, cover.url, rating, platforms.name;
		where rating > 75 & (%s);
		sort rating desc;
		limit %d;
		offset %d;
	`, platformClause, limit, offset)

	games, err := s.queryIGDB("games", query)
	if err != nil {
		log.Printf("Error querying IGDB for top games: %v", err)
		return s.getMockGames(page, limit)
	}

	// Get total count for pagination
	countQuery := fmt.Sprintf(`
		where rating > 75 & (%s);
	`, platformClause)

	count, err := s.getCount("games", countQuery)
	if err != nil {
		log.Printf("Error getting count: %v", err)
		count = 100 // Default to 100 if we can't get the real count
	}

	return games, count, nil
}

// SearchGames searches for games by name and platform
func (s *IGDBService) SearchGames(query, platform string, page, limit int) ([]*models.Game, int, error) {
	// Fetch real data if token is available, otherwise use mock data
	if err := s.ensureValidToken(); err != nil {
		log.Printf("Warning: using mock data due to token error: %v", err)
		return s.getMockGames(page, limit)
	}

	offset := (page - 1) * limit

	// Base query construction
	var searchQuery string

	// Case 1: Both query and platform specified
	if query != "" && platform != "" && platform != "All Platforms" {
		searchQuery = fmt.Sprintf(`
			fields name, summary, first_release_date, cover.url, rating, platforms.name, platforms.id;
			search "%s";
			where platforms.name ~ *"%s"*;
			limit %d;
			offset %d;
		`, query, platform, limit, offset)
		// Case 2: Only query specified (no platform filter)
	} else if query != "" {
		searchQuery = fmt.Sprintf(`
			fields name, summary, first_release_date, cover.url, rating, platforms.name, platforms.id;
			search "%s";
			limit %d;
			offset %d;
		`, query, limit, offset)
		// Case 3: Only platform specified (no search term)
	} else if platform != "" && platform != "All Platforms" {
		searchQuery = fmt.Sprintf(`
			fields name, summary, first_release_date, cover.url, rating, platforms.name, platforms.id;
			where platforms.name ~ *"%s"*;
			limit %d;
			offset %d;
		`, platform, limit, offset)
		// Case 4: Neither query nor platform specified - get top games
	} else {
		searchQuery = fmt.Sprintf(`
			fields name, summary, first_release_date, cover.url, rating, platforms.name, platforms.id;
			sort rating desc;
			limit %d;
			offset %d;
		`, limit, offset)
	}

	log.Printf("IGDB Search Query: %s", searchQuery)
	games, err := s.queryIGDB("games", searchQuery)
	if err != nil {
		log.Printf("Error querying IGDB for search: %v", err)
		return s.getMockGames(page, limit)
	}

	// Get total count for pagination
	var countQuery string

	// Use the same filtering logic for count
	if query != "" && platform != "" && platform != "All Platforms" {
		countQuery = fmt.Sprintf(`
			where search ~ "%s" & platforms.name ~ *"%s"*;
		`, query, platform)
	} else if query != "" {
		countQuery = fmt.Sprintf(`
			where search ~ "%s";
		`, query)
	} else if platform != "" && platform != "All Platforms" {
		countQuery = fmt.Sprintf(`
			where platforms.name ~ *"%s"*;
		`, platform)
	} else {
		countQuery = ";"
	}

	log.Printf("Count query: %s", countQuery)
	count, err := s.getCount("games", countQuery)
	if err != nil {
		log.Printf("Error getting count: %v", err)
		count = 100 // Default to 100 if we can't get the real count
	}

	return games, count, nil
}

// queryIGDB makes a query to the IGDB API
func (s *IGDBService) queryIGDB(endpoint, query string) ([]*models.Game, error) {
	url := fmt.Sprintf("https://api.igdb.com/v4/%s", endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", s.config.IGDBClientID)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error querying IGDB (status %d): %s", resp.StatusCode, string(body))
	}

	var igdbGames []struct {
		ID               int64   `json:"id"`
		Name             string  `json:"name"`
		Summary          string  `json:"summary"`
		FirstReleaseDate int64   `json:"first_release_date"`
		Rating           float64 `json:"rating"`
		Cover            struct {
			URL string `json:"url"`
		} `json:"cover"`
		Platforms []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"platforms"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&igdbGames); err != nil {
		return nil, err
	}

	var games []*models.Game
	for _, igdbGame := range igdbGames {
		// Process cover image if available
		coverURL := ""
		if igdbGame.Cover.URL != "" {
			coverURL = "https:" + igdbGame.Cover.URL
			coverURL = strings.Replace(coverURL, "t_thumb", "t_cover_big", 1)
		}

		// Process release date if available
		releaseDate := time.Time{}
		if igdbGame.FirstReleaseDate > 0 {
			releaseDate = time.Unix(igdbGame.FirstReleaseDate, 0)
		}

		// Process platforms - if we have a specific platform filter, prioritize it
		platform := ""
		s.platformMapMux.RLock()
		if len(igdbGame.Platforms) > 0 {
			// If no specific platform is prioritized, just use the first platform
			platform = igdbGame.Platforms[0].Name
		}
		s.platformMapMux.RUnlock()

		games = append(games, &models.Game{
			ID:          igdbGame.ID,
			Name:        igdbGame.Name,
			Platform:    platform,
			ReleaseDate: releaseDate,
			Description: igdbGame.Summary,
			Rating:      igdbGame.Rating,
			ImageURL:    coverURL,
		})
	}

	return games, nil
}

// getCount gets the count of entries that match a query
func (s *IGDBService) getCount(endpoint, query string) (int, error) {
	url := fmt.Sprintf("https://api.igdb.com/v4/%s/count", endpoint)

	// Do a simpler count query to avoid syntax errors
	// For the count endpoint, we typically don't need all the field selections
	// Just use the where clause if available
	cleanQuery := ""
	if strings.Contains(query, "where") {
		// Extract just the where clause
		parts := strings.Split(query, "where")
		if len(parts) > 1 {
			wherePart := parts[1]
			if idx := strings.Index(wherePart, ";"); idx > 0 {
				// Get just the condition part
				cleanQuery = "where " + wherePart[:idx] + ";"
			} else {
				cleanQuery = "where " + wherePart
			}
		}
	}

	// If we couldn't extract a where clause, use a basic query
	if cleanQuery == "" {
		cleanQuery = ";"
	}

	log.Printf("Count query: %s", cleanQuery)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(cleanQuery)))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Client-ID", s.config.IGDBClientID)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Count query failed with status %d: %s, query: %s", resp.StatusCode, string(body), cleanQuery)
		return 0, fmt.Errorf("error getting count (status %d): %s", resp.StatusCode, string(body))
	}

	var countResp struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&countResp); err != nil {
		return 0, err
	}

	log.Printf("Count result: %d", countResp.Count)
	return countResp.Count, nil
}

// getMockGames returns mock game data for testing
func (s *IGDBService) getMockGames(page, limit int) ([]*models.Game, int, error) {
	mockGames := []*models.Game{
		{
			ID:          1,
			Name:        "The Legend of Zelda: Breath of the Wild",
			Description: "Step into a world of discovery, exploration, and adventure in The Legend of Zelda: Breath of the Wild.",
			Platform:    "Nintendo Switch",
			ReleaseDate: time.Date(2017, 3, 3, 0, 0, 0, 0, time.UTC),
			Rating:      97.5,
			ImageURL:    "https://images.igdb.com/igdb/image/upload/t_cover_big/co3p2d.jpg",
		},
		{
			ID:          2,
			Name:        "Red Dead Redemption 2",
			Description: "America, 1899. The end of the wild west era has begun as lawmen hunt down the last remaining outlaw gangs.",
			Platform:    "PlayStation 4",
			ReleaseDate: time.Date(2018, 10, 26, 0, 0, 0, 0, time.UTC),
			Rating:      95.0,
			ImageURL:    "https://images.igdb.com/igdb/image/upload/t_cover_big/co1q1f.jpg",
		},
		{
			ID:          4,
			Name:        "God of War",
			Description: "His vengeance against the Gods of Olympus years behind him, Kratos now lives as a man in the realm of Norse Gods and monsters.",
			Platform:    "PlayStation 4",
			ReleaseDate: time.Date(2018, 4, 20, 0, 0, 0, 0, time.UTC),
			Rating:      94.0,
			ImageURL:    "https://images.igdb.com/igdb/image/upload/t_cover_big/co1tmu.jpg",
		},
		{
			ID:          5,
			Name:        "Bloodborne",
			Description: "An action RPG from the creators of the 'Souls' franchise, set in a Gothic world.",
			Platform:    "PlayStation 4",
			ReleaseDate: time.Date(2015, 3, 24, 0, 0, 0, 0, time.UTC),
			Rating:      91.0,
			ImageURL:    "https://images.igdb.com/igdb/image/upload/t_cover_big/co1rba.jpg",
		},
		{
			ID:          6,
			Name:        "Super Mario Odyssey",
			Description: "Join Mario on a massive, globe-trotting 3D adventure.",
			Platform:    "Nintendo Switch",
			ReleaseDate: time.Date(2017, 10, 27, 0, 0, 0, 0, time.UTC),
			Rating:      93.5,
			ImageURL:    "https://images.igdb.com/igdb/image/upload/t_cover_big/co1mxf.jpg",
		},
	}

	startIndex := (page - 1) * limit
	if startIndex >= len(mockGames) {
		return []*models.Game{}, len(mockGames), nil
	}

	endIndex := startIndex + limit
	if endIndex > len(mockGames) {
		endIndex = len(mockGames)
	}

	return mockGames[startIndex:endIndex], len(mockGames), nil
}

// GetGameByID returns a game by its ID
func (s *IGDBService) GetGameByID(id int64) (*models.Game, error) {
	// Fetch real data if token is available, otherwise use mock data
	if err := s.ensureValidToken(); err != nil {
		log.Printf("Warning: using mock data due to token error: %v", err)
		return s.getMockGameByID(id)
	}

	// 1. 기본 게임 정보 가져오기
	query := fmt.Sprintf(`
		fields name, summary, first_release_date, cover.url, rating, platforms.name;
		where id = %d;
	`, id)

	games, err := s.queryIGDB("games", query)
	if err != nil {
		log.Printf("Error querying IGDB for game ID %d: %v", id, err)
		return s.getMockGameByID(id)
	}

	if len(games) == 0 {
		return nil, fmt.Errorf("game not found with ID: %d", id)
	}

	game := games[0]

	// 2. 비디오 정보 가져오기
	videoQuery := fmt.Sprintf(`
		fields name, video_id;
		where game = %d;
	`, id)

	videos, err := s.getGameVideos("game_videos", videoQuery)
	if err != nil {
		log.Printf("Error querying IGDB for game videos, ID %d: %v", id, err)
	} else {
		// GameVideo 모델을 Video 모델로 변환
		var gameVideos []models.Video
		for _, video := range videos {
			gameVideos = append(gameVideos, models.Video{
				Title: video.Name,
				URL:   "https://www.youtube.com/watch?v=" + video.ID,
			})
		}
		game.Videos = gameVideos
	}

	// 3. 스크린샷 정보 가져오기
	screenshotQuery := fmt.Sprintf(`
		fields url;
		where game = %d;
	`, id)

	screenshots, err := s.getGameScreenshots("screenshots", screenshotQuery)
	if err != nil {
		log.Printf("Error querying IGDB for game screenshots, ID %d: %v", id, err)
	} else {
		game.Screenshots = screenshots
	}

	return game, nil
}

// getGameVideos fetches video information for a game
func (s *IGDBService) getGameVideos(endpoint, query string) ([]models.GameVideo, error) {
	url := fmt.Sprintf("https://api.igdb.com/v4/%s", endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", s.config.IGDBClientID)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error querying IGDB videos (status %d): %s", resp.StatusCode, string(body))
	}

	var igdbVideos []struct {
		Name    string `json:"name"`
		VideoID string `json:"video_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&igdbVideos); err != nil {
		return nil, err
	}

	var videos []models.GameVideo
	for _, video := range igdbVideos {
		videos = append(videos, models.GameVideo{
			ID:   video.VideoID,
			Name: video.Name,
		})
	}

	return videos, nil
}

// getGameScreenshots fetches screenshot information for a game
func (s *IGDBService) getGameScreenshots(endpoint, query string) ([]string, error) {
	url := fmt.Sprintf("https://api.igdb.com/v4/%s", endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", s.config.IGDBClientID)
	req.Header.Set("Authorization", "Bearer "+s.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error querying IGDB screenshots (status %d): %s", resp.StatusCode, string(body))
	}

	var igdbScreenshots []struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&igdbScreenshots); err != nil {
		return nil, err
	}

	var screenshots []string
	for _, screenshot := range igdbScreenshots {
		// IGDB는 상대 URL을 반환하므로 적절히 변환합니다
		screenshotURL := "https:" + screenshot.URL
		// 썸네일 대신 원본 크기 이미지 사용
		screenshotURL = strings.Replace(screenshotURL, "t_thumb", "t_original", 1)
		screenshots = append(screenshots, screenshotURL)
	}

	return screenshots, nil
}

// Helper methods to safely extract values from the JSON
func (s *IGDBService) getStringValue(data map[string]interface{}, key string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return ""
}

func (s *IGDBService) getFloatValue(data map[string]interface{}, key string) float64 {
	if value, ok := data[key].(float64); ok {
		return value
	}
	return 0
}

// getMockGameByID returns mock data for a specific game ID
func (s *IGDBService) getMockGameByID(id int64) (*models.Game, error) {
	// Get all mock games
	mockGames, _, _ := s.getMockGames(1, 10)

	// Find game with matching ID
	for _, game := range mockGames {
		if game.ID == id {
			return game, nil
		}
	}

	// If not found, create a mock game with the requested ID
	mockGame := &models.Game{
		ID:          id,
		Name:        fmt.Sprintf("Game %d", id),
		Description: "This is a mock game description.",
		Platform:    "Unknown Platform",
		ReleaseDate: time.Now().AddDate(-1, 0, 0),
		Rating:      85.0,
		ImageURL:    "https://via.placeholder.com/150",
		Videos: []models.Video{
			{
				Title: "Game Trailer",
				URL:   "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			},
			{
				Title: "Gameplay Demo",
				URL:   "https://www.youtube.com/watch?v=xvFZjo5PgG0",
			},
		},
		Screenshots: []string{
			"https://via.placeholder.com/1280x720?text=Game+Screenshot+1",
			"https://via.placeholder.com/1280x720?text=Game+Screenshot+2",
			"https://via.placeholder.com/1280x720?text=Game+Screenshot+3",
		},
	}
	return mockGame, nil
}
