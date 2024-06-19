package dataplane

import (
	"errors"
	"reflect"
	"testing"

	azcloud "github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/msi-dataplane/internal/swagger"
	"github.com/Azure/msi-dataplane/internal/test"
)

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
