package dataplane

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// TODO - Tie the module version to update automatically with new releases
const moduleVersion = "v0.0.1"

// TODO - Add parameter to specify module name in azcore.NewClient()
// NewClient creates a new Managed Identity Dataplane API client
func NewClient(aud, cloud string, cred azcore.TokenCredential) (*ManagedIdentityDataPlaneAPIClient, error) {
	plOpts := runtime.PipelineOptions{
		PerCall: []policy.Policy{
			newAuthenticator(cred, aud),
			&injectIdentityURLPolicy{
				msiHost: getMsiHost(cloud),
			},
		},
	}

	client, err := azcore.NewClient("managedidentitydataplane.APIClient", moduleVersion, plOpts, nil)
	if err != nil {
		return nil, err
	}

	return &ManagedIdentityDataPlaneAPIClient{internal: client}, nil
}

func getMsiHost(cloud string) string {
	switch cloud {
	case AzureUSGovCloud:
		return usGovMSIEndpoint
	default:
		return publicMSIEndpoint
	}
}
