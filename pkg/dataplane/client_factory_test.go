package dataplane

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type mockCredential struct {
	expectedOptions policy.TokenRequestOptions
}

func (m mockCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if diff := cmp.Diff(m.expectedOptions, options); diff != "" {
		return azcore.AccessToken{}, fmt.Errorf("unexpected options diff (-want +got):\n%s", diff)
	}
	return azcore.AccessToken{Token: "fake-token"}, nil
}

var _ azcore.TokenCredential = (*mockCredential)(nil)

type request struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}
type response struct {
	StatusCode int
	Body       []byte
}

func ptrTo[T any](v T) *T {
	return &v
}

func managedIdentityCredentials(delegatedResources []*DelegatedResource, explicitIdentities []*UserAssignedIdentityCredentials) ManagedIdentityCredentials {
	return ManagedIdentityCredentials{
		AuthenticationEndpoint: ptrTo("AuthenticationEndpoint"),
		CannotRenewAfter:       ptrTo("CannotRenewAfter"),
		ClientID:               ptrTo("ClientID"),
		ClientSecret:           ptrTo("ClientSecret"),
		ClientSecretURL:        ptrTo("ClientSecretURL"),
		CustomClaims:           ptrTo(customClaims()),
		DelegatedResources: func() []*DelegatedResource {
			if len(delegatedResources) > 0 {
				return delegatedResources
			}
			return nil
		}(),
		DelegationURL: ptrTo("DelegationURL"),
		ExplicitIdentities: func() []*UserAssignedIdentityCredentials {
			if len(explicitIdentities) > 0 {
				return explicitIdentities
			}
			return nil
		}(),
		InternalID:                 ptrTo("InternalID"),
		MtlsAuthenticationEndpoint: ptrTo("MtlsAuthenticationEndpoint"),
		NotAfter:                   ptrTo("NotAfter"),
		NotBefore:                  ptrTo("NotBefore"),
		ObjectID:                   ptrTo("ObjectID"),
		RenewAfter:                 ptrTo("RenewAfter"),
		TenantID:                   ptrTo("TenantID"),
	}
}

func delegatedResource(implicitIdentity *UserAssignedIdentityCredentials, explicitIdentities ...*UserAssignedIdentityCredentials) *DelegatedResource {
	return &DelegatedResource{
		DelegationID:  ptrTo("DelegationID"),
		DelegationURL: ptrTo("DelegationURL"),
		ExplicitIdentities: func() []*UserAssignedIdentityCredentials {
			if len(explicitIdentities) > 0 {
				return explicitIdentities
			}
			return nil
		}(),
		ImplicitIdentity: implicitIdentity,
		InternalID:       ptrTo("InternalID"),
		ResourceID:       ptrTo("ResourceID"),
	}
}

func userAssignedIdentityCredentials() *UserAssignedIdentityCredentials {
	return &UserAssignedIdentityCredentials{
		AuthenticationEndpoint:     ptrTo("AuthenticationEndpoint"),
		CannotRenewAfter:           ptrTo("CannotRenewAfter"),
		ClientID:                   ptrTo("ClientID"),
		ClientSecret:               ptrTo("ClientSecret"),
		ClientSecretURL:            ptrTo("ClientSecretURL"),
		CustomClaims:               ptrTo(customClaims()),
		MtlsAuthenticationEndpoint: ptrTo("MtlsAuthenticationEndpoint"),
		NotAfter:                   ptrTo("NotAfter"),
		NotBefore:                  ptrTo("NotBefore"),
		ObjectID:                   ptrTo("ObjectID"),
		RenewAfter:                 ptrTo("RenewAfter"),
		ResourceID:                 ptrTo("ResourceID"),
		TenantID:                   ptrTo("TenantID"),
	}
}

func customClaims() CustomClaims {
	return CustomClaims{
		XMSAzNwperimid: []*string{ptrTo("XMSAzNwperimid")},
		XMSAzTm:        ptrTo("XMSAzTm"),
	}
}

type transporter struct {
	http.RoundTripper
}

func (t *transporter) Do(req *http.Request) (*http.Response, error) {
	return t.RoundTripper.RoundTrip(req)
}

var _ policy.Transporter = (*transporter)(nil)

func TestClient(t *testing.T) {
	// the x-ms-identity-url header from ARM encodes a host, path, and query parameters
	// here, we need to use the host for our test server, but we can ensure that our
	// client correctly honors the host, path, and query passed to the constructor
	identityURLPath := "/subscriptions/a5d995f9-666e-40c6-953a-8a12c1010576/resourcegroups/resource-group/providers/Microsoft.Service/Objects/object/credentials/v2/identities"
	identityURLQuery := url.Values{
		"arpid":  {"d747ddeb-e7dd-4381-b79d-cfeb77e054a0"},
		"keyid":  {"60eae3c4-0a0a-4591-8f71-8110b40a3b50"},
		"sig":    {"RXJyb3I6IGdldDogbm90aGluZyB0byBkZWxldGUsIHR"},
		"sigver": {"1.0"},
		"tid":    {"8a1290f5-b9fc-4f74-87ac-9d5e98051efd"},
	}

	t.Log("setting up fake entra server")
	entraServer := httptest.NewTLSServer(nil) // TODO: expect challenge to send requests here
	defer entraServer.Close()
	entraServerURL, err := url.Parse(entraServer.URL)
	if err != nil {
		t.Fatalf("failed to parse entra server URL: %v", err)
	}
	entraServerURL.Path = identityURLQuery.Get("tid")

	mockCredentials := managedIdentityCredentials([]*DelegatedResource{
		delegatedResource(userAssignedIdentityCredentials(), userAssignedIdentityCredentials(), userAssignedIdentityCredentials()),
		delegatedResource(userAssignedIdentityCredentials(), userAssignedIdentityCredentials(), userAssignedIdentityCredentials()),
	}, []*UserAssignedIdentityCredentials{
		userAssignedIdentityCredentials(),
		userAssignedIdentityCredentials(),
	})
	encodedMockCredentials, err := json.Marshal(mockCredentials)
	if err != nil {
		t.Fatalf("failed to encode mock credentials: %v", err)
	}

	expectedQuery, err := url.ParseQuery(identityURLQuery.Encode())
	if err != nil {
		t.Fatalf("failed to parse query url: %v", err)
	}
	expectedQuery.Set("api-version", "2024-01-01")

	t.Log("setting up fake msi dataplane server")
	var expectedRequest request
	var expectedResponse response
	mutex := &sync.RWMutex{}
	msiDataplaneServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.RLock()
		defer mutex.RUnlock()
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer authorization="%s"`, entraServerURL.String()))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if diff := cmp.Diff(expectedRequest, request{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Body:   body,
		}); diff != "" {
			t.Errorf("unexpected request (-want +got):\n%s", diff)
		}

		w.WriteHeader(expectedResponse.StatusCode)
		if _, err := w.Write(expectedResponse.Body); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer msiDataplaneServer.Close()

	identityURL, err := url.Parse(msiDataplaneServer.URL)
	if err != nil {
		t.Fatalf("error parsing msi dataplane test server URL: %v", err)
	}
	identityURL.Path = identityURLPath
	identityURL.RawQuery = identityURLQuery.Encode()

	credential := &mockCredential{
		expectedOptions: policy.TokenRequestOptions{
			Claims:    "",
			EnableCAE: true,
			Scopes:    []string{"test-audience/.default"},
			TenantID:  "8a1290f5-b9fc-4f74-87ac-9d5e98051efd",
		},
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	factory := NewClientFactory(credential, "test-audience", &azcore.ClientOptions{
		Transport: &transporter{http.DefaultTransport},
	})
	msiClient, err := factory.NewClient(identityURL.String())
	if err != nil {
		t.Fatalf("error creating client: %v", err)
	}

	t.Run("get system assigned identity credentials", func(t *testing.T) {
		mutex.Lock()
		expectedRequest = request{
			Method: http.MethodGet,
			Path:   identityURL.Path,
			Query:  expectedQuery,
			Body:   []byte{},
		}
		expectedResponse = response{
			StatusCode: 200,
			Body:       encodedMockCredentials,
		}
		mutex.Unlock()
		credentials, err := msiClient.GetSystemAssignedIdentityCredentials(context.Background())
		if err != nil {
			t.Fatalf("error getting system assigned identity credentials: %v", err)
		}
		if diff := cmp.Diff(credentials, &mockCredentials); diff != "" {
			t.Fatalf("unexpected credentials (-want +got):\n%s", diff)
		}
	})

	t.Run("user assigned identities credentials", func(t *testing.T) {
		userAssignedRequest := UserAssignedIdentitiesRequest{
			CustomClaims:       ptrTo(customClaims()),
			DelegatedResources: []*string{ptrTo("first"), ptrTo("second")},
			IdentityIDs:        []*string{ptrTo("something"), ptrTo("else")},
		}
		encodedUserAssignedRequest, err := json.Marshal(userAssignedRequest)
		if err != nil {
			t.Fatalf("failed to encode user assigned identity request: %v", err)
		}
		t.Run("get without error", func(t *testing.T) {
			mutex.Lock()
			expectedRequest = request{
				Method: http.MethodPost,
				Path:   identityURL.Path,
				Query:  expectedQuery,
				Body:   encodedUserAssignedRequest,
			}
			expectedResponse = response{
				StatusCode: 200,
				Body:       encodedMockCredentials,
			}
			mutex.Unlock()
			userAssignedCredentials, err := msiClient.GetUserAssignedIdentitiesCredentials(context.Background(), userAssignedRequest)
			if err != nil {
				t.Errorf("error getting user assigned identity credentials: %v", err)
			}
			if diff := cmp.Diff(userAssignedCredentials, &mockCredentials); diff != "" {
				t.Errorf("unexpected credentials (-want +got):\n%s", diff)
			}
		})

		t.Run("get with empty error", func(t *testing.T) {
			mutex.Lock()
			expectedRequest = request{
				Method: http.MethodPost,
				Path:   identityURL.Path,
				Query:  expectedQuery,
				Body:   encodedUserAssignedRequest,
			}
			expectedResponse = response{
				StatusCode: 404,
				Body:       nil,
			}
			mutex.Unlock()
			userAssignedCredentials, err := msiClient.GetUserAssignedIdentitiesCredentials(context.Background(), userAssignedRequest)
			if err == nil {
				t.Error("expected an error getting user assigned identity credentials, got none")
			}
			if diff := cmp.Diff(userAssignedCredentials, &ManagedIdentityCredentials{}); diff != "" {
				t.Errorf("unexpected credentials (-want +got):\n%s", diff)
			}
		})

		t.Run("get with error body", func(t *testing.T) {
			mutex.Lock()
			expectedRequest = request{
				Method: http.MethodPost,
				Path:   identityURL.Path,
				Query:  expectedQuery,
				Body:   encodedUserAssignedRequest,
			}
			expectedResponse = response{
				StatusCode: 404,
				Body:       []byte(`{"code":"whatever"}`),
			}
			mutex.Unlock()
			userAssignedCredentials, err := msiClient.GetUserAssignedIdentitiesCredentials(context.Background(), userAssignedRequest)
			if err == nil {
				t.Error("expected an error getting user assigned identity credentials, got none")
			}
			expected := &azcore.ResponseError{}
			if !errors.As(err, &expected) {
				t.Errorf("expected error %T, got: %T", expected, err)
			}
			expected.RawResponse = nil
			if diff := cmp.Diff(expected, &azcore.ResponseError{
				StatusCode: 404,
				ErrorCode:  "whatever",
			}, cmpopts.IgnoreUnexported(azcore.ResponseError{})); diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(userAssignedCredentials, &ManagedIdentityCredentials{}); diff != "" {
				t.Errorf("unexpected credentials (-want +got):\n%s", diff)
			}
		})
	})

	t.Run("delete system assigned identity credentials", func(t *testing.T) {
		mutex.Lock()
		expectedRequest = request{
			Method: http.MethodDelete,
			Path:   identityURL.Path,
			Query:  expectedQuery,
			Body:   []byte{},
		}
		expectedResponse = response{
			StatusCode: 200,
			Body:       encodedMockCredentials,
		}
		mutex.Unlock()
		if err := msiClient.DeleteSystemAssignedIdentity(context.Background()); err != nil {
			t.Errorf("error deleting system assigned identity: %v", err)
		}
	})

	t.Run("move user assigned identity", func(t *testing.T) {
		moveRequest := MoveIdentityRequest{
			TargetResourceID: ptrTo("TargetResourceID"),
		}
		encodedMoveRequest, err := json.Marshal(moveRequest)
		if err != nil {
			t.Errorf("failed to encode move identity request: %v", err)
		}
		moveResponse := MoveIdentityResponse{
			IdentityURL: ptrTo("IdentityURL"),
		}
		encodedMoveResponse, err := json.Marshal(moveResponse)
		if err != nil {
			t.Errorf("failed to encode move identity response: %v", err)
		}
		mutex.Lock()
		expectedRequest = request{
			Method: http.MethodPost,
			Path:   identityURL.Path + "/proxy/move",
			Query:  expectedQuery,
			Body:   encodedMoveRequest,
		}
		expectedResponse = response{
			StatusCode: 200,
			Body:       encodedMoveResponse,
		}
		mutex.Unlock()
		movedIdentityResponse, err := msiClient.MoveIdentity(context.Background(), moveRequest)
		if err != nil {
			t.Errorf("error moving user assigned identity credentials: %v", err)
		}
		if diff := cmp.Diff(movedIdentityResponse, &moveResponse); diff != "" {
			t.Errorf("unexpected response (-want +got):\n%s", diff)
		}
	})
}
