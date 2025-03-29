# Game Retro Stock

A retro game information and price tracking website that integrates with IGDB API and PriceCharting data.

## Features

- Game search by platform and title
- Platform-based game browsing
- Top 100 games list from IGDB
- Detailed game information pages
- Game price history tracking
- YouTube video integration
- Price trend graphs

## Tech Stack

- Backend: Go (Gin framework)
- Frontend: React with TypeScript
- APIs: IGDB API, PriceCharting data
- Database: PostgreSQL

## Setup Instructions

1. Clone the repository
2. Set up environment variables in `.env`
3. Install backend dependencies:
   ```bash
   go mod tidy
   ```
4. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   ```
5. Start the backend server:
   ```bash
   go run main.go
   ```
6. Start the frontend development server:
   ```bash
   cd frontend
   npm start
   ```

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```
PORT=8080
IGDB_CLIENT_ID=your_client_id
IGDB_CLIENT_SECRET=your_client_secret
DATABASE_URL=postgresql://user:password@localhost:5432/gameretrostock
```

## API Documentation

### Backend Endpoints

- `GET /api/games` - Get list of games with filters
- `GET /api/games/:id` - Get detailed game information
- `GET /api/platforms` - Get list of gaming platforms
- `GET /api/games/:id/prices` - Get price history for a game 