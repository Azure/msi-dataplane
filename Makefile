all-test:
	@./hack/test.sh -a

generate:
	@echo "Generating code"
	@go generate ./...

integration-test:
	@./hack/test.sh -i

integration-test-record:
	@./hack/test.sh -r

tidy:
	@echo "Tidying up"
	@go mod tidy

unit-test:
	@./hack/test.sh -u