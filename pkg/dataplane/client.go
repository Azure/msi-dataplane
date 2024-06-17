package dataplane

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/msi-dataplane/internal/swagger"
	"github.com/go-playground/validator/v10"
)

//go:generate /bin/bash -c "../../hack/mockgen.sh mock_swagger_client/zz_generated_mocks.go client.go"

const (
	// TODO - Tie the module version to update automatically with new releases
	moduleVersion = "v0.0.1"

	resourceIDTag = "resource_id"
)

type ManagedIdentityClient struct {
	swaggerClient msiClient
}

type UserAssignedMSIRequest struct {
	IdentityURL string   `validate:"required,http_url"`
	ResourceID  []string `validate:"required,resource_id"`
	TenantID    string   `validate:"required,uuid"`
}

type msiClient interface {
	Getcreds(ctx context.Context, credRequest swagger.CredRequestDefinition, options *swagger.ManagedIdentityDataPlaneAPIClientGetcredsOptions) (swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse, error)
}

var _ msiClient = &swagger.ManagedIdentityDataPlaneAPIClient{}

var (
	// Errors returned by the Managed Identity Dataplane API client
	errGetCreds           = fmt.Errorf("failed to get credentials")
	errInvalidRequest     = fmt.Errorf("invalid request")
	errNilField           = fmt.Errorf("expected non-nil field in user-assigned managed identity")
	errNilMSI             = fmt.Errorf("expected non-nil user-assigned managed identity")
	errNotOneMSI          = fmt.Errorf("expected one user-assigned managed identity")
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
	validate.RegisterValidation(resourceIDTag, validateResourceID)
	if err := validate.Struct(request); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidRequest, err)
	}

	ctx = context.WithValue(ctx, identityURLKey, request.IdentityURL)
	credRequestDef := swagger.CredRequestDefinition{
		IdentityIDs: []*string{&request.ResourceID},
	}

	creds, err := c.swaggerClient.Getcreds(ctx, credRequestDef, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errGetCreds, err)
	}

	//
	// GetCreds can return multiple identities. We expect one. Return if we don't find one.
	//
	// If mulitple identities are found, we return because we don't want to accidentally process
	// an identitiy that wasn't requested.
	//
	if len(creds.ExplicitIdentities) != 1 {
		return nil, fmt.Errorf("%w, found %d identities instead", errNotOneMSI, len(creds.ExplicitIdentities))
	}

	if err := validateUserAssignedMSI(creds.ExplicitIdentities[0], request.ResourceID); err != nil {
		return nil, err
	}

	// Tenant ID is a header passed to RP frontend, so set it here if it's not set
	if *creds.ExplicitIdentities[0].TenantID == "" {
		*creds.ExplicitIdentities[0].TenantID = request.TenantID
	}

	return &CredentialsObject{CredentialsObject: creds.CredentialsObject}, nil
}

func validateResourceID(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Confirm we have a slice of strings
	if field.Kind() != reflect.Slice {
		return false
	}

	if field.Type().Elem().Kind() != reflect.String {
		return false
	}

	// Check we have at least one element
	if field.Len() < 1 {
		return false
	}

	// Check that all elements are valid resource IDs
	for i := 0; i < field.Len(); i++ {
		resourceID := field.Index(i).String()
		if !isUserAssignedMSIResource(resourceID) {
			return false
		}
	}

	return true
}

func isUserAssignedMSIResource(resourceID string) bool {
	_, err := arm.ParseResourceID(resourceID)
	if err != nil {
		return false
	}

	resourceType, err := arm.ParseResourceType(resourceID)
	if err != nil {
		return false
	}

	const expectedNamespace = "Microsoft.ManagedIdentity"
	const expectedResourceType = "userAssignedIdentities"

	return resourceType.Namespace == expectedNamespace && resourceType.Type == expectedResourceType
}

func validateUserAssignedMSI(identity *swagger.NestedCredentialsObject, resourceID string) error {
	if identity == nil {
		return errNilMSI
	}

	v := reflect.ValueOf(*identity)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsNil() {
			return fmt.Errorf("%w, field %s", errNilField, v.Type().Field(i).Name)
		}
	}

	if *identity.ResourceID != resourceID {
		return fmt.Errorf("%w, expected %s, got %s", errResourceIDMismatch, resourceID, *identity.ResourceID)
	}

	return nil
}

func getMsiHost(cloud string) string {
	switch cloud {
	case AzureUSGovCloud:
		return usGovMSIEndpoint
	default:
		return publicMSIEndpoint
	}
}
