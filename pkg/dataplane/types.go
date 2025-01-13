package dataplane

import "github.com/Azure/msi-dataplane/pkg/dataplane/internal"

type ManagedIdentityCredentials = internal.CredentialsObject
type UserAssignedIdentityCredentials = internal.NestedCredentialsObject

type UserAssignedIdentitiesRequest = internal.CredRequestDefinition

type MoveIdentityRequest = internal.MoveRequestBodyDefinition
type MoveIdentityResponse = internal.MoveIdentityResponse

// ResponseError adapts the generated response error into something implementing a Go error,
// while exposing the internals so that upstream users can use errors.As to inspect the values.
type ResponseError struct {
	WrappedError internal.ErrorResponse
}

func (e *ResponseError) Error() string {
	if e.WrappedError.Error.Message != nil {
		return *e.WrappedError.Error.Message
	}
	return "An unknown error occurred."
}
