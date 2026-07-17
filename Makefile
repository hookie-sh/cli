.PHONY: proto build-gui build-cli build-cli-dev build-cli-prod ensure-cli-gui-dist test deps

CLI_ENV_PREFIX = if [ -f .env ]; then set -a; . ./.env; set +a; fi; if [ -n "$$CLERK_PUBLISHABLE_KEY" ] && [ -z "$$NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY" ]; then export NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY="$$CLERK_PUBLISHABLE_KEY"; fi;
CLI_DEV_LDFLAGS = -X github.com/hookie-sh/cli/internal/auth.PublishableKey=$$CLERK_PUBLISHABLE_KEY -X github.com/hookie-sh/cli/internal/auth.WebAppURL=$${HOOKIE_APP_URL:-https://app.hookie.sh}

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/relay.proto

build-gui:
	cd packages/gui && pnpm exec vite build

build-cli: build-cli-dev

build-cli-dev:
	$(CLI_ENV_PREFIX) $(MAKE) build-gui
	rm -rf internal/gui/dist
	cp -r packages/gui/dist internal/gui/dist
	$(CLI_ENV_PREFIX) go build -tags dev -ldflags "$(CLI_DEV_LDFLAGS)" -o bin/hookie .

build-cli-prod: build-gui
	rm -rf internal/gui/dist
	cp -r packages/gui/dist internal/gui/dist
	go build -o bin/hookie .

ensure-cli-gui-dist:
	@if [ ! -f internal/gui/dist/index.html ]; then \
		mkdir -p internal/gui/dist && \
		printf '%s\n' '<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><title>Hookie</title></head><body></body></html>' > internal/gui/dist/index.html; \
	fi

test: ensure-cli-gui-dist
	go test ./...

deps:
	go mod download && go mod tidy
