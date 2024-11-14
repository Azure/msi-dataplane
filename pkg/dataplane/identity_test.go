//go:build unit

package dataplane

import (
	"errors"
	"reflect"
	"testing"

	azcloud "github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/msi-dataplane/internal/test"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
)

func TestIsUserAssigned(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		credenialsObject swagger.CredentialsObject
		expected         bool
	}{
		{
			name: "User assigned MSI",
			credenialsObject: swagger.CredentialsObject{
				ExplicitIdentities: []*swagger.NestedCredentialsObject{
					test.GetTestMSI(test.ValidResourceID),
				},
			},
			expected: true,
		},
		{
			name:             "System assigned MSI",
			credenialsObject: swagger.CredentialsObject{},
			expected:         false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			msi := CredentialsObject{
				CredentialsObject: tc.credenialsObject,
			}

			if result := msi.IsUserAssigned(); result != tc.expected {
				t.Errorf("expected: `%t` but got: `%t`", tc.expected, result)
			}
		})
	}
}

func TestNewUserAssignedIdentities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		c           CredentialsObject
		expectedErr error
	}{
		{
			name: "No user-assigned managed identities",
			c: CredentialsObject{
				CredentialsObject: swagger.CredentialsObject{
					ExplicitIdentities: []*swagger.NestedCredentialsObject{},
				},
			},
			expectedErr: errNoUserAssignedMSIs,
		},
		{
			name: "User-assigned managed identities present",
			c: CredentialsObject{
				CredentialsObject: swagger.CredentialsObject{
					ExplicitIdentities: []*swagger.NestedCredentialsObject{
						test.GetTestMSI(test.Bogus),
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewUserAssignedIdentities(tc.c, test.Bogus); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}
}

func TestGetCredential(t *testing.T) {
	t.Parallel()

	validIdentity := test.GetTestMSI(test.ValidResourceID)
	validIdentity.ClientSecret = test.StringPtr(test.MockClientSecret)
	validIdentity.TenantID = test.StringPtr(test.ValidTenantID)
	validIdentity.AuthenticationEndpoint = test.StringPtr(test.ValidAuthenticationEndpoint)

	testCases := []struct {
		name         string
		uaIdentities UserAssignedIdentities
		resourceID   string
		expectedErr  error
	}{
		{
			name: "empty resourceID",
			uaIdentities: UserAssignedIdentities{
				CredentialsObject: CredentialsObject{
					CredentialsObject: swagger.CredentialsObject{
						ExplicitIdentities: []*swagger.NestedCredentialsObject{
							test.GetTestMSI(test.ValidResourceID),
						},
					},
				},
			},
			resourceID:  "",
			expectedErr: errResourceIDNotFound,
		},
		{
			name:         "no identities present",
			uaIdentities: UserAssignedIdentities{},
			resourceID:   test.ValidResourceID,
			expectedErr:  errResourceIDNotFound,
		},
		{
			name: "Invalid client secret",
			uaIdentities: UserAssignedIdentities{
				CredentialsObject: CredentialsObject{
					CredentialsObject: swagger.CredentialsObject{
						ExplicitIdentities: []*swagger.NestedCredentialsObject{
							test.GetTestMSI(test.ValidResourceID),
						},
					},
				},
			},
			resourceID:  test.ValidResourceID,
			expectedErr: errDecodeClientSecret,
		},
		{
			name: "success",
			uaIdentities: UserAssignedIdentities{
				CredentialsObject: CredentialsObject{
					CredentialsObject: swagger.CredentialsObject{
						ExplicitIdentities: []*swagger.NestedCredentialsObject{
							validIdentity,
						},
					},
				},
			},
			resourceID:  test.ValidResourceID,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := tc.uaIdentities.GetCredential(tc.resourceID); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: `%s` but got: `%s`", tc.expectedErr, err)
			}
		})
	}
}

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
				AuthenticationEndpoint: test.StringPtr(test.ValidAuthenticationEndpoint),
				ClientID:               test.StringPtr(test.Bogus),
				ClientSecret:           test.StringPtr(test.MockClientSecret),
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
				testMSI := test.GetTestMSI(test.Bogus)
				resourceID := test.ValidResourceID + "-bogus"
				testMSI.ResourceID = test.StringPtr(resourceID)
				return []*swagger.NestedCredentialsObject{testMSI}
			},
			resourceIDs: []string{test.ValidResourceID},
			expectedErr: errResourceIDNotFound,
		},
		{
			name: "success",
			getMSI: func() []*swagger.NestedCredentialsObject {
				testMSI := test.GetTestMSI(test.ValidResourceID)
				return []*swagger.NestedCredentialsObject{testMSI}
			},
			resourceIDs: []string{test.ValidResourceID},
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
