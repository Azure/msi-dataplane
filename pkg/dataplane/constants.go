package dataplane

type contextKey string

const (
	IdentityURLKey contextKey = MSIIdentityURLHeader

	// MSI Headers
	MSIIdentityURLHeader  = "x-ms-identity-url"
	apiVersionParameter   = "api-version"
	HeaderAuthorization   = "authorization"
	HeaderWWWAuthenticate = "WWW-Authenticate"

	// MSI Endpoints sub domain
	publicMSIEndpoint = "identity.azure.net"
	usGovMSIEndpoint  = "identity.usgovcloudapi.net"

	// Cloud Environments
	AzurePublicCloud = "AZUREPUBLICCLOUD"
	AzureUSGovCloud  = "AZUREUSGOVERNMENTCLOUD"

	https = "https"
)
