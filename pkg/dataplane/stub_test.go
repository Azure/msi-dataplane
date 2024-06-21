package dataplane

import (
	"context"
	"testing"
)

func TestStub(t *testing.T) {
	stub, err := NewStub(AzurePublicCloud)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	identityURL := "https://bogus.identity.azure.net"

	_, err = stub.GetUserAssignedIdentities(context.Background(), UserAssignedMSIRequest{IdentityURL: identityURL})
	t.Fatalf("unexpected error: %s", err)
}
