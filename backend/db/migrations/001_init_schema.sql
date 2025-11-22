-- Up
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    creator_name TEXT NOT NULL,
    created_at_utc TEXT NOT NULL,
    expires_at_utc TEXT,
    archived_at_utc TEXT
);

CREATE TABLE IF NOT EXISTS timeslots (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    start_utc TEXT NOT NULL,
    end_utc TEXT NOT NULL,
    FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS votes (
    id TEXT PRIMARY KEY,
    timeslot_id TEXT NOT NULL,
    voter_name TEXT NOT NULL,
    note TEXT,
    created_at_utc TEXT NOT NULL,
    FOREIGN KEY(timeslot_id) REFERENCES timeslots(id) ON DELETE CASCADE,
    UNIQUE(timeslot_id, voter_name)
);


