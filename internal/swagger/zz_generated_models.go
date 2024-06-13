//go:build go1.18
// +build go1.18

// Code generated by Microsoft (R) AutoRest Code Generator (autorest: 3.10.2, generator: @autorest/go@4.0.0-preview.63)
// Changes may cause incorrect behavior and will be lost if the code is regenerated.
// Code generated by @autorest/go. DO NOT EDIT.

package swagger

type CredRequestDefinition struct {
	CustomClaims       *CustomClaims
	DelegatedResources []*string
	IdentityIDs        []*string
}

type CredentialsObject struct {
	AuthenticationEndpoint     *string
	CannotRenewAfter           *string
	ClientID                   *string
	ClientSecret               *string
	ClientSecretURL            *string
	CustomClaims               *CustomClaims
	DelegatedResources         []*DelegatedResourceObject
	DelegationURL              *string
	ExplicitIdentities         []*NestedCredentialsObject
	InternalID                 *string
	MtlsAuthenticationEndpoint *string
	NotAfter                   *string
	NotBefore                  *string
	ObjectID                   *string
	RenewAfter                 *string
	TenantID                   *string
}

type CustomClaims struct {
	XMSAzNwperimid []*string
	XMSAzTm        *string
}

type DelegatedResourceObject struct {
	DelegationID       *string
	DelegationURL      *string
	ExplicitIdentities []*NestedCredentialsObject
	ImplicitIdentity   *NestedCredentialsObject
	InternalID         *string
	ResourceID         *string
}

type ErrorResponse struct {
	Error *ErrorResponseError
}

type ErrorResponseError struct {
	Code    *string
	Message *string
}

type MoveIdentityResponse struct {
	IdentityURL *string
}

type MoveRequestBodyDefinition struct {
	TargetResourceID *string
}

type NestedCredentialsObject struct {
	AuthenticationEndpoint     *string
	CannotRenewAfter           *string
	ClientID                   *string
	ClientSecret               *string
	ClientSecretURL            *string
	CustomClaims               *CustomClaims
	MtlsAuthenticationEndpoint *string
	NotAfter                   *string
	NotBefore                  *string
	ObjectID                   *string
	RenewAfter                 *string
	ResourceID                 *string
	TenantID                   *string
}