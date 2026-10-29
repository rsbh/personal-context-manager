.PHONY: build swagger client web install restart release-check test web-test web-lint web-fmt e2e e2e-ui fmt lint vet

SWAG := go run github.com/swaggo/swag/cmd/swag@v1.16.6

# Generate the swagger spec into docs/swagger/.
swagger:
	$(SWAG) init -g cmd/pcm-server/main.go -o docs/swagger

# Generate the TS API client (web/src/client) from the swagger spec.
client: swagger
	cd web && npm run gen:api

# Build the web UI into web/dist/app (embedded by pcm-server).
web:
	cd web && npm ci && npm run build

# Build pcm and pcm-server into bin/, after the swagger spec and the UI.
build: swagger web
	go build -o bin/ ./cmd/...

# Install pcm and pcm-server into $(go env GOPATH)/bin, with the UI, then the Claude Code skill.
install: web
	go install ./cmd/...
	$(shell go env GOPATH)/bin/pcm skill install

# Check the release config and build a snapshot (no publish) into dist/.
release-check:
	goreleaser check
	goreleaser release --snapshot --clean
	scripts/release-check.sh

# Restart the background server with the installed binary.
restart:
	-pcm server stop
	pcm server start

# Unit tests: Go and web.
test: web-test
	go test ./...

web-test:
	cd web && npm ci && npm run typecheck && npm test

# BDD specs (godog), real binaries. Skips @ui features.
e2e:
	go test -v -tags e2e ./e2e/...

# UI BDD specs (@ui features): Playwright + playwright-bdd, real binaries with the UI.
# First time: cd web && npx playwright install chromium
e2e-ui: web
	cd web && npm run e2e

# Web lint: oxlint, oxfmt check, react-doctor.
web-lint:
	cd web && npm run lint && npm run fmt:check && npm run doctor

web-fmt:
	cd web && npm run fmt

fmt: web-fmt
	gofmt -w .

# Go lint: golangci-lint (pinned as a go.mod tool). Config: .golangci.yml.
lint:
	go tool golangci-lint run ./...

vet:
	go vet ./...
	go vet -tags e2e ./e2e/...
