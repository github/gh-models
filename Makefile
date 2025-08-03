check: fmt vet tidy test
.PHONY: check

ci-lint:
	@echo "==> running Go linter <=="
	golangci-lint run --timeout 5m ./...
.PHONY: ci-lint

fmt:
	@echo "==> running Go format <=="
	gofmt -s -l -w .
.PHONY: fmt

vet:
	@echo "==> vetting Go code <=="
	go vet ./...
.PHONY: vet

tidy:
	@echo "==> running Go mod tidy <=="
	go mod tidy
.PHONY: tidy

test:
	@echo "==> running Go tests <=="
	go test -race -cover ./...
.PHONY: test

build:
	script/build
.PHONY: build

clean:
	@echo "==> cleaning up <=="
	rm -rf ./gh-models
.PHONY: clean

integration-test:
	@echo "==> running integration tests <=="
	@echo "Building binary first..."
	@$(MAKE) build
	@echo "Running integration tests against compiled binary..."
	go test -tags=integration -v ./integration/...
.PHONY: integration-test
