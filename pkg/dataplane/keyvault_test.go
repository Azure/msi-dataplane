package dataplane

import (
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/google/go-cmp/cmp"

	"github.com/Azure/msi-dataplane/pkg/dataplane/internal/client"
)

func TestFormatManagedIdentityCredentialsForStorage(t *testing.T) {
	var notAfter, notBefore, renewAfter, cannotRenewAfter time.Time
	for raw, into := range map[string]*time.Time{
		"2006-01-02T15:04:05Z": &notAfter,
		"2001-01-02T15:04:05Z": &notBefore,
		"2003-01-02T15:04:05Z": &renewAfter,
		"2023-01-02T15:04:05Z": &cannotRenewAfter,
	} {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", raw, err)
		}
		*into = parsed
	}

	for _, testCase := range []struct {
		name             string
		identifier       string
		credentials      ManagedIdentityCredentials
		keyVaultItemName string
		parameters       azsecrets.SetSecretParameters
		err              bool
	}{
		{
			name:       "valid msi credentials",
			identifier: "test",
			credentials: ManagedIdentityCredentials{
				ClientSecretURL:  ptrTo("whatever"),
				CannotRenewAfter: ptrTo(cannotRenewAfter.Format(time.RFC3339)),
				NotAfter:         ptrTo(notAfter.Format(time.RFC3339)),
				NotBefore:        ptrTo(notBefore.Format(time.RFC3339)),
				RenewAfter:       ptrTo(renewAfter.Format(time.RFC3339)),
			},
			keyVaultItemName: "msi-test",
			parameters: azsecrets.SetSecretParameters{
				Value: ptrTo(`{"cannot_renew_after":"2023-01-02T15:04:05Z","client_secret_url":"whatever","not_after":"2006-01-02T15:04:05Z","not_before":"2001-01-02T15:04:05Z","renew_after":"2003-01-02T15:04:05Z"}`),
				SecretAttributes: &azsecrets.SecretAttributes{
					Enabled:   ptrTo(true),
					Expires:   ptrTo(notAfter),
					NotBefore: ptrTo(notBefore),
				},
				Tags: map[string]*string{
					"renew_after":        ptrTo(renewAfter.Format(time.RFC3339)),
					"cannot_renew_after": ptrTo(cannotRenewAfter.Format(time.RFC3339)),
				},
			},
		},
		{
			name:       "invalid msi credentials",
			identifier: "test",
			credentials: ManagedIdentityCredentials{
				ClientSecretURL:  ptrTo("whatever"),
				CannotRenewAfter: ptrTo(cannotRenewAfter.Format(time.RFC3339)),
				NotAfter:         ptrTo(notAfter.Format(time.RFC3339)),
				NotBefore:        ptrTo("oops"),
				RenewAfter:       ptrTo(renewAfter.Format(time.RFC3339)),
			},
			err: true,
		},
		{
			name:       "valid uamsi credentials",
			identifier: "test",
			credentials: ManagedIdentityCredentials{
				ExplicitIdentities: []client.UserAssignedIdentityCredentials{
					{
						ClientSecretURL:  ptrTo("whatever"),
						CannotRenewAfter: ptrTo(cannotRenewAfter.Format(time.RFC3339)),
						NotAfter:         ptrTo(notAfter.Format(time.RFC3339)),
						NotBefore:        ptrTo(notBefore.Format(time.RFC3339)),
						RenewAfter:       ptrTo(renewAfter.Format(time.RFC3339)),
					},
				},
			},
			keyVaultItemName: "msi-test",
			parameters: azsecrets.SetSecretParameters{
				Value: ptrTo(`{"explicit_identities":[{"cannot_renew_after":"2023-01-02T15:04:05Z","client_secret_url":"whatever","not_after":"2006-01-02T15:04:05Z","not_before":"2001-01-02T15:04:05Z","renew_after":"2003-01-02T15:04:05Z"}]}`),
				SecretAttributes: &azsecrets.SecretAttributes{
					Enabled:   ptrTo(true),
					Expires:   ptrTo(notAfter),
					NotBefore: ptrTo(notBefore),
				},
				Tags: map[string]*string{
					"renew_after":        ptrTo(renewAfter.Format(time.RFC3339)),
					"cannot_renew_after": ptrTo(cannotRenewAfter.Format(time.RFC3339)),
				},
			},
		},
		{
			name:       "invalid uamsi credentials",
			identifier: "test",
			credentials: ManagedIdentityCredentials{
				ExplicitIdentities: []client.UserAssignedIdentityCredentials{
					{
						ClientSecretURL:  ptrTo("whatever"),
						CannotRenewAfter: ptrTo(cannotRenewAfter.Format(time.RFC3339)),
						NotAfter:         ptrTo("oops"),
						NotBefore:        ptrTo(notBefore.Format(time.RFC3339)),
						RenewAfter:       ptrTo(renewAfter.Format(time.RFC3339)),
					},
				},
			},
			err: true,
		},
		{
			name:       "too many uamsi credentials",
			identifier: "test",
			credentials: ManagedIdentityCredentials{
				ExplicitIdentities: []client.UserAssignedIdentityCredentials{
					{
						ClientSecretURL: ptrTo("first"),
					},
					{
						ClientSecretURL: ptrTo("second"),
					},
				},
			},
			err: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			name, parameters, err := FormatManagedIdentityCredentialsForStorage(testCase.identifier, testCase.credentials)
			if err == nil && testCase.err {
				t.Fatalf("%s: expected error, got none", testCase.name)
			}
			if err != nil && !testCase.err {
				t.Fatalf("%s: expected no error, got %v", testCase.name, err)
			}

			if name != testCase.keyVaultItemName {
				t.Errorf("%s: expected name %q, got %q", testCase.name, testCase.keyVaultItemName, name)
			}
			if diff := cmp.Diff(parameters, testCase.parameters); diff != "" {
				t.Errorf("%s: parameters (-want, +got) = %v", testCase.name, diff)
			}
		})
	}
}
