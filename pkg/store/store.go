package store

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/Azure/msi-dataplane/pkg/dataplane"
)

type MsiKeyVaultStore struct {
	kvClient KeyVaultClient
}

func NewMsiKeyVaultStore(kvClient KeyVaultClient) *MsiKeyVaultStore {
	return &MsiKeyVaultStore{kvClient: kvClient}
}

// Delete a credentials object from key vault using the specified secret name.
// Delete applies to all versions of the secret.
func (s *MsiKeyVaultStore) DeleteCredentialsObject(ctx context.Context, secretName string) error {
	if _, err := s.kvClient.DeleteSecret(ctx, secretName, nil); err != nil {
		return err
	}

	return nil
}

// Get a credentials object from the key vault using the specified secret name.
// The latest version of the secret will always be returned.
func (s *MsiKeyVaultStore) GetCredentialsObject(ctx context.Context, secretName string) (dataplane.CredentialsObject, error) {
	// https://github.com/Azure/azure-sdk-for-go/blob/3fab729f1bd43098837ddc34931fec6c342fa3ef/sdk/security/keyvault/azsecrets/client.go#L197
	latestSecretVersion := ""
	secret, err := s.kvClient.GetSecret(ctx, secretName, latestSecretVersion, nil)
	if err != nil {
		return dataplane.CredentialsObject{}, err
	}

	var credentialsObject dataplane.CredentialsObject
	if err := credentialsObject.UnmarshalJSON([]byte(*secret.Value)); err != nil {
		return dataplane.CredentialsObject{}, err
	}

	return credentialsObject, nil
}

// Get a pager for listing credentials objects from the key vault.
func (s *MsiKeyVaultStore) GetCredentialsObjectPager() *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	return s.kvClient.NewListSecretPropertiesPager(nil)
}

// Purge a deleted credentials object from the key vault using the specified secret name.
// This operation is only applicable in vaults enabled for soft-delete.
func (s *MsiKeyVaultStore) PurgeDeletedCredentialsObject(ctx context.Context, secretName string) error {
	if _, err := s.kvClient.PurgeDeletedSecret(ctx, secretName, nil); err != nil {
		return err
	}

	return nil
}

// Set a credentials object in the key vault using the specified secret name.
// If the secret already exists, key vault will create a new version of the secret.
func (s *MsiKeyVaultStore) SetCredentialsObject(ctx context.Context, secretName string, credentialsObject dataplane.CredentialsObject) error {
	credentialsObjectBuffer, err := credentialsObject.MarshalJSON()
	if err != nil {
		return err
	}

	credentialsObjectString := string(credentialsObjectBuffer)
	setSecretParameters := azsecrets.SetSecretParameters{
		Value: &credentialsObjectString,
	}
	if _, err := s.kvClient.SetSecret(ctx, secretName, setSecretParameters, nil); err != nil {
		return err
	}

	return nil
}
