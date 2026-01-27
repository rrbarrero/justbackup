ALTER TABLE backups ADD COLUMN next_run_at TIMESTAMPTZ;
ALTER TABLE backups ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX idx_backups_next_run_at ON backups(next_run_at);
