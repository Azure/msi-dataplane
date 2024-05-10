all-test:
	@echo "Running all tests"
	@./hack/test.sh -a

generate:
	@echo "Generating code"
	@go generate ./...

integration-test:
	@./hack/test.sh -i

integration-test-record:
	@./hack/test.sh -r

lint:
	@echo "Running linter"
	@golangci-lint run

tidy:
	@echo "Tidying up"
	@go mod tidy

unit-test:
	@./hack/test.sh -u