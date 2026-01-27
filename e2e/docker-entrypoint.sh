#!/bin/bash
set -e

# Setup authorized keys if mounted
if [ -f "/mnt/public_key" ]; then
    echo "Adding public key to authorized_keys..."
    cat /mnt/public_key > /home/backup-test/.ssh/authorized_keys
    chmod 600 /home/backup-test/.ssh/authorized_keys
    chown backup-test:backup-test /home/backup-test/.ssh/authorized_keys
fi

# Fix permissions on data directory if needed (to ensure backup-test can read it)
# We assume data is mounted at /mnt/source_data
if [ -d "/mnt/source_data" ]; then
    # In a real scenario we might need to be careful with chown on valid volumes,
    # but for this test container it's fine.
    chown -R backup-test:backup-test /mnt/source_data || true
fi

exec "$@"
