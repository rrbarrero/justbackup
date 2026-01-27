#!/bin/bash
# Params:
# HOOK_PARAM_HOST (default: localhost)
# HOOK_PARAM_PORT (default: 5432)
# HOOK_PARAM_USER
# HOOK_PARAM_PASSWORD
# HOOK_PARAM_DB_NAME
# Context:
# BACKUP_DEST

FILENAME="db_dump_${HOOK_PARAM_DB_NAME}.sql.gz"
FULL_PATH="${BACKUP_DEST}/${FILENAME}"

echo "Starting Postgres dump for ${HOOK_PARAM_DB_NAME} to ${FULL_PATH}..."

export PGPASSWORD="${HOOK_PARAM_PASSWORD}"
pg_dump -h "${HOOK_PARAM_HOST:-localhost}" \
        -p "${HOOK_PARAM_PORT:-5432}" \
        -U "${HOOK_PARAM_USER}" \
        "${HOOK_PARAM_DB_NAME}" | gzip > "${FULL_PATH}"

if [ $? -eq 0 ]; then
  echo "Postgres dump completed successfully."
else
  echo "Error: Postgres dump failed."
  exit 1
fi
