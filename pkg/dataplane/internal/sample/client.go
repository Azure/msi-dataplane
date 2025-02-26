package main

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

	"github.com/Azure/msi-dataplane/pkg/dataplane"
)

func main() {
	azureCredential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to initialize azure credentials: %v", err)
	}

	// create a client for the MSI dataplane
	factory := dataplane.NewClientFactory(azureCredential, "audience", nil)

	identityURL := "" // value of the x-ms-identity-url header from ARM
	msiClient, err := factory.NewClient(identityURL)
	if err != nil {
		log.Fatalf("error creating msi dataplane client: %v", err)
	}

	// get the credential for some identities
	credential, err := msiClient.GetUserAssignedIdentitiesCredentials(context.Background(), dataplane.UserAssignedIdentitiesRequest{
		IdentityIDs: []string{
			"someIdentity",
			"someOtherIdentity",
		},
	})
	if err != nil {
		log.Fatalf("error retrieving credential: %v", err)
	}

	// create a client for KeyVault
	keyVaultUrl := "" // from your configuration
	secretsClient, err := azsecrets.NewClient(keyVaultUrl, azureCredential, nil)
	if err != nil {
		log.Fatalf("error creating secrets client: %v", err)
	}

	// either store as a single msi in KeyVault
	identifier := "" // something meaningful to you
	name, params, err := dataplane.FormatManagedIdentityCredentialsForStorage(identifier, *credential)
	if err != nil {
		log.Fatalf("error formatting managed identity credentials: %v", err)
	}
	if _, err := secretsClient.SetSecret(context.Background(), name, params, nil); err != nil {
		log.Fatalf("error uploading managed identity credentials to key vault: %v", err)
	}

	// or store individual uamsi values
	for _, identity := range credential.ExplicitIdentities {
		name, params, err := dataplane.FormatUserAssignedIdentityCredentialsForStorage(identifier, identity)
		if err != nil {
			log.Fatalf("error formatting user-assigned managed identity credentials: %v", err)
		}
		if _, err := secretsClient.SetSecret(context.Background(), name, params, nil); err != nil {
			log.Fatalf("error uploading user-assigned managed identity credentials to key vault: %v", err)
		}
	}
}
