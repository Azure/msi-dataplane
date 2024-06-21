//go:build unit

package dataplane

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/Azure/msi-dataplane/internal/swagger"
	"github.com/Azure/msi-dataplane/internal/test"
)

func TestNewStub(t *testing.T) {
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

func TestDo(t *testing.T) {
	t.Parallel()
	uaMSI := test.GetTestMSI(test.ValidResourceID)
	credObject := &CredentialsObject{
		swagger.CredentialsObject{
			ExplicitIdentities: []*swagger.NestedCredentialsObject{uaMSI},
		},
	}
	credRequest := &swagger.CredRequestDefinition{
		IdentityIDs: []*string{test.StringPtr(test.ValidResourceID)},
	}
	credRequestBytes, err := credRequest.MarshalJSON()
	if err != nil {
		t.Fatalf("unable to marshal request: %s", err)
	}
	credRequestBody := bytes.NewBuffer(credRequestBytes)

	testCases := []struct {
		name               string
		body               *bytes.Buffer
		method             string
		expectedErr        error
		expectedStatusCode int
	}{
		{
			name:               "Post",
			body:               credRequestBody,
			method:             http.MethodPost,
			expectedErr:        nil,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Get",
			body:               bytes.NewBufferString(""),
			method:             http.MethodGet,
			expectedErr:        nil,
			expectedStatusCode: http.StatusNotImplemented,
		},
		{
			name:               "Put",
			body:               bytes.NewBufferString(""),
			method:             http.MethodPut,
			expectedErr:        nil,
			expectedStatusCode: http.StatusNotImplemented,
		},
		{
			name:               "Delete",
			body:               bytes.NewBufferString(""),
			method:             http.MethodDelete,
			expectedErr:        nil,
			expectedStatusCode: http.StatusNotImplemented,
		},
		{
			name:               "Patch",
			body:               bytes.NewBufferString(""),
			method:             http.MethodPatch,
			expectedErr:        nil,
			expectedStatusCode: http.StatusNotImplemented,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req, err := http.NewRequest(tc.method, "http://localhost", io.NopCloser(tc.body))
			if err != nil {
				t.Fatalf("unable to create request: %s", err)
			}
			testStub := NewStub([]*CredentialsObject{credObject})
			response, err := testStub.Do(req)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %s but got %s", tc.expectedErr, err)
			}
			if response.StatusCode != tc.expectedStatusCode {
				t.Errorf("expected status code %d but got %d", tc.expectedStatusCode, response.StatusCode)
			}
		})
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

	testCases := []struct {
		name               string
		body               io.Reader
		expectedStatusCode int
		expectedErr        error
	}{
		{
			name:               "No body",
			body:               http.NoBody,
			expectedStatusCode: http.StatusBadRequest,
			expectedErr:        errStubRequestBody,
		},
		{
			name:               "Non conforming body",
			body:               bytes.NewBufferString(test.Bogus),
			expectedStatusCode: http.StatusBadRequest,
			expectedErr:        errStubRequestBody,
		},
		{
			name:               "ResourceID not found",
			body:               bytes.NewBufferString(`{"identityIds": ["bogus"]}`),
			expectedStatusCode: http.StatusNotFound,
			expectedErr:        nil,
		},
		{
			name:               "Success",
			body:               bytes.NewBufferString(`{"identityIds": ["` + test.ValidResourceID + `"]}`),
			expectedStatusCode: http.StatusOK,
			expectedErr:        nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testStub := NewStub([]*CredentialsObject{credObject})
			req, err := http.NewRequest(http.MethodPost, "http://localhost", io.NopCloser(tc.body))
			if err != nil {
				t.Fatalf("unable to create request: %s", err)
			}
			response, err := testStub.post(req)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %s but got %s", tc.expectedErr, err)
			}
			if response.StatusCode != tc.expectedStatusCode {
				t.Errorf("expected status code %d but got %d", tc.expectedStatusCode, response.StatusCode)
			}
		})
	}

}

func TestStubWithClient(t *testing.T) {
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
