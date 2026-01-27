#!/bin/bash
# Params:
# HOOK_PARAM_HOST (default: localhost)
# HOOK_PARAM_PORT (default: 27017)
# HOOK_PARAM_USER
# HOOK_PARAM_PASSWORD
# HOOK_PARAM_DB_NAME
# HOOK_PARAM_AUTH_DB (default: admin)
# Context:
# BACKUP_DEST

FILENAME="db_dump_${HOOK_PARAM_DB_NAME:-all}.archive.gz"
FULL_PATH="${BACKUP_DEST}/${FILENAME}"

echo "Starting MongoDB dump for ${HOOK_PARAM_DB_NAME:-all} to ${FULL_PATH}..."

AUTH_ARGS=""
if [ -n "${HOOK_PARAM_USER}" ]; then
  AUTH_ARGS="--username ${HOOK_PARAM_USER} --password ${HOOK_PARAM_PASSWORD} --authenticationDatabase ${HOOK_PARAM_AUTH_DB:-admin}"
fi

DB_ARG=""
if [ -n "${HOOK_PARAM_DB_NAME}" ]; then
  DB_ARG="--db ${HOOK_PARAM_DB_NAME}"
fi

mongodump --host "${HOOK_PARAM_HOST:-localhost}" \
          --port "${HOOK_PARAM_PORT:-27017}" \
          ${AUTH_ARGS} \
          ${DB_ARG} \
          --archive --gzip > "${FULL_PATH}"

if [ $? -eq 0 ]; then
  echo "MongoDB dump completed successfully."
else
  echo "Error: MongoDB dump failed."
  exit 1
fi
