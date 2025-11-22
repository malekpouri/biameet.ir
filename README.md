# BiaMeet

BiaMeet is a simple, efficient meeting scheduler.

## Overview
- **Backend**: Go + Fiber
- **Database**: SQLite
- **Frontend**: Static HTML/JS + Tailwind CSS
- **Timezone**: All times stored in UTC, displayed in Jalali (Solar Hijri).

## Local Development

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Node.js (for frontend build tools)

### Setup
1. Clone the repository.
2. Navigate to `backend` and run `go mod download`.
3. Navigate to `frontend` and run build scripts (TBD).

### Running with Docker
```bash
docker-compose up --build
```

## Branching Policy
- Feature branches: `feat/<descriptor>-<ticket>`
- Fix branches: `fix/<descriptor>-<ticket>`
- Commit messages: Conventional Commits (e.g., `feat(session): add create endpoint`)

## Migrations
SQL migrations are located in `/backend/db/migrations`.
