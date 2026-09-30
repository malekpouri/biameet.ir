# BiaMeet

BiaMeet is a simple, efficient meeting scheduler and voting system.

## Overview

BiaMeet allows users to create sessions with multiple time slots and invite others to vote on their preferred times. It is designed to be lightweight, fast, and easy to deploy: a single ~20 MB container that idles at about 20 MB of RAM.

## Features

- **Flexible Session Types**:
  - **Specific Times**: Propose exact time slots for voting.
  - **Weekly Pattern**: Define allowed days and time ranges for recurring meetings, with a link that expires after a chosen number of weeks.
  - **Time Range**: Set a date and time window for users to propose their own times.
- **Smart Sharing**: Native Web Share API support with fallback to clipboard copy, plus rich link previews (Open Graph image and per-session titles).
- **Admin Dashboard**: View system statistics (Sessions, Timeslots, Votes) at `/admin`, protected by `ADMIN_TOKEN`.
- **User Experience**:
  - Jalali (Persian) calendar, self-hosted Vazirmatn font, automatic dark mode.
  - Mobile-optimized responsive design; the most-voted slot is highlighted.
  - Voter transparency (names displayed under votes).
  - **Secure Voting**: Optional password protection so only the voter can edit their vote later. The browser keeps a signed token, never the password.
  - **Secure Timeslots**: Participants propose timeslots under their name and password and can delete their own proposals while nobody else has voted for them.

## API Endpoints

Errors are returned as `{"error": "<code>", "message": "<Persian text>"}` with a matching HTTP status.

### Sessions

- `POST /api/v1/sessions`: Create a new session.
- `GET /api/v1/sessions/:id`: Get session details.
- `POST /api/v1/sessions/:id/timeslots`: Propose a timeslot (weekly/time-range sessions).
- `DELETE /api/v1/sessions/:id/timeslots/:ts_id`: Delete your own proposed timeslot.

### Votes

- `POST /api/v1/sessions/:id/vote`: Submit or replace your votes (an empty list withdraws them). Returns a `token` usable instead of the password for later edits.

### Admin

- `GET /api/v1/admin/stats`: System statistics. Requires `Authorization: Bearer <ADMIN_TOKEN>`.

## Architecture

- **Backend**: Go + Fiber, one static binary that also serves the frontend (embedded, pre-compressed with brotli/gzip).
- **Database**: SQLite in WAL mode (Docker volume).
- **Frontend**: Vanilla JS modules bundled with esbuild + Tailwind CSS.
- **Deployment**: Docker Compose, one container built `FROM scratch`.

See [ARCHITECTURE.md](ARCHITECTURE.md) for details.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | HTTP port |
| `DB_PATH` | `biameet.db` (`/data/biameet.db` in Docker) | SQLite file |
| `BASE_URL` | `https://biameet.ir` | Used for canonical URLs, Open Graph, sitemap |
| `ADMIN_TOKEN` | *(empty: admin disabled)* | Token for `/admin` |
| `PROXY_HEADER` | *(empty)* | Header carrying the client IP from your reverse proxy, e.g. `X-Forwarded-For` |
| `WRITE_RATE_LIMIT` | `30` | Write requests per minute per IP (`0` disables) |
| `SECRET_KEY` | *(generated and stored in the DB)* | Key that signs vote tokens |
| `LOG_REQUESTS` | *(off)* | `1` logs every request |
| `WEB_DIR` | *(empty: embedded files)* | Serve the frontend from disk (development) |

## Local Development

### Prerequisites

- Go 1.24+
- Node.js 20+
- Docker & Docker Compose (for deployment)

### Running Locally

1. **Clone the repository**:

   ```bash
   git clone https://github.com/malekpouri/biameet.ir.git
   cd biameet.ir
   ```

2. **Start with Docker Compose**:

   ```bash
   docker compose up --build
   ```

   Access the app at `http://localhost:8085`.

3. **Manual run** (two terminals):

   ```bash
   cd frontend && npm install && npm run dev      # rebuilds into backend/web/dist on change
   cd backend && WEB_DIR=web/dist go run ./cmd    # http://localhost:8080
   ```

   For a production-like binary, run `npm run build` once and then `go build ./cmd`: the frontend is embedded.

4. **Tests**:

   ```bash
   cd backend && go test ./tests/...
   ```

## Upgrading from 1.x

Version 2 replaces the nginx + backend containers with a single container that uses the same `sqlite_data` volume (now mounted at `/data`). Just pull and rebuild:

```bash
docker compose down
docker compose up -d --build
```

On start the app takes ownership of `/data` (1.x left it owned by root), switches to an unprivileged user, and keeps all existing data; old migrations are detected and not re-run. It is still published on port 8085, so a reverse proxy in front needs no change.

## Versioning

The version lives in `backend/version/version.go`. It is shown in the page footer and startup log, returned by `GET /health`, and printed by `docker compose exec app /biameet version`.

## Branching Policy

- **Feature Branches**: `feat/<short-desc>-<ticket>`
- **Fix Branches**: `fix/<short-desc>-<ticket>`
- **Commits**: Conventional Commits (e.g., `feat(auth): add login`).

## License

MIT
