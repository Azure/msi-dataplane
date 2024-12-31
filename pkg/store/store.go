package store

import (
	"context"
	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/Azure/msi-dataplane/pkg/dataplane"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
)

var (
	errNilSecretValue = errors.New("secret value is nil")
)

type DeletedSecretProperties struct {
	Name          string
	RecoveryLevel string
	DeletedDate   time.Time
}

type DeletedCredentialsObjectSecretResponse struct {
	CredentialsObject dataplane.CredentialsObject
	Properties        DeletedSecretProperties
}

type DeletedNestedCredentialsObjectSecretResponse struct {
	NestedCredentialsObject swagger.NestedCredentialsObject
	Properties              DeletedSecretProperties
}

type MsiKeyVaultStore struct {
	kvClient KeyVaultClient
}

type SecretProperties struct {
	Enabled   bool
	Expires   time.Time
	Name      string
	NotBefore time.Time
}

type CredentialsObjectSecretResponse struct {
	CredentialsObject dataplane.CredentialsObject
	Properties        SecretProperties
}

type NestedCredentialsObjectSecretResponse struct {
	NestedCredentialsObject swagger.NestedCredentialsObject
	Properties              SecretProperties
}

type secretObject struct {
	value      string
	properties SecretProperties
}

type deletedSecretObject struct {
	value      string
	properties DeletedSecretProperties
}

func NewMsiKeyVaultStore(kvClient KeyVaultClient) *MsiKeyVaultStore {
	return &MsiKeyVaultStore{kvClient: kvClient}
}

// Delete a credentials object from key vault using the specified secret name.
// Delete applies to all versions of the secret.
func (s *MsiKeyVaultStore) DeleteSecret(ctx context.Context, secretName string) error {
	if _, err := s.kvClient.DeleteSecret(ctx, secretName, nil); err != nil {
		return err
	}

	return nil
}

// Get a credentials object from the key vault using the specified secret name.
// The latest version of the secret will always be returned.
func (s *MsiKeyVaultStore) GetCredentialsObject(ctx context.Context, secretName string) (*CredentialsObjectSecretResponse, error) {
	secretObject, err := s.getSecret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	var credentialsObject dataplane.CredentialsObject
	if err := credentialsObject.UnmarshalJSON([]byte(secretObject.value)); err != nil {
		return nil, err
	}

	return &CredentialsObjectSecretResponse{CredentialsObject: credentialsObject, Properties: secretObject.properties}, nil
}

// Get a nested credentials object from the key vault using the specified secret name.
func (s *MsiKeyVaultStore) GetNestedCredentialsObject(ctx context.Context, secretName string) (*NestedCredentialsObjectSecretResponse, error) {
	secretObject, err := s.getSecret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	var nestedCredentialsObject swagger.NestedCredentialsObject
	if err := nestedCredentialsObject.UnmarshalJSON([]byte(secretObject.value)); err != nil {
		return nil, err
	}

	return &NestedCredentialsObjectSecretResponse{NestedCredentialsObject: nestedCredentialsObject, Properties: secretObject.properties}, nil
}

// Get a secret from the key vault using the specified secret name.
// The latest version of the secret will always be returned.
func (s *MsiKeyVaultStore) getSecret(ctx context.Context, secretName string) (*secretObject, error) {
	// https://github.com/Azure/azure-sdk-for-go/blob/3fab729f1bd43098837ddc34931fec6c342fa3ef/sdk/security/keyvault/azsecrets/client.go#L197
	latestSecretVersion := ""
	secret, err := s.kvClient.GetSecret(ctx, secretName, latestSecretVersion, nil)
	if err != nil {
		return nil, err
	}

	if secret.Value == nil {
		return nil, errNilSecretValue
	}

	secretProperties := SecretProperties{
		Name:      secretName,
		Enabled:   true, // Default to true
		Expires:   time.Time{},
		NotBefore: time.Time{},
	}

	if secret.Attributes != nil {
		// Override defaults if values are present
		if secret.Attributes.Enabled != nil {
			secretProperties.Enabled = *secret.Attributes.Enabled
		}
		if secret.Attributes.Expires != nil {
			secretProperties.Expires = *secret.Attributes.Expires
		}
		if secret.Attributes.NotBefore != nil {
			secretProperties.NotBefore = *secret.Attributes.NotBefore
		}
	}
	return &secretObject{value: *secret.Value, properties: secretProperties}, nil
}

// Get a deleted credentials object from the key vault using the specified secret name.
func (s *MsiKeyVaultStore) GetDeletedCredentialsObject(ctx context.Context, secretName string) (*DeletedCredentialsObjectSecretResponse, error) {
	deletedSecretObject, err := s.getDeletedSecret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	var credentialsObject dataplane.CredentialsObject
	if err := credentialsObject.UnmarshalJSON([]byte(deletedSecretObject.value)); err != nil {
		return nil, err
	}

	return &DeletedCredentialsObjectSecretResponse{CredentialsObject: credentialsObject, Properties: deletedSecretObject.properties}, nil
}

// Get a deleted nested credentials object from the key vault using the specified secret name.
func (s *MsiKeyVaultStore) GetDeletedNestedCredentialsObject(ctx context.Context, secretName string) (*DeletedNestedCredentialsObjectSecretResponse, error) {
	deletedSecretObject, err := s.getDeletedSecret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	var nestedCredentialsObject swagger.NestedCredentialsObject
	if err := nestedCredentialsObject.UnmarshalJSON([]byte(deletedSecretObject.value)); err != nil {
		return nil, err
	}

	return &DeletedNestedCredentialsObjectSecretResponse{NestedCredentialsObject: nestedCredentialsObject, Properties: deletedSecretObject.properties}, nil
}

// Get a deleted secret from the key vault using the specified secret name.
func (s *MsiKeyVaultStore) getDeletedSecret(ctx context.Context, secretName string) (*deletedSecretObject, error) {
	response, err := s.kvClient.GetDeletedSecret(ctx, secretName, nil)
	if err != nil {
		return nil, err
	}

	if response.Value == nil {
		return nil, errNilSecretValue
	}

	deletedSecretProperties := DeletedSecretProperties{
		Name:          secretName,
		RecoveryLevel: "",
		DeletedDate:   time.Time{},
	}

	if response.DeletedDate != nil {
		deletedSecretProperties.DeletedDate = *response.DeletedDate
	}

	if response.Attributes != nil {
		// Override defaults if values are present
		if response.Attributes.RecoveryLevel != nil {
			deletedSecretProperties.RecoveryLevel = *response.Attributes.RecoveryLevel
		}
	}

	return &deletedSecretObject{value: *response.Value, properties: deletedSecretProperties}, nil
}

// Get a pager for listing Secret objects from the key vault.
func (s *MsiKeyVaultStore) GetSecretObjectPager() *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	return s.kvClient.NewListSecretPropertiesPager(nil)
}

// Get a pager for listing deleted Secret objects from the key vault.
func (s *MsiKeyVaultStore) GetDeletedSecretObjectPager() *runtime.Pager[azsecrets.ListDeletedSecretPropertiesResponse] {
	return s.kvClient.NewListDeletedSecretPropertiesPager(nil)
}

// Purge a deleted Secret object from the key vault using the specified secret name.
// This operation is only applicable in vaults enabled for soft-delete.
func (s *MsiKeyVaultStore) PurgeDeletedSecretObject(ctx context.Context, secretName string) error {
	if _, err := s.kvClient.PurgeDeletedSecret(ctx, secretName, nil); err != nil {
		return err
	}

	return nil
}

// Set a credentials object in the key vault using the specified secret name.
func (s *MsiKeyVaultStore) SetCredentialsObject(ctx context.Context, properties SecretProperties, credentialsObject dataplane.CredentialsObject) error {
	credentialsObjectBuffer, err := credentialsObject.MarshalJSON()
	if err != nil {
		return err
	}

	credentialsObjectString := string(credentialsObjectBuffer)
	return s.setSecret(ctx, properties, &credentialsObjectString)
}

// Set a nested credentials object in the key vault using the specified secret name.
func (s *MsiKeyVaultStore) SetNestedCredentialsObject(ctx context.Context, properties SecretProperties, nestedCredentialsObject swagger.NestedCredentialsObject) error {
	nestedCredentialsObjectBuffer, err := nestedCredentialsObject.MarshalJSON()
	if err != nil {
		return err
	}

	nestedCredentialsObjectString := string(nestedCredentialsObjectBuffer)
	return s.setSecret(ctx, properties, &nestedCredentialsObjectString)
}

// Set a secret in the key vault using the specified secret name.
// If the secret already exists, key vault will create a new version of the secret.
func (s *MsiKeyVaultStore) setSecret(ctx context.Context, properties SecretProperties, secretValue *string) error {
	setSecretParameters := azsecrets.SetSecretParameters{
		Value: secretValue,
		SecretAttributes: &azsecrets.SecretAttributes{
			Enabled:   &properties.Enabled,
			Expires:   &properties.Expires,
			NotBefore: &properties.NotBefore,
		},
	}
	if _, err := s.kvClient.SetSecret(ctx, properties.Name, setSecretParameters, nil); err != nil {
		return err
	}

	return nil
}
