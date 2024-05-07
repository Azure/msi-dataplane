all-test:
	@echo "Running all tests"
	@./hack/test.sh -a

generate:
	@echo "Generating code"
	@go generate ./...

integration-test:
	@echo "Running integration tests"
	@./hack/test.sh -i

integration-test-record:
	@echo "Running integration tests with recording"
	@./hack/test.sh -r

lint:
	@echo "Running linter"
	@golangci-lint run

tidy:
	@echo "Tidying up"
	@go mod tidy

unit-test:
	@echo "Running unit tests"
	@./hack/test.sh -u