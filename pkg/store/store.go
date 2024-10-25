package store

import (
	"context"
	"errors"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/Azure/msi-dataplane/pkg/dataplane"
)

var (
	errNilSecretValue = errors.New("secret value is nil")
)

type DeletedSecretProperties struct {
	Name          string
	RecoveryLevel string
	DeletedDate   time.Time
}

type DeletedSecretResponse struct {
	CredentialsObject dataplane.CredentialsObject
	Properties        DeletedSecretProperties
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

type CertificateSecretResponse struct {
	Certificate string
	Properties  SecretProperties
}

func NewMsiKeyVaultStore(kvClient KeyVaultClient) *MsiKeyVaultStore {
	return &MsiKeyVaultStore{kvClient: kvClient}
}

// Delete a credentials object from key vault using the specified secret name.
// Delete applies to all versions of the secret.
func (s *MsiKeyVaultStore) DeleteCredentialsObject(ctx context.Context, secretName string) error {
	return s.deleteSecret(ctx, secretName)
}

// Delete a certificate object from key vault using the specified secret name.
// Delete applies to all versions of the secret.
func (s *MsiKeyVaultStore) DeleteCertificateObject(ctx context.Context, secretName string) error {
	return s.deleteSecret(ctx, secretName)
}

func (s *MsiKeyVaultStore) deleteSecret(ctx context.Context, secretName string) error {
	if _, err := s.kvClient.DeleteSecret(ctx, secretName, nil); err != nil {
		return err
	}

	return nil
}

// Get a credentials object from the key vault using the specified secret name.
// The latest version of the secret will always be returned.
func (s *MsiKeyVaultStore) GetCredentialsObject(ctx context.Context, secretName string) (*CredentialsObjectSecretResponse, error) {
	secret, properties, err := s.getSecret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	var credentialsObject dataplane.CredentialsObject
	if err := credentialsObject.UnmarshalJSON([]byte(secret)); err != nil {
		return nil, err
	}

	return &CredentialsObjectSecretResponse{CredentialsObject: credentialsObject, Properties: properties}, nil
}

// Get a certificate object from the key vault using the specified secret name.
// The latest version of the secret will always be returned.
func (s *MsiKeyVaultStore) GetCertificateObject(ctx context.Context, secretName string) (*CertificateSecretResponse, error) {
	secret, properties, err := s.getSecret(ctx, secretName)
	if err != nil {
		return nil, err
	}

	return &CertificateSecretResponse{Certificate: secret, Properties: properties}, nil
}

func (s *MsiKeyVaultStore) getSecret(ctx context.Context, secretName string) (string, SecretProperties, error) {
	// https://github.com/Azure/azure-sdk-for-go/blob/3fab729f1bd43098837ddc34931fec6c342fa3ef/sdk/security/keyvault/azsecrets/client.go#L197
	latestSecretVersion := ""
	secretResponse, err := s.kvClient.GetSecret(ctx, secretName, latestSecretVersion, nil)
	if err != nil {
		return "", SecretProperties{}, err
	}

	if secretResponse.Value == nil {
		return "", SecretProperties{}, errNilSecretValue
	}

	secretProperties := SecretProperties{
		Name:      secretName,
		Enabled:   true, // Default to true
		Expires:   time.Time{},
		NotBefore: time.Time{},
	}

	if secretResponse.Attributes != nil {
		// Override defaults if values are present
		if secretResponse.Attributes.Enabled != nil {
			secretProperties.Enabled = *secretResponse.Attributes.Enabled
		}
		if secretResponse.Attributes.Expires != nil {
			secretProperties.Expires = *secretResponse.Attributes.Expires
		}
		if secretResponse.Attributes.NotBefore != nil {
			secretProperties.NotBefore = *secretResponse.Attributes.NotBefore
		}
	}

	return *secretResponse.Value, secretProperties, nil
}

// Get a deleted credentials object from the key vault using the specified secret name.
func (s *MsiKeyVaultStore) GetDeletedCredentialsObject(ctx context.Context, secretName string) (*DeletedSecretResponse, error) {
	response, err := s.kvClient.GetDeletedSecret(ctx, secretName, nil)
	if err != nil {
		return nil, err
	}

	if response.Value == nil {
		return nil, errNilSecretValue
	}

	var credentialsObject dataplane.CredentialsObject
	if err := credentialsObject.UnmarshalJSON([]byte(*response.Value)); err != nil {
		return nil, err
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

	return &DeletedSecretResponse{CredentialsObject: credentialsObject, Properties: deletedSecretProperties}, nil
}

// Get a pager for listing credentials objects from the key vault.
func (s *MsiKeyVaultStore) GetCredentialsObjectPager() *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	return s.kvClient.NewListSecretPropertiesPager(nil)
}

// Get a pager for listing deleted credentials objects from the key vault.
func (s *MsiKeyVaultStore) GetDeletedCredentialsObjectPager() *runtime.Pager[azsecrets.ListDeletedSecretPropertiesResponse] {
	return s.kvClient.NewListDeletedSecretPropertiesPager(nil)
}

// Purge a deleted credentials object from the key vault using the specified secret name.
// This operation is only applicable in vaults enabled for soft-delete.
func (s *MsiKeyVaultStore) PurgeDeletedCredentialsObject(ctx context.Context, secretName string) error {
	if _, err := s.kvClient.PurgeDeletedSecret(ctx, secretName, nil); err != nil {
		return err
	}

	return nil
}

// Purge a deleted certificate object from the key vault using the specified secret name.
// This operation is only applicable in vaults enabled for soft-delete.
func (s *MsiKeyVaultStore) PurgeDeletedCertificateObject(ctx context.Context, secretName string) error {
	if _, err := s.kvClient.PurgeDeletedSecret(ctx, secretName, nil); err != nil {
		return err
	}

	return nil
}

// Set a credentials object in the key vault using the specified secret name.
// If the secret already exists, key vault will create a new version of the secret.
func (s *MsiKeyVaultStore) SetCredentialsObject(ctx context.Context, properties SecretProperties, credentialsObject dataplane.CredentialsObject) error {
	credentialsObjectBuffer, err := credentialsObject.MarshalJSON()
	if err != nil {
		return err
	}

	credentialsObjectString := string(credentialsObjectBuffer)
	return s.setSecret(ctx, properties, credentialsObjectString)
}

// Set a backing certificate in the key vault using the specified secret name
func (s *MsiKeyVaultStore) SetBackingCertificate(ctx context.Context, properties SecretProperties, backingCertificate string) error {
	return s.setSecret(ctx, properties, backingCertificate)
}

func (s *MsiKeyVaultStore) setSecret(ctx context.Context, properties SecretProperties, secret string) error {
	setSecretParameters := azsecrets.SetSecretParameters{
		Value: &secret,
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
