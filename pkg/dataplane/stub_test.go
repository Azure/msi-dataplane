package dataplane

import (
	"context"
	"reflect"
	"testing"

	"github.com/Azure/msi-dataplane/internal/swagger"
	"github.com/Azure/msi-dataplane/internal/test"
)

func TestNewStubClient(t *testing.T) {
	t.Parallel()
	uaMSI := test.GetTestMSI(test.ValidResourceID)
	credObject := &CredentialsObject{
		swagger.CredentialsObject{
			ExplicitIdentities: []*swagger.NestedCredentialsObject{uaMSI},
		},
	}
	testStub := NewStub([]*CredentialsObject{credObject})
	if testStub == nil {
		t.Error("expected non-nil stub")
	}
}

func TestPost(t *testing.T) {
	t.Parallel()
	uaMSI := test.GetTestMSI(test.ValidResourceID)
	credObject := &CredentialsObject{
		swagger.CredentialsObject{
			ExplicitIdentities: []*swagger.NestedCredentialsObject{uaMSI},
		},
	}
	testStub := NewStub([]*CredentialsObject{credObject})
	client, err := NewStubClient(AzurePublicCloud, testStub)
	if err != nil {
		t.Fatalf("unable to create stub client: %s", err)
	}
	request := UserAssignedMSIRequest{
		ResourceIDs: []string{test.ValidResourceID},
		IdentityURL: test.ValidIdentityURL,
		TenantID:    test.ValidTenantID,
	}
	identities, err := client.GetUserAssignedIdentities(context.Background(), request)
	if err != nil {
		t.Fatalf("unable to get user assigned msi: %s", err)
	}

	if len(identities.ExplicitIdentities) != 1 {
		t.Errorf("expected 1 identity but got %d", len(identities.ExplicitIdentities))
	}

	if !reflect.DeepEqual(identities.ExplicitIdentities[0], uaMSI) {
		t.Errorf("returned identity does not match expected identity")
	}
}
