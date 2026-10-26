.PHONY: build swagger install restart test e2e fmt vet

SWAG := go run github.com/swaggo/swag/cmd/swag@v1.16.6

# Build pcm and pcm-server into bin/, after the swagger spec.
build: swagger
	go build -o bin/ ./cmd/...

# Generate the swagger spec into docs/swagger/.
swagger:
	$(SWAG) init -g cmd/pcm-server/main.go -o docs/swagger

# Install pcm and pcm-server into $(go env GOPATH)/bin, then the Claude Code skill.
install:
	go install ./cmd/...
	$(shell go env GOPATH)/bin/pcm skill install

# Restart the background server with the installed binary.
restart:
	-pcm server stop
	pcm server start

# Unit tests.
test:
	go test ./...

# BDD specs (godog), real binaries.
e2e:
	go test -v -tags e2e ./e2e/...

fmt:
	gofmt -w .

vet:
	go vet ./...
	go vet -tags e2e ./e2e/...
