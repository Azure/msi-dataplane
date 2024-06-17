package dataplane

import (
	"context"
	"errors"
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

	resourceIDsTag = "resource_ids"
)

type ManagedIdentityClient struct {
	swaggerClient msiClient
}

type UserAssignedMSIRequest struct {
	IdentityURL string   `validate:"required,http_url"`
	ResourceIDs []string `validate:"required,resource_ids"`
	TenantID    string   `validate:"required,uuid"`
}

type msiClient interface {
	Getcreds(ctx context.Context, credRequest swagger.CredRequestDefinition, options *swagger.ManagedIdentityDataPlaneAPIClientGetcredsOptions) (swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse, error)
}

var _ msiClient = &swagger.ManagedIdentityDataPlaneAPIClient{}

var (
	// Errors returned by the Managed Identity Dataplane API client
	errGetCreds           = errors.New("failed to get credentials")
	errInvalidRequest     = errors.New("invalid request")
	errNilField           = errors.New("expected non-nil field in user-assigned managed identity")
	errNilMSI             = errors.New("expected non-nil user-assigned managed identity")
	errNumberOfMSIs       = errors.New("returned MSIs does not match number of requested MSIs")
	errResourceIDMismatch = errors.New("requested resource ID not found in response")
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
	validate.RegisterValidation(resourceIDsTag, validateResourceIDs)
	if err := validate.Struct(request); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidRequest, err)
	}

	identityIDs := make([]*string, len(request.ResourceIDs))
	for idx, r := range request.ResourceIDs {
		identityIDs[idx] = &r
	}

	ctx = context.WithValue(ctx, identityURLKey, request.IdentityURL)
	credRequestDef := swagger.CredRequestDefinition{
		IdentityIDs: identityIDs,
	}

	creds, err := c.swaggerClient.Getcreds(ctx, credRequestDef, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errGetCreds, err)
	}

	if len(creds.ExplicitIdentities) != len(request.ResourceIDs) {
		return nil, fmt.Errorf("%w, found %d identities instead", errNumberOfMSIs, len(creds.ExplicitIdentities))
	}

	if err := validateUserAssignedMSIs(creds.ExplicitIdentities, request.ResourceIDs); err != nil {
		return nil, err
	}

	// Tenant ID is a header passed to RP frontend, so set it here if it's not set
	for _, identity := range creds.ExplicitIdentities {
		if *identity.TenantID == "" {
			*identity.TenantID = request.TenantID
		}
	}

	return &CredentialsObject{CredentialsObject: creds.CredentialsObject}, nil
}

func validateResourceIDs(fl validator.FieldLevel) bool {
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

func validateUserAssignedMSIs(identities []*swagger.NestedCredentialsObject, resourceIDs []string) error {
	resourceIDMap := make(map[string]interface{})
	for _, identity := range identities {
		if identity == nil {
			return errNilMSI
		}

		v := reflect.ValueOf(*identity)
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).IsNil() {
				return fmt.Errorf("%w, field %s", errNilField, v.Type().Field(i).Name)
			}
		}
		resourceIDMap[*identity.ResourceID] = true
	}

	for _, resourceID := range resourceIDs {
		if _, ok := resourceIDMap[resourceID]; !ok {
			return fmt.Errorf("%w, resource ID %s", errResourceIDMismatch, resourceID)
		}
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
