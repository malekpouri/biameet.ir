# BiaMeet

BiaMeet is a simple, efficient meeting scheduler and voting system.

## Overview

BiaMeet allows users to create sessions with multiple time slots and invite others to vote on their preferred times. It is designed to be lightweight, fast, and easy to deploy.

## Features

- **Flexible Session Types**:
  - **Specific Times**: Propose exact time slots for voting.
  - **Weekly Pattern**: Define allowed days and time ranges for recurring meetings.
  - **Time Range**: Set a date and time window for users to propose their own times.
- **Smart Sharing**: Native Web Share API support with fallback to clipboard copy.
- **Admin Dashboard**: View system statistics (Sessions, Timeslots, Votes) at `/admin`.
- **User Experience**:
  - Beautiful Toast Notifications.
  - Mobile-optimized responsive design.
  - Jalali (Persian) Calendar support.
  - Voter transparency (names displayed under votes).

## API Endpoints

### Sessions

- `POST /api/v1/sessions`: Create a new session.
- `GET /api/v1/sessions/:id`: Get session details.
- `POST /api/v1/sessions/:id/timeslots`: Add a dynamic timeslot.

### Votes

- `POST /api/v1/sessions/:id/vote`: Submit a vote.

### Admin

- `GET /api/v1/admin/stats`: Get system statistics.

## Architecture

- **Backend**: Go + Fiber
- **Database**: SQLite (Docker volume)
- **Frontend**: Vanilla JS + Tailwind CSS
- **Deployment**: Docker Compose (Nginx + App)

## Local Development

### Prerequisites

- Go 1.21+
- Docker & Docker Compose

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

   Access the app at `http://localhost:8085`.

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
