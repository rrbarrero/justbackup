-- Migration to simplify hook parameter encryption format
-- From: {"encrypted": true, "data": "..."}
-- To: "..." (JSON string)

-- 1. Transform existing compliant rows
UPDATE backup_hooks
SET
    params = to_jsonb(params ->> 'data')
WHERE
    params ? 'encrypted'
    AND params ? 'data';

-- 2. Delete any rows that are still objects (not migrated or invalid)
-- This ensures that only the simple encrypted string remains.
DELETE FROM backup_hooks WHERE jsonb_typeof(params) = 'object';