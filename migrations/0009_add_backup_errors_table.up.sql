CREATE TABLE IF NOT EXISTS backup_errors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    job_id VARCHAR(255) NOT NULL,
    backup_id UUID NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    error_message TEXT NOT NULL,
    CONSTRAINT fk_backup FOREIGN KEY (backup_id) REFERENCES backups (id) ON DELETE CASCADE
);