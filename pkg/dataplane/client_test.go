package dataplane

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/Azure/msi-dataplane/internal/swagger"
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

	const validIdentityURL = "https://bogus.com"
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
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not specified",
			goMockCall:  func() {},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceID: validResourceID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not a UUID",
			goMockCall:  func() {},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceID: validResourceID, TenantID: "bogus"},
			expectedErr: errInvalidRequest,
		},
		{
			name: "Zero MSIs returned",
			goMockCall: func() {
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceID: validResourceID, TenantID: validTenantID},
			expectedErr: errExpectedOneMSI,
		},
		{
			name: "Multiple MSIs returned",
			goMockCall: func() {
				identities := []*swagger.NestedCredentialsObject{nil, nil}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceID: validResourceID, TenantID: validTenantID},
			expectedErr: errExpectedOneMSI,
		},
		{
			name: "MSI is nil",
			goMockCall: func() {
				identities := []*swagger.NestedCredentialsObject{nil}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceID: validResourceID, TenantID: validTenantID},
			expectedErr: errExpectedNonNilMSI,
		},
		{
			name: "ResourceID mismatch",
			goMockCall: func() {
				bogusResourceID := "bogus"
				uaMSI := getTestUaMSI("bogus")
				uaMSI.ResourceID = &bogusResourceID

				identities := []*swagger.NestedCredentialsObject{&uaMSI}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceID: validResourceID, TenantID: validTenantID},
			expectedErr: errResourceIDMismatch,
		},
		{
			name: "Valid request",
			goMockCall: func() {
				resourceID := validResourceID
				tenantID := validTenantID

				uaMSI := getTestUaMSI("bogus")
				uaMSI.ResourceID = &resourceID
				uaMSI.TenantID = &tenantID

				identities := []*swagger.NestedCredentialsObject{&uaMSI}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceID: validResourceID, TenantID: validTenantID},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.goMockCall()
			if _, err := msiClient.GetUserAssignedMSI(context.Background(), tc.request); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}
}

func getTestUaMSI(placeHolder string) swagger.NestedCredentialsObject {
	return swagger.NestedCredentialsObject{
		AuthenticationEndpoint:     &placeHolder,
		CannotRenewAfter:           &placeHolder,
		ClientID:                   &placeHolder,
		ClientSecret:               &placeHolder,
		ClientSecretURL:            &placeHolder,
		CustomClaims:               &swagger.CustomClaims{},
		MtlsAuthenticationEndpoint: &placeHolder,
		NotAfter:                   &placeHolder,
		NotBefore:                  &placeHolder,
		ObjectID:                   &placeHolder,
		RenewAfter:                 &placeHolder,
		ResourceID:                 &placeHolder,
		TenantID:                   &placeHolder,
	}
}
