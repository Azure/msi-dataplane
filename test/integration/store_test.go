//go:build integration

package integration

import (
	"context"
	"os"
	"path"
	"reflect"
	"testing"

	"msi-prototype/internal/test"
	"msi-prototype/pkg/dataplane"
	"msi-prototype/pkg/store"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	// Cassette directory and names
	CASSETTE_DIR  = "cassettes"
	CASSETTE_NAME = "store"

	// Environment variables for test options
	RECORD_MODE  = "RECORD_MODE"
	KEYVAULT_URL = "KEYVAULT_URL"
)

func TestStore(t *testing.T) {
	t.Parallel()

	recordMode := getRecordMode()

	// Get the credential based on record mode
	var cred azcore.TokenCredential
	var err error
	switch recordMode {
	case recorder.ModeReplayOnly:
		// Use a fake credential for replay mode
		cred = &test.FakeCredential{}
	case recorder.ModeRecordOnly:
		// Use the default Azure credential for record mode
		cred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			t.Fatalf("Failed to create credential: %s", err)
		}
	}

	// Create a recorder for the secrets client
	r, err := getRecorder(recordMode, getCassettePath())
	if err != nil {
		t.Fatalf("Failed to create recorder: %s", err)
	}
	defer r.Stop()

	store, err := createMsiStore(r, cred)
	if err != nil {
		t.Fatalf("Failed to create MSI store: %s", err)
	}

	// Add a test credentials object to the store
	bogus := "bogus"
	testCredentialsObject := dataplane.CredentialsObject{
		ClientID: &bogus,
	}

	if err := store.SetCredentialsObject(context.Background(), bogus, testCredentialsObject); err != nil {
		// Fatal here since rest of test cannot proceed
		t.Fatalf("Failed to set credentials object: %s", err)
	}

	// Get the credentials object from the store
	returnedCredentialsObject, err := store.GetCredentialsObject(context.Background(), bogus)
	if err != nil {
		// Fatal here since rest of test cannot proceed
		t.Fatalf("Failed to get credentials object: %s", err)
	}

	if !reflect.DeepEqual(testCredentialsObject, returnedCredentialsObject) {
		t.Errorf(`Credential objects do not match. 
		          Returned has client ID %s, expected %s`, *returnedCredentialsObject.ClientID, *testCredentialsObject.ClientID)
	}

	// Delete the credentials object from the store
	if err := store.DeleteCredentialsObject(context.Background(), bogus); err != nil {
		t.Errorf("Failed to delete credentials object: %s", err)
	}
}

func createMsiStore(r *recorder.Recorder, cred azcore.TokenCredential) (*store.MsiKeyVaultStore, error) {
	vaultURL, err := getVaultURL()
	if err != nil {
		return nil, err
	}
	secretsClient, err := getSecretsClient(r, vaultURL, cred)
	if err != nil {
		return nil, err
	}

	return store.NewMsiKeyVaultStore(secretsClient), nil
}

func getCassettePath() string {
	return path.Join(CASSETTE_DIR, CASSETTE_NAME)
}

func getRecorder(recordMode recorder.Mode, cassettePath string) (*recorder.Recorder, error) {
	const authHeader = "Authorization"

	recorderOpts := &recorder.Options{
		CassetteName: cassettePath,
		Mode:         recordMode,
	}

	r, err := recorder.NewWithOptions(recorderOpts)
	if err != nil {
		return nil, err
	}

	hook := func(i *cassette.Interaction) error {
		// Remove the authorization header from the request
		delete(i.Request.Headers, authHeader)
		return nil
	}
	r.AddHook(hook, recorder.BeforeSaveHook)

	return r, nil
}

func getRecordMode() recorder.Mode {
	if os.Getenv(RECORD_MODE) == "" {
		return recorder.ModeReplayOnly
	}
	return recorder.ModeRecordOnly
}

func getSecretsClient(r *recorder.Recorder, vaultURL string, cred azcore.TokenCredential) (*azsecrets.Client, error) {
	azcoreClientOpts := azcore.ClientOptions{
		Transport: r.GetDefaultClient(),
	}
	azsecretsClientOpts := &azsecrets.ClientOptions{
		ClientOptions: azcoreClientOpts,
	}

	return azsecrets.NewClient(vaultURL, cred, azsecretsClientOpts)
}

func getVaultURL() (string, error) {
	if url := os.Getenv(KEYVAULT_URL); url != "" {
		return url, nil
	}

	// Read the URL from the cassette
	storeCassette, err := cassette.Load(getCassettePath())
	if err != nil {
		return "", err
	}
	host := storeCassette.Interactions[0].Request.Host
	return "https://" + host + "/", nil
}
