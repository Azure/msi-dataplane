package dataplane

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/Azure/msi-dataplane/internal/test"
	mock "github.com/Azure/msi-dataplane/pkg/dataplane/mock_swagger_client"
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

func TestGetUserAssignedMSI(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	swaggerClient := mock.NewMockswaggerMSIClient(mockCtrl)
	msiClient := &ManagedIdentityClient{swaggerClient: swaggerClient}

	const validTenantID = "00000000-0000-0000-0000-000000000000"
	const validResourceID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/msi"

	testCases := []struct {
		name        string
		goMockCall  func()
		request     UserAssignedMSIRequest
		expectedErr error
	}{
		{
			name:        "IdenityURL not specified",
			goMockCall:  func() {},
			request:     UserAssignedMSIRequest{ResourceID: validResourceID, TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "IdenityURL not a URL",
			goMockCall:  func() {},
			request:     UserAssignedMSIRequest{IdentityURL: "bogus", ResourceID: validResourceID, TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "ResourceID not specified",
			goMockCall:  func() {},
			request:     UserAssignedMSIRequest{IdentityURL: "https://bogus.com", TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not specified",
			goMockCall:  func() {},
			request:     UserAssignedMSIRequest{IdentityURL: "https://bogus.com", ResourceID: validResourceID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not a UUID",
			goMockCall:  func() {},
			request:     UserAssignedMSIRequest{IdentityURL: "https://bogus.com", ResourceID: validResourceID, TenantID: "bogus"},
			expectedErr: errInvalidRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.goMockCall()
			if _, err := msiClient.GetUserAssignedMSI(context.Background(), tc.request); !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %s but got: %s", tc.expectedErr, err)
			}
		})
	}
}
