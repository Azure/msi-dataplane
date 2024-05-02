module github.com/Azure/msi-dataplane

go 1.22

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.11.1
	go.uber.org/mock v0.4.0
)

require github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.0.0 // indirect

require (
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets v1.1.0
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
