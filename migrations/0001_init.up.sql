-- Schema initialization for Postgres repositories (hosts, backups)
CREATE TABLE IF NOT EXISTS hosts (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    hostname TEXT NOT NULL,
    "user" TEXT NOT NULL,
    port INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS backups (
    id UUID PRIMARY KEY,
    host_id UUID NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    destination TEXT NOT NULL,
    status TEXT NOT NULL,
    schedule TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_run TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_backups_host_id ON backups(host_id);
CREATE INDEX IF NOT EXISTS idx_backups_status ON backups(status);
