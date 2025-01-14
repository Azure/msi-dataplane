package dataplane

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	azcloud "github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
	"github.com/fsnotify/fsnotify"
)

var (
	// Errors returned when processing idenities
	errDecodeClientSecret = errors.New("failed to decode client secret")
	errParseCertificate   = errors.New("failed to parse certificate")
	errParseResourceID    = errors.New("failed to parse resource ID")
	errNilField           = errors.New("expected non nil field in identity")
	errNoUserAssignedMSIs = errors.New("credentials object does not contain user-assigned managed identities")
	errResourceIDNotFound = errors.New("resource ID not found in user-assigned managed identity")
	errloadCredentials    = errors.New("failed to load credentials from file")
	errCreateFileWatcher  = errors.New("failed to create file watcher")
)

// CredentialsObject is a wrapper around the swagger.CredentialsObject to add additional functionality
// swagger.Credentials object can represent either system or user-assigned managed identity
type CredentialsObject struct {
	Values swagger.CredentialsObject
	Cloud  string
}

// NestedCredentialsObject is a wrapper around the swagger.NestedCredentialsObject to add additional functionality
// swagger.NestedCredentials object can represent only user-assigned managed identity
type NestedCredentialsObject struct {
	Values      swagger.NestedCredentialsObject
	Cloud       string
	File        *File
	fileWatcher *fsnotify.Watcher
	watchOnce   sync.Once
	mu          sync.Mutex
}

// This method may be used by clients to check if they can use the object as a user-assigned managed identity
// Ex: get credentials object from key vault store and check if it is a user-assigned managed identity to call client for object refresh.
func (c CredentialsObject) IsUserAssigned() bool {
	return len(c.Values.ExplicitIdentities) > 0
}

// Get an AzIdentity credential for the given credential object user-assigned identity resource ID
// Clients can use the credential to get a token for the user-assigned identity
func (c CredentialsObject) GetCredential(requestedResourceID string) (*azidentity.ClientCertificateCredential, error) {
	requestedARMResourceID, err := arm.ParseResourceID(requestedResourceID)
	if err != nil {
		return nil, fmt.Errorf("%w for requested resource ID %s: %w", errParseResourceID, requestedResourceID, err)
	}
	requestedResourceID = requestedARMResourceID.String()

	for _, id := range c.Values.ExplicitIdentities {
		if id != nil && id.ResourceID != nil {
			idARMResourceID, err := arm.ParseResourceID(*id.ResourceID)
			if err != nil {
				return nil, fmt.Errorf("%w for identity resource ID %s: %w", errParseResourceID, *id.ResourceID, err)
			}
			if requestedResourceID == idARMResourceID.String() {
				return getClientCertificateCredential(*id, c.Cloud)
			}
		}
	}

	return nil, errResourceIDNotFound
}

// Get an AzIdentity credential for the given nested credential object
// Clients can use the credential to get a token for the user-assigned identity
func (n *NestedCredentialsObject) GetCredential() (*azidentity.ClientCertificateCredential, error) {
	if n.Values.ResourceID != nil {
		return getClientCertificateCredential(n.Values, n.Cloud)
	}

	return nil, errResourceIDNotFound
}

func (n *NestedCredentialsObject) ReloadCredntialsOnChange() error {
	if err := n.File.checkFileExists(); err != nil {
		return err
	}

	err := n.initializeWatcher()
	if err != nil {
		return err
	}
	n.File.initializeFileLock()

	// watch file events under new go routine.
	// this will be called only once
	n.watchOnce.Do(func() {
		go n.watchEvents()
	})

	// we close the file watcher if adding the file to watch fails.
	// this will also close the new go routine created to watch the file
	err = n.fileWatcher.Add(n.File.Path)
	if err != nil {
		n.CloseFileWatch()
		return err
	}

	return nil
}

// initializeWatcher creates a new file watcher if it doesn't already exist
func (n *NestedCredentialsObject) initializeWatcher() error {
	if n.fileWatcher != nil {
		return nil
	}

	var err error
	n.fileWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("%w: %w", errCreateFileWatcher, err)
	}

	return nil
}

func (n *NestedCredentialsObject) watchEvents() {
	for {
		select {
		case event, ok := <-n.fileWatcher.Events:
			if !ok {
				return
			}
			if event.Op.Has(fsnotify.Write) {
				if err := n.loadValuesFromFile(); err != nil {
					log.Printf("%v: %v", errloadCredentials, err)
				}
			}
		case err, ok := <-n.fileWatcher.Errors:
			if !ok {
				return
			}
			log.Printf("%v: %v", errloadCredentials, err)
		}
	}
}

// loadValuesFromFile reads the file and unmarshals the contents into the NestedCredentialsObject
func (n *NestedCredentialsObject) loadValuesFromFile() error {
	if err := n.File.checkFileExists(); err != nil {
		return err
	}

	file, err := os.Open(n.File.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locked, err := n.File.fileLock.TryLockContext(ctx, time.Second)
	if err != nil {
		return err
	}

	if locked {
		defer n.File.fileLock.Unlock()

		byteValue, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		n.mu.Lock()
		defer n.mu.Unlock()

		err = n.Values.UnmarshalJSON(byteValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *NestedCredentialsObject) CloseFileWatch() {
	n.fileWatcher.Close()
}

func getAzCoreCloud(cloud string) azcloud.Configuration {
	switch cloud {
	case AzureUSGovCloud:
		return azcloud.AzureGovernment
	default:
		return azcloud.AzurePublic
	}
}

func getClientCertificateCredential(identity swagger.NestedCredentialsObject, cloud string) (*azidentity.ClientCertificateCredential, error) {
	// Double check nil pointers so we don't panic
	fieldsToCheck := map[string]*string{
		"clientID":               identity.ClientID,
		"tenantID":               identity.TenantID,
		"clientSecret":           identity.ClientSecret,
		"authenticationEndpoint": identity.AuthenticationEndpoint,
	}
	missing := make([]string, 0)
	for field, val := range fieldsToCheck {
		if val == nil {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: %s", errNilField, strings.Join(missing, ","))
	}

	opts := &azidentity.ClientCertificateCredentialOptions{
		ClientOptions: azcore.ClientOptions{
			Cloud: getAzCoreCloud(cloud),
		},

		// x5c header required: https://eng.ms/docs/products/arm/rbac/managed_identities/msionboardingrequestingatoken
		SendCertificateChain: true,

		// Disable instance discovery because MSI credential may have regional AAD endpoint that instance discovery endpoint doesn't support
		// e.g. when MSI credential has westus2.login.microsoft.com, it will cause instance discovery to fail with HTTP 400
		DisableInstanceDiscovery: true,
	}

	// Set the regional AAD endpoint
	// https://eng.ms/docs/products/arm/rbac/managed_identities/msionboardingcredentialapiversion2019-08-31
	opts.Cloud.ActiveDirectoryAuthorityHost = *identity.AuthenticationEndpoint

	// Parse the certificate and private key from the base64 encoded secret
	decodedSecret, err := base64.StdEncoding.DecodeString(*identity.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errDecodeClientSecret, err)
	}
	// Note - ParseCertificates does not currently support pkcs12 SHA256 MAC certs, so if
	// managed identity team changes the cert format, double check this code
	crt, key, err := azidentity.ParseCertificates(decodedSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errParseCertificate, err)
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
		if identity.ResourceID == nil {
			return fmt.Errorf("%w, resource ID", errNilField)
		}
		armResourceID, err := arm.ParseResourceID(*identity.ResourceID)
		if err != nil {
			return fmt.Errorf("%w for received resource ID %s: %w", errParseResourceID, *identity.ResourceID, err)
		}

		resourceIDMap[armResourceID.String()] = true
	}

	for _, resourceID := range resourceIDs {
		armResourceID, err := arm.ParseResourceID(resourceID)
		if err != nil {
			return fmt.Errorf("%w for requested resource ID %s: %w", errParseResourceID, resourceID, err)
		}
		if _, ok := resourceIDMap[armResourceID.String()]; !ok {
			return fmt.Errorf("%w, resource ID %s", errResourceIDNotFound, resourceID)
		}
	}

	return nil
}
