# Architecture

## Tech Stack

- **Language**: Go (Golang)
- **Framework**: Fiber v2 (fasthttp)
- **Database**: SQLite3 (pure-Go `modernc.org/sqlite`, WAL mode)
- **Frontend**: Vanilla JS modules bundled by esbuild, Tailwind CSS, self-hosted Vazirmatn font
- **Containerization**: one multi-stage build (Node → Go → `scratch`). The Go binary embeds the built frontend and serves everything; there is no nginx.

## Request Flow

- `/api/v1/*`: JSON API. Write endpoints are rate-limited per client IP.
- `/`, `/admin`, `/<5-char id>`: `frontend/src/index.html` rendered as a Go `html/template` with per-page title, description, Open Graph and robots tags. The home page is rendered and compressed once and then cached. Session, admin and 404 pages are `noindex`.
- `/assets/*`, `/favicon.svg`, `/og.png`: embedded static files, compressed once at startup. The template links them with a `?v=<content hash>` query, which makes them cacheable for a year.
- `/robots.txt`, `/sitemap.xml`: generated from `BASE_URL`.

## Data Flow

1. Client sends requests with UTC ISO 8601 timestamps.
2. Backend validates and normalizes them to `YYYY-MM-DDTHH:MM:SS.000Z`, then stores them in SQLite.
3. Backend returns UTC ISO 8601 timestamps.
4. Client converts UTC to Jalali for display.

## Participants and Passwords

- There are no accounts. A participant is identified by a (session, name) pair, with an optional bcrypt password hash.
- A successful vote returns an HMAC token bound to that hash. The browser stores the token (never the password) and sends it for later edits.
- bcrypt runs outside write transactions, so it never holds the SQLite write lock.

## Directory Structure

- `/backend`: Go source code (`web/dist` receives the frontend build and is embedded)
- `/frontend`: frontend sources and build script
- `Dockerfile`, `docker-compose.yml`: deployment

## Database Schema

See `backend/db/migrations`. Migrations are embedded in the binary and applied once each, tracked in `schema_migrations`.
