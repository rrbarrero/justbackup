CREATE TABLE backup_hooks (
    id UUID PRIMARY KEY,
    backup_id UUID NOT NULL REFERENCES backups (id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phase VARCHAR(50) NOT NULL, -- 'pre' or 'post'
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    params JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_backup_hooks_backup_id ON backup_hooks (backup_id);