# Include .env file to get WORKER_UID/GID
-include .env
export

.PHONY: build-server build-client server-test frontend-validate lint secrets deploy install up

build-server:
	go build -o server cmd/server/main.go

build-client:
	go build -ldflags "-X main.DefaultBackendURL=http://localhost:8080" -o justbackup cmd/client/main.go

install-client: build-client
	echo "Installing justbackup to ~/.local/bin/"
	install -m 755 justbackup ~/.local/bin/

server-test:
	ENVIRONMENT=dev go test ./...

frontend-validate:
	npm run validate --prefix web/

lint:
	golangci-lint run

format:
	go fmt ./...
	npm run format --prefix web/

swagger:
	swag init -g cmd/server/main.go

check: format swagger lint server-test frontend-validate

install:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Please copy env.example to .env and configure it."; \
		exit 1; \
	fi
	@mkdir -p secrets/ssh
	@if [ ! -f secrets/ssh/id_ed25519_backup ]; then \
		echo "Generating SSH keys..."; \
		ssh-keygen -t ed25519 -f secrets/ssh/id_ed25519_backup -C "backup@local" -N ""; \
	fi
	@if [ ! -f secrets/ssh/known_hosts ]; then \
		touch secrets/ssh/known_hosts; \
	fi
	$(MAKE) fix-secrets
	@echo "Starting the application..."
	docker compose up --build -d
	@echo "Deployment complete! Access the UI at the URL configured in your .env"

secrets:
	@mkdir -p secrets/ssh
	@if [ ! -f secrets/ssh/id_ed25519_backup ]; then \
		ssh-keygen -t ed25519 -f secrets/ssh/id_ed25519_backup -C "backup@local" -N ""; \
	else \
		echo "Secrets already exist, skipping generation."; \
	fi
	@if [ ! -f secrets/ssh/known_hosts ]; then \
		touch secrets/ssh/known_hosts; \
	fi
	$(MAKE) fix-secrets

fix-secrets:
	@# Allow host user to enter the directory but keep files restricted
	chmod 755 secrets/ssh
	chmod 644 secrets/ssh/id_ed25519_backup.pub
	@if [ -f secrets/ssh/known_hosts ]; then chmod 644 secrets/ssh/known_hosts; fi
	@if [ -n "$(WORKER_UID)" ]; then \
		echo "Fixing ownership for UID $(WORKER_UID)..."; \
		docker run --rm -v $$(pwd)/secrets/ssh:/mnt/ssh alpine sh -c "chown -R $(WORKER_UID):$(WORKER_GID) /mnt/ssh && chmod 600 /mnt/ssh/id_ed25519_backup"; \
	fi

up: swagger
	docker compose up --build -d && docker compose logs -f

test-e2e:
	@mkdir -p fixtures/destination fixtures/decrypted
	@chmod 777 fixtures/destination fixtures/decrypted || true
	TEST="$(TEST)" ENVIRONMENT=dev docker compose -f docker-compose.yml -f docker-compose.e2e.yml up --build --exit-code-from e2e && jq '.status' e2e/test-results/.last-run.json
	@echo "Cleaning up fixtures..."
	@docker run --rm -v $(PWD)/fixtures/destination:/mnt/dest -v $(PWD)/fixtures/decrypted:/mnt/decrypted alpine sh -c "rm -rf /mnt/dest/* /mnt/decrypted/* && chmod 777 /mnt/dest /mnt/decrypted || true"

deploy:
	git pull && docker compose -f docker-compose.prod.yml up --build -d && docker compose logs -f

