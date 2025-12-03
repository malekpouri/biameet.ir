-- Up
ALTER TABLE timeslots ADD COLUMN created_by TEXT;
ALTER TABLE timeslots ADD COLUMN password_hash TEXT;

-- Down
-- SQLite does not support dropping columns easily in older versions, 
-- but for this app we can ignore down migration or use a more complex one if needed.
-- For now we leave it empty as per previous pattern or just comment.
