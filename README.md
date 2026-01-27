# JustBackup

[![CI](https://github.com/rrbarrero/justbackup/actions/workflows/ci.yml/badge.svg)](https://github.com/rrbarrero/justbackup/actions/workflows/ci.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rrbarrero/justbackup)](https://github.com/rrbarrero/justbackup)

![Web UI Placeholder](/assets/dashboard.webp)

**JustBackup** is a self-hosted backup orchestrator designed for simplicity and reliability. It manages centralized backups across your infrastructure using standard tools: **SSH** and **rsync**.

## Key Features

*   🚀 **Agentless & Scalable**: Works with standard SSH and rsync on your target hosts. Uses Redis as a robust communication backbone for asynchronous task queueing.
*   🔐 **Data Protection & Privacy**: Support for incremental and encrypted backups. Files are stored using standard structures—no proprietary formats or vendor lock-in.
*   💻 **First-class CLI & API**: Full-featured command-line interface and a modern web dashboard with built-in API (Swagger) for automation and monitoring.
*   🪝 **Extensible & Ready-to-use**: Flexible hook system and built-in plugins (e.g., PostgreSQL dumps) to handle complex backup scenarios out of the box.
*   🛡️ **Solid Architecture**: Built following Clean Architecture and Domain-Driven Design (DDD) principles to ensure stability and maintainability.

## Quick start (Docker Compose)

### Prerequisites

- Docker + Docker Compose
- A host directory for backups (e.g., `/mnt/backups`). **RAID setup is highly recommended** for data redundancy and reliability.

### 1) Configure environment

```bash
cp env.example .env
```

Edit `.env` and set at least:

- `BACKUP_ROOT` (host path where backups will live)
- `ENCRYPTION_KEY` (32+ chars; used for optional encryption)
- `JWT_SECRET` (API auth)
- `CORS_ALLOWED_ORIGIN` (where the web UI is served from)

### 2) Install and Start

Simply run:

```bash
make install
```

This command automates the rest of the process:
- Generates SSH keys for the worker (`secrets/ssh/id_ed25519_backup`).
- Sets the correct permissions for the secrets.
- Builds and starts all services in the background.

### 4) Open the UI

We strongly recommend putting an HTTPS reverse proxy in front (Nginx Proxy Manager, Traefik, etc.) and serving the UI and API over your HTTPS domain.

### Optional: Quick Start with Traefik
If you want automatic HTTPS with Traefik, make sure you have `APP_DOMAIN` and `ACME_EMAIL` set in your `.env` file, then run:

```bash
docker compose -f docker-compose.yml -f docker-compose.traefik.yml up -d
```

- Web UI: `https://your-domain.example`
- API (v1): `https://your-domain.example/api/v1`
- Swagger UI: `https://your-domain.example/api/v1/swagger/`

The first time you access the UI, you will be prompted to set an admin password.

## CLI (justbackup)

We provide the CLI directly from the web UI. The binary is compiled during the Docker Compose build (and can also be compiled at any time from source). Open the **CLI** page in the dashboard to download it and see the exact configuration steps for your instance (including the backend URL and where to find your API token).

The CLI stores its config at `~/.config/joblist.conf`.

![CLI Placeholder](/assets/cli.webp)

### Common tasks

List hosts:

```bash
justbackup hosts
```

Bootstrap a host (installs the worker public key and registers the host):

```bash
justbackup bootstrap --host 192.168.1.10 --name my-server --user root --port 22
```

Create a backup task:

```bash
justbackup add-backup \
  --host-id <host-id> \
  --path /etc \
  --dest etc \
  --schedule "0 2 * * *" \
  --excludes "*.tmp,*.cache" \
  --incremental
```

Run a backup immediately:

```bash
justbackup run <backup-id>
```

List backups and explore files:

```bash
justbackup backups
justbackup files <backup-id> --path /etc/nginx
```

Search across backups:

```bash
justbackup search "*.conf"
```

Restore to your local machine:

```bash
justbackup restore <backup-id> --local --path /etc/nginx --dest ./restore
```

Restore to a remote host via rsync:

```bash
justbackup restore <backup-id> --remote --path /etc/nginx --to-host <target-host-id> --to-path /srv/restore
```

Decrypt an encrypted backup artifact offline:

```bash
justbackup decrypt --file /path/to/backup.tar.gz.enc --out ./backup.tar.gz --id <backup-id> --key <master-key>
```

## Extensibility

JustBackup is designed to be extended:

- **Swagger/OpenAPI**: browse or generate clients via `http://localhost:8080/swagger/`.
- **API-first**: everything the UI and CLI do is available over HTTP.
- **Interfaces**: core domain abstractions live under `internal/*/domain/interfaces`, with concrete implementations in `internal/*/infrastructure`.

To regenerate API docs locally:

```bash
swag init -g cmd/server/main.go
```

Swagger outputs to `docs/swagger.yaml` and `docs/swagger.json`.

## Configuration reference

See `env.example` for the full list. Key variables:

- `ENVIRONMENT`: `dev` for in-memory repositories, otherwise production
- `WORKER_INSTANCES`: number of worker nodes to run (adjust based on system load)
- `BACKUP_ROOT`: host path where backups are stored
- `ENCRYPTION_KEY`: master key for optional encryption
- `JWT_SECRET`: API auth signing key
- `REDIS_HOST` / `REDIS_PORT`
- `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME`
- `CORS_ALLOWED_ORIGIN`

## Troubleshooting

- **Workers cannot SSH**: verify `secrets/ssh/id_ed25519_backup.pub` is installed on the target host under `~/.ssh/authorized_keys`.
- **Permission errors**: ensure `BACKUP_ROOT` exists and the worker UID/GID can write to it (see `WORKER_UID`/`WORKER_GID`).
- **No API access**: confirm `JWT_SECRET` is set and the CLI token matches `/login` output.

## Disclaimer

**USE AT YOUR OWN RISK.** This software is provided "as is" without any warranty of any kind. The authors and contributors shall not be held liable for any data loss, system failure, or other damages arising from the use of this software. Always ensure you have independent, verified backups of critical data before using any backup orchestration tool.

## License

This project is licensed under the **GNU General Public License v3.0**. See the [LICENSE](LICENSE) file for details.
