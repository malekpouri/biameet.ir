-- Up
-- Every session page load filters timeslots by session and sorts by start time.
-- (votes are already covered by the UNIQUE(timeslot_id, voter_name) index.)
CREATE INDEX IF NOT EXISTS idx_timeslots_session_start ON timeslots(session_id, start_utc);

-- Server-side settings such as the vote-token signing key.
CREATE TABLE IF NOT EXISTS app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
