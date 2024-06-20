package dataplane

import (
	"errors"
	"reflect"
	"testing"

	azcloud "github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/msi-dataplane/internal/swagger"
	"github.com/Azure/msi-dataplane/internal/test"
)

func TestGetClientCertificateCredential(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		identity    swagger.NestedCredentialsObject
		cloud       string
		expectedErr error
	}{
		{
			name: "nil clientID",
			identity: swagger.NestedCredentialsObject{
				AuthenticationEndpoint: test.StringPtr(test.Bogus),
				ClientSecret:           test.StringPtr(test.Bogus),
				ClientID:               nil,
				TenantID:               test.StringPtr(test.Bogus),
			},
			cloud:       AzurePublicCloud,
			expectedErr: errNilField,
		},
		{
			name: "nil tenantID",
			identity: swagger.NestedCredentialsObject{
				AuthenticationEndpoint: test.StringPtr(test.Bogus),
				ClientID:               test.StringPtr(test.Bogus),
				ClientSecret:           test.StringPtr(test.Bogus),
				TenantID:               nil,
			},
			cloud:       AzurePublicCloud,
			expectedErr: errNilField,
		},
		{
			name: "nil clientSecret",
			identity: swagger.NestedCredentialsObject{
				AuthenticationEndpoint: test.StringPtr(test.Bogus),
				ClientID:               test.StringPtr(test.Bogus),
				ClientSecret:           nil,
				TenantID:               test.StringPtr(test.Bogus),
			},
			cloud:       AzurePublicCloud,
			expectedErr: errNilField,
		},
		{
			name: "nil authenticationEndpoint",
			identity: swagger.NestedCredentialsObject{
				AuthenticationEndpoint: nil,
				ClientID:               test.StringPtr(test.Bogus),
				ClientSecret:           test.StringPtr(test.Bogus),
				TenantID:               test.StringPtr(test.Bogus),
			},
			cloud:       AzurePublicCloud,
			expectedErr: errNilField,
		},
		{
			name: "invalid client secret causes failure to decode",
			identity: swagger.NestedCredentialsObject{
				AuthenticationEndpoint: test.StringPtr(test.Bogus),
				ClientID:               test.StringPtr(test.Bogus),
				ClientSecret:           test.StringPtr(test.Bogus),
				TenantID:               test.StringPtr(test.Bogus),
			},
			cloud:       AzurePublicCloud,
			expectedErr: errDecodeClientSecret,
		},
		{
			name: "empty client secret causes failure to parse",
			identity: swagger.NestedCredentialsObject{
				AuthenticationEndpoint: test.StringPtr(test.Bogus),
				ClientID:               test.StringPtr(test.Bogus),
				ClientSecret:           test.StringPtr(""),
				TenantID:               test.StringPtr(test.Bogus),
			},
			cloud:       AzurePublicCloud,
			expectedErr: errParseCertificate,
		},
		{
			name: "success",
			identity: swagger.NestedCredentialsObject{
				AuthenticationEndpoint: test.StringPtr("https://login.microsoftonline.com/"),
				ClientID:               test.StringPtr(test.Bogus),
				ClientSecret:           test.StringPtr(test.MockCertificate),
				TenantID:               test.StringPtr(test.Bogus),
			},
			cloud:       AzurePublicCloud,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := getClientCertificateCredential(tc.identity, tc.cloud); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}

}

func TestValidateUserAssignedMSI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		getMSI      func() []*swagger.NestedCredentialsObject
		resourceIDs []string
		expectedErr error
	}{
		{
			name:        "nil credential object slice",
			getMSI:      func() []*swagger.NestedCredentialsObject { return nil },
			resourceIDs: []string{"someResourceID"},
			expectedErr: errNumberOfMSIs,
		},
		{
			name: "nil fields",
			getMSI: func() []*swagger.NestedCredentialsObject {
				return []*swagger.NestedCredentialsObject{&swagger.NestedCredentialsObject{}}
			},
			resourceIDs: []string{"someResourceID"},
			expectedErr: errNilField,
		},
		{
			name: "mismatched resourceID",
			getMSI: func() []*swagger.NestedCredentialsObject {
				testMSI := test.GetTestMSI("bogus")
				return []*swagger.NestedCredentialsObject{testMSI}
			},
			resourceIDs: []string{"someResourceID"},
			expectedErr: errResourceIDNotFound,
		},
		{
			name: "success",
			getMSI: func() []*swagger.NestedCredentialsObject {
				testMSI := test.GetTestMSI(validResourceID)
				return []*swagger.NestedCredentialsObject{testMSI}
			},
			resourceIDs: []string{validResourceID},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			msi := tc.getMSI()
			if err := validateUserAssignedMSIs(msi, tc.resourceIDs); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}
}

func TestGetAzCoreCloud(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		azureEnv       string
		expectedResult azcloud.Configuration
	}{
		{
			name:           "AzurePublicCloud",
			azureEnv:       AzurePublicCloud,
			expectedResult: azcloud.AzurePublic,
		},
		{
			name:           "AzureUSGovernmentCloud",
			azureEnv:       AzureUSGovCloud,
			expectedResult: azcloud.AzureGovernment,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := getAzCoreCloud(tc.azureEnv)
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("expected: `%s` but got: `%s`", tc.expectedResult, result)
			}
		})
	}
}
