CREATE TABLE maintenance_tasks (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    schedule VARCHAR(50) NOT NULL,
    next_run_at TIMESTAMP WITH TIME ZONE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial purge task
INSERT INTO
    maintenance_tasks (
        id,
        name,
        type,
        schedule,
        next_run_at
    )
VALUES (
        gen_random_uuid (),
        'Purge Incremental Backups',
        'purge',
        '09 6 * * *',
        CURRENT_TIMESTAMP
    );