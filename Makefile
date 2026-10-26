.PHONY: build install restart test e2e fmt vet

# Build pcm and pcm-server into bin/.
build:
	go build -o bin/ ./cmd/...

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
