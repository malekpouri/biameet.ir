# BiaMeet

BiaMeet is a simple, efficient meeting scheduler and voting system.

## Overview

BiaMeet allows users to create sessions with multiple time slots and invite others to vote on their preferred times. It is designed to be lightweight, fast, and easy to deploy.

## Features

- Create meeting sessions with title and creator name.
- Define multiple time slots (UTC).
- Share links for voting.
- Users vote for time slots (Yes/No/Maybe notes).
- Real-time-ish updates (on refresh).
- Jalali date support for UI.

## Architecture

- **Backend**: Go + Fiber
- **Database**: SQLite (Docker volume)
- **Frontend**: Static HTML/JS + Tailwind CSS
- **Deployment**: Docker Compose (Nginx + App)

## Local Development

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Node.js (optional, for frontend tooling if needed later)

### Running Locally

1. **Clone the repository**:

   ```bash
   git clone https://github.com/malekpouri/biameet.ir.git
   cd biameet.ir
   ```

2. **Start with Docker Compose**:

   ```bash
   docker-compose up --build
   ```

   Access the app at `http://localhost:80`.

3. **Manual Backend Run**:

   ```bash
   cd backend
   go run cmd/main.go
   ```

4. **Manual Frontend Serve**:
   Serve the `frontend/src` directory with any static file server.

## Branching Policy

- **Feature Branches**: `feat/<short-desc>-<ticket>`
- **Fix Branches**: `fix/<short-desc>-<ticket>`
- **Commits**: Conventional Commits (e.g., `feat(auth): add login`).

## License

MIT
