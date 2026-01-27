ALTER TABLE hosts ADD COLUMN host_path TEXT NOT NULL DEFAULT '';

-- Update existing records with a slugified version of the name
UPDATE hosts SET host_path = LOWER(REPLACE(name, ' ', '-'));
