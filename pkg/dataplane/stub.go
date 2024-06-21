package dataplane

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/msi-dataplane/internal/swagger"
	"github.com/Azure/msi-dataplane/internal/test"
)

// TODO - add support for system-assigned managed identities
type stubTransporter struct {
}

func (tp stubTransporter) Do(req *http.Request) (*http.Response, error) {
	fmt.Printf("Request: %v\n", *req)

	// Read the request body
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err // Handle error
	}
	// Replace the request body so it can be read again
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// Convert the body to a string and print
	bodyString := string(bodyBytes)
	fmt.Printf("Request body: %s\n", bodyString)

	return &http.Response{}, nil
}

type ManagedIdentityStub struct {
	swaggerClient msiClient
}

func NewStub(cloud string) (*ManagedIdentityStub, error) {
	plOpts := runtime.PipelineOptions{
		PerCall: []policy.Policy{
			&injectIdentityURLPolicy{
				msiHost: getMsiHost(cloud),
			},
		},
	}

	clientOpts := &policy.ClientOptions{
		Transport: &stubTransporter{},
	}

	azCoreClient, err := azcore.NewClient("managedidentitydataplane.APIClient", moduleVersion, plOpts, clientOpts)
	if err != nil {
		return nil, err
	}
	swaggerClient := swagger.NewSwaggerClient(azCoreClient)
	stub := &ManagedIdentityStub{
		swaggerClient: swaggerClient,
	}
	return stub, nil
}

func (s *ManagedIdentityStub) GetUserAssignedIdentities(ctx context.Context, request UserAssignedMSIRequest) (*UserAssignedIdentities, error) {
	ctx = context.WithValue(ctx, identityURLKey, request.IdentityURL)
	_, err := s.swaggerClient.Getcreds(ctx, swagger.CredRequestDefinition{
		IdentityIDs: []*string{test.StringPtr("test")},
	}, nil)

	return nil, err
}
