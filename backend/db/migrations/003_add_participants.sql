-- Up
CREATE TABLE IF NOT EXISTS participants (
    session_id TEXT NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT,
    created_at_utc TEXT NOT NULL,
    PRIMARY KEY (session_id, name),
    FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE
);
