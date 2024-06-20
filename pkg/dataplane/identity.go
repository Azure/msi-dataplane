package dataplane

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	azcloud "github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/internal/swagger"
	"golang.org/x/crypto/pkcs12"
)

var (
	// Errors returned when processing idenities
	errDecodeClientSecret = errors.New("failed to decode client secret")
	errParseCertificate   = errors.New("failed to parse certificate")
	errNilField           = errors.New("expected non nil field in identity")
	errNotRSAKey          = errors.New("pkcs#12 certificate must contain a RSA private key")
	errResourceIDNotFound = errors.New("resource ID not found in user-assigned managed identity	")
)

// CredentialsObject is a wrapper around the swagger.CredentialsObject to add additional functionality
// swagger.Credentials object can represent either system or user-assigned managed identity
type CredentialsObject struct {
	swagger.CredentialsObject
	cloud string
}

type UserAssignedIdentities struct {
	CredentialsObject
}

// This method may be used by clients to check if they can use the object as a user-assigned managed identity
// Ex: get credentials object from key vault store and check if it is a user-assigned managed identity to call client for object refresh.
func (c CredentialsObject) IsUserAssigned() bool {
	return len(c.ExplicitIdentities) > 0
}

// Return an AzIdentity credential for the given user-assigned identity resource ID
// Clients can use the credential to get a token for the user-assigned identity
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
	// Double check nil pointers so we don't panic
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
	decodedSecret, err := base64.StdEncoding.DecodeString(*identity.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errDecodeClientSecret, err)
	}
	key, crt, err := pkcs12.Decode(decodedSecret, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errParseCertificate, err)
	}
	rsaKey, isRsaKey := key.(*rsa.PrivateKey)
	if !isRsaKey {
		return nil, errNotRSAKey
	}

	return azidentity.NewClientCertificateCredential(*identity.TenantID, *identity.ClientID, []*x509.Certificate{crt}, rsaKey, opts)
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
			return fmt.Errorf("%w, resource ID %s", errResourceIDNotFound, resourceID)
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
