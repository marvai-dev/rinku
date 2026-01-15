.PHONY: build run clean generate release validate last-release test goreleaser snapshot confidence confidence-reset confidence-dry lint

BINARY=bin/rinku

build: generate
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/rinku

generate:
	go generate ./cmd/rinku/...

test: generate
	go test ./...

lint:
	go vet ./...
	go tool staticcheck ./...
	go tool golangci-lint run ./...

release:
ifndef TAG
	$(error TAG is required. Usage: make release TAG=v0.1.0)
endif
	git tag $(TAG)
	git push origin $(TAG)

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
	rm -f cmd/rinku/index_gen.go

install: build
	cp $(BINARY) /usr/local/bin/rinku

validate:
	@./scripts/validate.sh

last-release:
	@git describe --tags --abbrev=0 2>/dev/null || echo "no tags"

goreleaser:
	GITHUB_TOKEN=$$(gh auth token) goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean

confidence:
	go run ./cmd/confidence

confidence-reset:
	go run ./cmd/confidence --reset

confidence-dry:
	go run ./cmd/confidence --dry-run
