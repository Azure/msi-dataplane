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

// TODO - Tie the module version to update automatically with new releases
const moduleVersion = "v0.0.1"

type ManagedIdentityClient struct {
	swaggerClient *swagger.ManagedIdentityDataPlaneAPIClient
}

type UserAssignedMSIRequest struct {
	IdentityURL string `validate:"required,url"`
	ResourceID  string `validate:"required"`
	TenantID    string `validate:"required,uuid"`
}

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
		return nil, err
	}

	ctx = context.WithValue(ctx, identityURLKey, request.IdentityURL)
	credRequestDef := swagger.CredRequestDefinition{
		IdentityIDs: []*string{&request.ResourceID},
	}

	creds, err := c.swaggerClient.Getcreds(ctx, credRequestDef, nil)
	if err != nil {
		return nil, err
	}

	if len(creds.ExplicitIdentities) != 1 {
		return nil, fmt.Errorf("expected one user-assigned managed identity, found %d", len(creds.ExplicitIdentities))
	}

	for _, identity := range creds.ExplicitIdentities {
		if *identity.ResourceID != request.ResourceID {
			return nil, fmt.Errorf("resourceID mismatch, expected %s, got %s", request.ResourceID, *identity.ResourceID)
		}
		if *identity.TenantID == "" {
			*identity.TenantID = request.TenantID
		}
	}

	credObject := &CredentialsObject{
		CredentialsObject: creds.CredentialsObject,
	}

	return credObject, err
}

func getMsiHost(cloud string) string {
	switch cloud {
	case AzureUSGovCloud:
		return usGovMSIEndpoint
	default:
		return publicMSIEndpoint
	}
}
