DROP INDEX IF EXISTS idx_backups_next_run_at;
ALTER TABLE backups DROP COLUMN IF EXISTS enabled;
ALTER TABLE backups DROP COLUMN IF EXISTS next_run_at;
