include ./.bingo/Variables.mk

# ADO does not expose an unauthenticated API for fetching one file, and the git server
# does not have support for `git archive`, so we need to do some silly things to grab
# just the OpenAPIv2 specification we need instead of the 10GiB in the rest of the repo...
pkg/dataplane/internal/msi-credentials-data-plane.openapi.v2.json:
	git clone -n --depth=1 --filter=tree:0 https://msazure.visualstudio.com/One/_git/ManagedIdentity-MIRP
	cd ManagedIdentity-MIRP && git sparse-checkout set --no-cone src/Product/MSI/swagger/CredentialsDataPlane/2024-01-01/ && git checkout
	mv ManagedIdentity-MIRP/src/Product/MSI/swagger/CredentialsDataPlane/2024-01-01/msi-credentials-data-plane-2024-01-01.json $@
	rm -rf ManagedIdentity-MIRP

pkg/dataplane/internal/msi-credentials-data-plane.openapi.v3.yaml: pkg/dataplane/internal/msi-credentials-data-plane.openapi.v2.json
	docker run -d -p 8080:8080 --name swagger-converter swaggerapi/swagger-converter:latest
	sleep 2 # wait for server to spin up in the container, could be a poll to speed it up
	curl -s -H 'Accept: application/yaml' -H 'Content-Type: application/json' -X POST --data @pkg/dataplane/internal/msi-credentials-data-plane.openapi.v2.json localhost:8080/api/convert > $@
	docker stop swagger-converter && docker rm swagger-converter

pkg/dataplane/internal/generated_client.go: $(OAPI_CODEGEN) pkg/dataplane/internal/msi-credentials-data-plane.openapi.v3.yaml
	 $(OAPI_CODEGEN) --generate client,models --package internal pkg/dataplane/internal/msi-credentials-data-plane.openapi.v3.yaml > $@

test:
	@echo "Running all tests"
	go test ./...

generate: pkg/dataplane/internal/generated_client.go

lint: $(GOLANGCI_LINT)
	@echo "Running linter"
	$(GOLANGCI_LINT) run

tidy:
	@echo "Tidying up"
	go mod tidy

fmt: $(OPENSHIFT_GOIMPORTS)
	$(OPENSHIFT_GOIMPORTS) --module github.com/Azure/msi-dataplane

verify: fmt lint tidy test generate
	if ! git diff --quiet HEAD; then \
		git diff; \
		echo "You need to run 'make generate' to update generated files and commit them"; \
		exit 1; \
	fi

_antlr-docker-image:
	mkdir -p /tmp/antlr4
	git clone https://github.com/antlr/antlr4.git --depth 1 /tmp/antlr4
	cd /tmp/antlr4/docker && docker build -t antlr/antlr4 .
	docker inspect antlr/antlr4 > $@

pkg/dataplane/internal/challenge/challenge_parser.go: _antlr-docker-image
	docker run --rm -v $(PWD)/$(dir $@):/work:Z antlr/antlr4 -Dlanguage=Go -package challenge Challenge.g4