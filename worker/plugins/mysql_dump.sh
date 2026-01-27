#!/bin/bash
# Params:
# HOOK_PARAM_HOST (default: localhost)
# HOOK_PARAM_PORT (default: 3306)
# HOOK_PARAM_USER
# HOOK_PARAM_PASSWORD
# HOOK_PARAM_DB_NAME
# Context:
# BACKUP_DEST

FILENAME="db_dump_${HOOK_PARAM_DB_NAME}.sql.gz"
FULL_PATH="${BACKUP_DEST}/${FILENAME}"

echo "Starting MySQL dump for ${HOOK_PARAM_DB_NAME} to ${FULL_PATH}..."

mysqldump -h "${HOOK_PARAM_HOST:-localhost}" \
          -P "${HOOK_PARAM_PORT:-3306}" \
          -u "${HOOK_PARAM_USER}" \
          -p"${HOOK_PARAM_PASSWORD}" \
          "${HOOK_PARAM_DB_NAME}" | gzip > "${FULL_PATH}"

if [ $? -eq 0 ]; then
  echo "MySQL dump completed successfully."
else
  echo "Error: MySQL dump failed."
  exit 1
fi
