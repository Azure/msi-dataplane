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

const (
	validIdentityURL = "https://bogus.com"
	validTenantID    = "00000000-0000-0000-0000-000000000000"
	validResourceID  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/msi"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	aud := "aud"
	cloud := AzurePublicCloud
	cred := &test.FakeCredential{}

	// Create a new client
	client, err := NewClient(aud, cloud, cred)
	if err != nil {
		t.Fatalf("Failed to create a new client: %s", err)
	}

	// Check if the client is not nil
	if client == nil {
		t.Fatalf("Client is nil")
	}
}

func TestGetUserAssignedMSI(t *testing.T) {
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
			request:     UserAssignedMSIRequest{ResourceIDs: []string{validResourceID}, TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "IdenityURL not a URL",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: "bogus", ResourceIDs: []string{validResourceID}, TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "ResourceID not specified",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "ResourceID not valid",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceIDs: []string{"bogus"}, TenantID: validTenantID},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not specified",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceIDs: []string{validResourceID}},
			expectedErr: errInvalidRequest,
		},
		{
			name:        "TenantID not a UUID",
			goMockCall:  func(swaggerClient *mock.MockswaggerMSIClient) {},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceIDs: []string{validResourceID}, TenantID: "bogus"},
			expectedErr: errInvalidRequest,
		},
		{
			name: "Swagger client error",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, errors.New("bogus"))
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceIDs: []string{validResourceID}, TenantID: validTenantID},
			expectedErr: errGetCreds,
		},
		{
			name: "Zero MSIs returned",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceIDs: []string{validResourceID}, TenantID: validTenantID},
			expectedErr: errNumberOfMSIs,
		},
		{
			name: "Valid request - single MSI",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				resourceID := validResourceID
				tenantID := validTenantID

				uaMSI := getTestMSI("bogus")
				uaMSI.ResourceID = &resourceID
				uaMSI.TenantID = &tenantID

				identities := []*swagger.NestedCredentialsObject{uaMSI}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceIDs: []string{validResourceID}, TenantID: validTenantID},
			expectedErr: nil,
		},
		{
			name: "Valid request - multiple MSIs",
			goMockCall: func(swaggerClient *mock.MockswaggerMSIClient) {
				resourceID := validResourceID
				tenantID := validTenantID

				uaMSI := getTestMSI("bogus")
				uaMSI.ResourceID = &resourceID
				uaMSI.TenantID = &tenantID

				identities := []*swagger.NestedCredentialsObject{uaMSI, uaMSI}
				swaggerClient.EXPECT().Getcreds(gomock.Any(), gomock.Any(), gomock.Any()).Return(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse{
					CredentialsObject: swagger.CredentialsObject{ExplicitIdentities: identities},
				}, nil)
			},
			request:     UserAssignedMSIRequest{IdentityURL: validIdentityURL, ResourceIDs: []string{validResourceID, validResourceID}, TenantID: validTenantID},
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
			if _, err := msiClient.GetUserAssignedMSI(context.Background(), tc.request); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}
}

/*
func TestValidateUserAssignedMSIs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		getMSI      func() *swagger.NestedCredentialsObject
		resourceID  string
		expectedErr error
	}{
		{
			name:        "nil identity",
			getMSI:      func() *swagger.NestedCredentialsObject { return nil },
			resourceID:  "someResourceID",
			expectedErr: errNilMSI,
		},
		{
			name:        "nil fields",
			getMSI:      func() *swagger.NestedCredentialsObject { return &swagger.NestedCredentialsObject{} },
			resourceID:  "someResourceID",
			expectedErr: errNilField,
		},
		{
			name: "mismatched resourceID",
			getMSI: func() *swagger.NestedCredentialsObject {
				return getTestMSI("bogus")
			},
			resourceID:  "someResourceID",
			expectedErr: errResourceIDMismatch,
		},
		{
			name:        "success",
			getMSI:      func() *swagger.NestedCredentialsObject { return getTestMSI(validResourceID) },
			resourceID:  validResourceID,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			msi := tc.getMSI()
			if err := validateUserAssignedMSIs(msi, tc.resourceID); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}
}
*/

func getTestMSI(placeHolder string) *swagger.NestedCredentialsObject {
	return &swagger.NestedCredentialsObject{
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
