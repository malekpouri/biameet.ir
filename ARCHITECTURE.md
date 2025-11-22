# BiaMeet Architecture

## Database Schema (SQLite)

All timestamps must be stored as ISO 8601 UTC strings (`YYYY-MM-DDTHH:MM:SSZ`).

```sql
-- Migrations will be placed in /backend/db/migrations

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY, -- UUID
    title TEXT NOT NULL,
    creator_name TEXT NOT NULL,
    created_at_utc TEXT NOT NULL, -- ISO 8601 UTC
    expires_at_utc TEXT NOT NULL, -- ISO 8601 UTC
    archived_at_utc TEXT -- ISO 8601 UTC, NULL if active
);

CREATE TABLE IF NOT EXISTS timeslots (
    id TEXT PRIMARY KEY, -- UUID
    session_id TEXT NOT NULL,
    start_utc TEXT NOT NULL, -- ISO 8601 UTC
    end_utc TEXT NOT NULL, -- ISO 8601 UTC
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS votes (
    id TEXT PRIMARY KEY, -- UUID
    timeslot_id TEXT NOT NULL,
    voter_name TEXT NOT NULL,
    note TEXT,
    created_at_utc TEXT NOT NULL, -- ISO 8601 UTC
    FOREIGN KEY (timeslot_id) REFERENCES timeslots(id) ON DELETE CASCADE,
    UNIQUE(timeslot_id, voter_name) -- Prevent double voting for same slot
);
```

## API Specification

Base URL: `/api/v1`

### 1. Create Session
**POST** `/sessions`

**Request Body:**
```json
{
  "title": "Weekly Sync",
  "creator_name": "Ali",
  "timeslots": [
    {
      "start_utc": "2023-10-27T10:00:00Z",
      "end_utc": "2023-10-27T10:30:00Z"
    },
    {
      "start_utc": "2023-10-27T10:30:00Z",
      "end_utc": "2023-10-27T11:00:00Z"
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "id": "uuid-string",
  "link": "/meet/uuid-string"
}
```

### 2. Get Session Details
**GET** `/sessions/:id`

**Response (200 OK):**
```json
{
  "id": "uuid-string",
  "title": "Weekly Sync",
  "creator_name": "Ali",
  "created_at_utc": "2023-10-27T09:00:00Z",
  "timeslots": [
    {
      "id": "slot-uuid-1",
      "start_utc": "2023-10-27T10:00:00Z",
      "end_utc": "2023-10-27T10:30:00Z",
      "votes": [
        {
          "voter_name": "Reza",
          "note": "Prefer this"
        }
      ],
      "vote_count": 1
    }
  ]
}
```

### 3. Submit Vote
**POST** `/sessions/:id/vote`

**Request Body:**
```json
{
  "voter_name": "Sara",
  "votes": [
    {
      "timeslot_id": "slot-uuid-1",
      "note": "I might be late"
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "updated_session": { ... } // Optional: return updated session data
}
```

## Directory Structure
- `/backend`: Go + Fiber
- `/frontend`: Static HTML/JS + Tailwind
- `/docker`: Docker configurations
