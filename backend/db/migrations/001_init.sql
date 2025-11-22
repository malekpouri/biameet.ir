-- Up
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

-- Down
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS timeslots;
DROP TABLE IF EXISTS sessions;
