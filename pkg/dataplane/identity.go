package dataplane

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	azcloud "github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/internal/swagger"
)

var (
	errNilField           = errors.New("expected non nil field in identity")
	errResourceIDNotFound = errors.New("resource ID not found in user-assigned managed identity	")
)

type CredentialsObject struct {
	swagger.CredentialsObject
	cloud string
}

type UserAssignedIdentities struct {
	CredentialsObject
}

// swagger.Credentials object can represent either system or user-assigned managed identity
// This method may be used by clients to check if they can use the object is a user-assigned managed identity
func (c CredentialsObject) IsUserAssigned() bool {
	return len(c.ExplicitIdentities) > 0
}

func (u UserAssignedIdentities) GetCredential(resourceID string) (*azidentity.ClientCertificateCredential, error) {
	for _, id := range u.ExplicitIdentities {
		if id != nil && id.ResourceID != nil {
			if *id.ResourceID == resourceID {
				return getClientCertificateCredential(*id, u.cloud)
			}
		}
	}

	return nil, errResourceIDNotFound
}

func getClientCertificateCredential(identity swagger.NestedCredentialsObject, cloud string) (*azidentity.ClientCertificateCredential, error) {
	// Double check no nil pointers so we don't accidentally panic
	if identity.ClientID == nil {
		return nil, fmt.Errorf("%w: clientID", errNilField)
	}
	if identity.TenantID == nil {
		return nil, fmt.Errorf("%w: tenantID", errNilField)
	}
	if identity.ClientSecret == nil {
		return nil, fmt.Errorf("%w: clientSecret", errNilField)
	}
	if identity.AuthenticationEndpoint == nil {
		return nil, fmt.Errorf("%w: authenticationEndpoint", errNilField)
	}

	// Set the regional AAD endpoint
	// https://eng.ms/docs/products/arm/rbac/managed_identities/msionboardingcredentialapiversion2019-08-31
	opts := &azidentity.ClientCertificateCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: getAzCoreCloud(cloud),
		},
	}
	opts.Cloud.ActiveDirectoryAuthorityHost = *identity.AuthenticationEndpoint

	// Parse the certificate and private key from the base64 encoded secret
	pkcs12, err := base64.StdEncoding.DecodeString(*identity.ClientSecret)
	if err != nil {
		return nil, err
	}
	crt, key, err := azidentity.ParseCertificates(pkcs12, nil)
	if err != nil {
		return nil, err
	}

	return azidentity.NewClientCertificateCredential(*identity.TenantID, *identity.ClientID, crt, key, opts)
}

func validateUserAssignedMSIs(identities []*swagger.NestedCredentialsObject, resourceIDs []string) error {
	if len(identities) != len(resourceIDs) {
		return fmt.Errorf("%w, found %d identities instead", errNumberOfMSIs, len(identities))
	}

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

func getAzCoreCloud(cloud string) azcloud.Configuration {
	switch cloud {
	case AzureUSGovCloud:
		return azcloud.AzureGovernment
	default:
		return azcloud.AzurePublic
	}
}
