# Architecture

## Tech Stack

- **Language**: Go (Golang)
- **Framework**: Fiber (v2/v3)
- **Database**: SQLite3
- **Frontend**: Vanilla JS + Tailwind CSS (No build step required for JS, Tailwind via CDN or CLI)
- **Containerization**:
  - Backend: Multi-stage Docker build (Go -> Alpine)
  - Frontend: Nginx serving static files
  - Network: Internal Docker network for API communication (Frontend proxies /api to Backend:8080)

## Data Flow

1. Client sends requests with UTC ISO 8601 timestamps.
2. Backend validates and stores timestamps as-is in SQLite.
3. Backend returns UTC ISO 8601 timestamps.
4. Client converts UTC to Jalali for display.

## Directory Structure

- `/backend`: Go source code
- `/frontend`: Static assets
- `/docker`: Dockerfiles
- `/scripts`: Utility scripts

## Database Schema

See `backend/db/migrations` for SQL files.
