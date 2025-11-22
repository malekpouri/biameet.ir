ALTER TABLE sessions ADD COLUMN type TEXT DEFAULT 'fixed';
ALTER TABLE sessions ADD COLUMN dynamic_config TEXT;
