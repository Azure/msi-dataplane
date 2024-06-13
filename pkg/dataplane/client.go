package dataplane

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/msi-dataplane/internal/swagger"
	"github.com/go-playground/validator/v10"
)

//go:generate /bin/bash -c "../../hack/mockgen.sh mock_swagger_client/zz_generated_mocks.go client.go"

// TODO - Tie the module version to update automatically with new releases
const moduleVersion = "v0.0.1"

type ManagedIdentityClient struct {
	swaggerClient swaggerMSIClient
}

type UserAssignedMSIRequest struct {
	IdentityURL string `validate:"required,url"`
	ResourceID  string `validate:"required"`
	TenantID    string `validate:"required,uuid"`
}

type swaggerMSIClient interface {
	Getcreds(ctx context.Context, credRequest swagger.CredRequestDefinition, options *swagger.ManagedIdentityDataPlaneAPIClientGetcredsOptions) (swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse, error)
}

var _ swaggerMSIClient = &swagger.ManagedIdentityDataPlaneAPIClient{}

var (
	errExpectedOneMSI     = fmt.Errorf("expected one user-assigned managed identity")
	errExpectedNonNilMSI  = fmt.Errorf("expected non-nil user-assigned managed identity")
	errInvalidRequest     = fmt.Errorf("invalid request")
	errResourceIDMismatch = fmt.Errorf("resource ID mismatch")
)

// TODO - Add parameter to specify module name in azcore.NewClient()
// NewClient creates a new Managed Identity Dataplane API client
func NewClient(aud, cloud string, cred azcore.TokenCredential) (*ManagedIdentityClient, error) {
	plOpts := runtime.PipelineOptions{
		PerCall: []policy.Policy{
			newAuthenticator(cred, aud),
			&injectIdentityURLPolicy{
				msiHost: getMsiHost(cloud),
			},
		},
	}

	azCoreClient, err := azcore.NewClient("managedidentitydataplane.APIClient", moduleVersion, plOpts, nil)
	if err != nil {
		return nil, err
	}
	swaggerClient := swagger.NewSwaggerClient(azCoreClient)

	return &ManagedIdentityClient{swaggerClient: swaggerClient}, nil
}

func (c *ManagedIdentityClient) GetUserAssignedMSI(ctx context.Context, request UserAssignedMSIRequest) (*CredentialsObject, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(request); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidRequest, err)
	}

	ctx = context.WithValue(ctx, identityURLKey, request.IdentityURL)
	credRequestDef := swagger.CredRequestDefinition{
		IdentityIDs: []*string{&request.ResourceID},
	}

	creds, err := c.swaggerClient.Getcreds(ctx, credRequestDef, nil)
	if err != nil {
		return nil, err
	}

	//
	// GetCreds can return multiple identities. We expect one. Return if we don't find one.
	//
	// If mulitple identities are found, we return because we don't want to accidentally process
	// an identitiy that wasn't requested.
	//
	if len(creds.ExplicitIdentities) != 1 {
		return nil, fmt.Errorf("%w, found %d", errExpectedOneMSI, len(creds.ExplicitIdentities))
	}

	uaMSI := creds.ExplicitIdentities[0]
	if uaMSI == nil {
		return nil, errExpectedNonNilMSI
	}
	if *uaMSI.ResourceID != request.ResourceID {
		return nil, fmt.Errorf("%w, expected %s, got %s", errResourceIDMismatch, request.ResourceID, *uaMSI.ResourceID)
	}

	// Tenant ID is a header passed to RP frontend, so set it here if it's not set
	if *uaMSI.TenantID == "" {
		*uaMSI.TenantID = request.TenantID
	}

	credObject := &CredentialsObject{
		CredentialsObject: creds.CredentialsObject,
	}

	return credObject, nil
}

func getMsiHost(cloud string) string {
	switch cloud {
	case AzureUSGovCloud:
		return usGovMSIEndpoint
	default:
		return publicMSIEndpoint
	}
}
