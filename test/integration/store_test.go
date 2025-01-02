package integration

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/Azure/msi-dataplane/internal/test"
	"github.com/Azure/msi-dataplane/pkg/dataplane"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
	"github.com/Azure/msi-dataplane/pkg/store"
	"github.com/google/go-cmp/cmp"
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

	msiStore, err := createMsiStore(r, cred)
	if err != nil {
		t.Fatalf("Failed to create MSI store: %s", err)
	}

	// Add a test credentials object to the store
	bogus := test.Bogus
	testCredentialsObject := dataplane.CredentialsObject{
		Values: swagger.CredentialsObject{
			ClientID: &bogus,
		},
	}
	expirationDate, err := time.Parse(time.RFC3339, time.Now().Add(time.Hour).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to parse expiry date: %s", err)
	}
	notBeforeDate, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to parse notBefore date: %s", err)
	}
	props := store.SecretProperties{
		Name:      bogus,
		Enabled:   true,
		Expires:   expirationDate,
		NotBefore: notBeforeDate,
	}

	testCredentialsObjectSecretResponse := store.CredentialsObjectSecretResponse{
		Properties:        props,
		CredentialsObject: testCredentialsObject,
	}

	if err := msiStore.SetCredentialsObject(context.Background(), props, testCredentialsObject); err != nil {
		// Fatal here since rest of test cannot proceed
		t.Fatalf("Failed to set credentials object: %s", err)
	}

	// Get the credentials object from the store
	resp, err := msiStore.GetCredentialsObject(context.Background(), bogus)
	if err != nil {
		// Fatal here since rest of test cannot proceed
		t.Fatalf("Failed to get credentials object: %s", err)
	}

<<<<<<< HEAD
	if !reflect.DeepEqual(testCredentialsObject, resp.CredentialsObject) {
		t.Errorf(`Credential objects do not match. 
		          Returned has client ID %s, expected %s`, *resp.CredentialsObject.Values.ClientID, *testCredentialsObject.Values.ClientID)
=======
	if diff := cmp.Diff(resp, &testCredentialsObjectSecretResponse); diff != "" {
		t.Errorf("Expected credentials object %+v\n but got: %+v", &testCredentialsObjectSecretResponse, resp)
>>>>>>> update integration test
	}

	// Delete the credentials object from the store
	if err := msiStore.DeleteSecret(context.Background(), bogus); err != nil {
		t.Errorf("Failed to delete credentials object: %s", err)
	}
}

func TestNestedCredentialsObjectStore(t *testing.T) {
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

	msiStore, err := createMsiStore(r, cred)
	if err != nil {
		t.Fatalf("Failed to create MSI store: %s", err)
	}

	// Add a test nested credentials object to the store
	bogus := "NestedBogus"
	testNestedCredentialsObject := swagger.NestedCredentialsObject{
		ClientSecret: &bogus,
	}
	expirationDate, err := time.Parse(time.RFC3339, time.Now().Add(time.Hour).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to parse expiry date: %s", err)
	}
	notBeforeDate, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to parse notBefore date: %s", err)
	}
	props := store.SecretProperties{
		Name:      bogus,
		Enabled:   true,
		Expires:   expirationDate,
		NotBefore: notBeforeDate,
	}

	testNestedCredentialsObjectSecretResponse := store.NestedCredentialsObjectSecretResponse{
		Properties:              props,
		NestedCredentialsObject: testNestedCredentialsObject,
	}

	if err := msiStore.SetNestedCredentialsObject(context.Background(), props, testNestedCredentialsObject); err != nil {
		// Fatal here since rest of test cannot proceed
		t.Fatalf("Failed to set nested credentials object: %s", err)
	}

	// Get the credentials object from the store
	resp, err := msiStore.GetNestedCredentialsObject(context.Background(), bogus)
	if err != nil {
		// Fatal here since rest of test cannot proceed
		t.Fatalf("Failed to get nested credentials object: %s", err)
	}

	if diff := cmp.Diff(resp, &testNestedCredentialsObjectSecretResponse); diff != "" {
		t.Errorf("Expected nested credentials object %+v\n but got: %+v", &testNestedCredentialsObjectSecretResponse, resp)
	}

	// Delete the credentials object from the store
	if err := msiStore.DeleteSecret(context.Background(), bogus); err != nil {
		t.Errorf("Failed to nested delete credentials object: %s", err)
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
