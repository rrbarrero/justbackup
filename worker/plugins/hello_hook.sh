#!/bin/bash
echo "Hello from Hook! Phase: ${HOOK_PHASE:-unknown}. Destination: ${BACKUP_DEST}"
echo "Param foo: ${HOOK_PARAM_FOO}"
exit 0
