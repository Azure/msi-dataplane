//go:build unit

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
	authenticator := NewAuthenticator(cred, cloud)

	// Create a new client
	client, err := NewClient(aud, authenticator, nil)
	if err != nil {
		t.Fatalf("Failed to create a new client: %s", err)
	}

	// Check if the client is not nil
	if client == nil {
		t.Fatalf("Client is nil")
	}
}

func TestGetUserAssignedIdentities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		goMockCall  func(swaggerClient *mock.MockswaggerMSIClient)
		request     UserAssignedMSIRequest
		expectedErr error
	}{
		{
			name:        "IdenityURL not specified",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{ResourceIDs: []string{test.ValidResourceID}, TenantID: test.ValidTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "IdenityURL not a URL",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: test.Bogus, ResourceIDs: []string{test.ValidResourceID}, TenantID: test.ValidTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "ResourceID not specified",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, TenantID: test.ValidTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "ResourceID not valid",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.Bogus}, TenantID: test.ValidTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not specified",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.ValidResourceID}},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not a UUID",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.ValidResourceID}, TenantID: test.Bogus},
			expectedErr: errInvalidRequest,
		},
		{
			name: "Swagger client error",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, errors.New(test.Bogus))
			},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.ValidResourceID}, TenantID: test.ValidTenantID},
			expectedErr: errGetCreds,
		},
		{
			name: "Zero MSIs returned",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.ValidResourceID}, TenantID: test.ValidTenantID},
			expectedErr: errNumberOfMSIs,
		},
		{
			name: "Mismatched number of MSIs",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				uaMSI := test.GetTestMSI(test.Bogus)
				identities := []*swagger.NestedCredentialsObject{uaMSI}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.ValidResourceID, test.ValidResourceID}, TenantID: test.ValidTenantID},
			expectedErr: errNumberOfMSIs,
		},
		{
			name: "Valid request - single MSI",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				resourceID := test.ValidResourceID
				tenantID := test.ValidTenantID

				uaMSI := test.GetTestMSI(test.Bogus)
				uaMSI.ResourceID = &resourceID
				uaMSI.TenantID = &tenantID

				identities := []*swagger.NestedCredentialsObject{uaMSI}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.ValidResourceID}, TenantID: test.ValidTenantID},
			expectedErr: nil,
		},
		{
			name: "Valid request - multiple MSIs",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				resourceID := test.ValidResourceID
				tenantID := test.ValidTenantID

				uaMSI := test.GetTestMSI(test.Bogus)
				uaMSI.ResourceID = &resourceID
				uaMSI.TenantID = &tenantID

				identities := []*swagger.NestedCredentialsObject{uaMSI, uaMSI}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: test.ValidIdentityURL, ResourceIDs: []string{test.ValidResourceID, test.ValidResourceID}, TenantID: test.ValidTenantID},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			swaggerClient := mock.NewMockswaggerMSIClient(mockCtrl)
			tc.goMockCall(swaggerClient)

			msiClient := &ManagedIdentityClient{swaggerClient: swaggerClient}
			if _, err := msiClient.GetUserAssignedIdentities(context.Background(), tc.request); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}
}
