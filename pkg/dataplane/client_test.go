//go:build unit

package dataplane

import (
	"testing"

	"github.com/Azure/msi-dataplane/internal/test"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	aud := "aud"
	cloud := AzurePublicCloud
	cred := &test.FakeCredential{}

	// Create a new client
	client, err := NewClient(aud, cloud, cred)
	if err != nil {
		t.Fatalf("Failed to create a new client: %v", err)
	}

	// Check if the client is not nil
	if client == nil {
		t.Fatalf("Client is nil")
	}
}
