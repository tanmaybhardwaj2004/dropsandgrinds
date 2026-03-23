# DropsAndGrinds
All in one gaming platform to track your library, find the best deals, get price drops alert, read gaming news, and predict the lowest prices based on historical data.

## Features

- Game library with statuses: Want to buy, Purchased, Planning to play, Playing, Completed, Dropped
- Best deals and price comparison across platforms
- All-time low price tracking and notifications
- Price drop predictions based on historical patterns
- Ongoing and upcoming sale tracking
- Preferred payment method availability checker
- Latest gaming news linked to games

## Tech Stack

- Backend: Go (Golang)
- Database: PostgreSQL
- Cache: Redis
- Auth: JWT
- Frontend: HTML, CSS, JavaScript
- Infrastructure: AWS EC2, RDS, ElastiCache, CloudFront
- DevOps: Docker, Docker Compose, GitHub Actions
- Data APIs: RAWG, Steam, IGDB

## Running Locally

Prerequisites: Go, Docker

1. Clone the repository 
   git clone https://github.com/tanmaybhardwaj2004/dropsandgrinds.git

2. Navigate into the project
  cd dropsandgrinds

3. Start the server
   go run cmd/server/main.go

4. Health check
   http://localhost:8080/health

## Status

In active development.