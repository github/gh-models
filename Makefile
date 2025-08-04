check: fmt vet tidy test
.PHONY: check

clean:
	@echo "==> cleaning up <=="
	rm -rf ./gh-models
.PHONY: clean

build:
	@echo "==> building gh-models binary <=="
	script/build
.PHONY: build

ci-lint:
	@echo "==> running Go linter <=="
	golangci-lint run --timeout 5m ./...
.PHONY: ci-lint

integration: check build
	@echo "==> running integration tests <=="
	cd integration && go mod tidy && go test -v -timeout=5m
.PHONY: integration

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
