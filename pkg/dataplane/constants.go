package dataplane

type contextKey string

const (
	IdentityURLKey contextKey = msiIdentityURLHeader

	// Cloud Environments
	AzurePublicCloud = "AZUREPUBLICCLOUD"
	AzureUSGovCloud  = "AZUREUSGOVERNMENTCLOUD"

	// MSI Headers
	msiIdentityURLHeader  = "x-ms-identity-url"
	apiVersionParameter   = "api-version"
	headerAuthorization   = "authorization"
	headerWWWAuthenticate = "WWW-Authenticate"

	// MSI Endpoints sub domain
	publicMSIEndpoint = "identity.azure.net"
	usGovMSIEndpoint  = "identity.usgovcloudapi.net"

	https = "https"
)
