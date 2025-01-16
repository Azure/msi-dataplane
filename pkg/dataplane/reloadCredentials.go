package dataplane

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
	"github.com/fsnotify/fsnotify"
)

var (
	// Errors returned when reloading credentials
	errCreateFileWatcher = errors.New("failed to create file watcher")
	errAddFileToWatcher  = errors.New("failed to add credentialFile to file watcher")
	errLoadCredentials   = errors.New("failed to load credentials from file")
)

type reloadingCredential struct {
	currentValue *azidentity.ClientCertificateCredential
	cloud        string
	lock         *sync.RWMutex
	logger       *log.Logger
}

type Option func(*reloadingCredential)

// WithLogger sets a custom logger for the reloadingCredential.
// This can be useful for debugging or logging purposes.
func WithLogger(logger *log.Logger) Option {
	return func(c *reloadingCredential) {
		c.logger = logger
	}
}

// NewUserAssignedIdentityCredential creates a new reloadingCredential for a user-assigned identity.
// ctx is used to manage the lifecycle of the credential, allowing for cancellation and timeouts.
// cloud specifies the cloud environment.
// credentialPath is the path to the credential file.
// opts allows for additional configuration, such as setting a custom logger.
//
// The function ensures that a valid token is loaded before returning the credential.
// It also starts a background process to watch for changes to the credential file and reloads it as necessary.
// In any case the maximum time that we wait before reload is six hours.
// Note that while the credential will attempt to keep the token up-to-date, there may be a small delay between
// when the token expires and when it is reloaded. Users should handle token expiration errors appropriately.
func NewUserAssignedIdentityCredential(ctx context.Context, cloud string, credentialPath string, opts ...Option) (azcore.TokenCredential, error) {
	credential := &reloadingCredential{
		cloud:  cloud,
		lock:   &sync.RWMutex{},
		logger: log.Default(),
	}

	for _, opt := range opts {
		opt(credential)
	}

	// load once to validate everything and ensure we have a useful token before we return
	if err := credential.load(cloud, credentialPath); err != nil {
		return nil, err
	}
	// start the process of watching - the caller can cancel ctx if they want to stop
	credential.start(ctx, cloud, credentialPath)
	return credential, nil
}

// GetToken retrieves the current token from the reloadingCredential.
// It uses a read lock to ensure that the token is not being modified while it is being read.
// options specifies additional options for the token request.
func (r *reloadingCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.currentValue.GetToken(ctx, options)
}

func (r *reloadingCredential) start(ctx context.Context, cloud, credentialFile string) {
	// set up the file watcher, call load() when we see events or on some timer in case no events are delivered
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.logger.Printf("%v, %v", errCreateFileWatcher, err)
	}
	// we close the file watcher if adding the file to watch fails.
	// this will also close the new go routine created to watch the file
	err = fileWatcher.Add(credentialFile)
	if err != nil {
		fileWatcher.Close()
		r.logger.Printf("%v, %v", errAddFileToWatcher, err)
		return
	}

	go func() {
		defer fileWatcher.Close()
		for {
			select {
			case event, ok := <-fileWatcher.Events:
				if !ok {
					return
				}
				if event.Op.Has(fsnotify.Write) {
					if err := r.load(cloud, credentialFile); err != nil && r.logger != nil {
						r.logger.Printf("%v, %v", errLoadCredentials, err)
					}
				}
			case err, ok := <-fileWatcher.Errors:
				if !ok {
					return
				}
				r.logger.Printf("%v, %v", errLoadCredentials, err)
			// Keep the function running until the context is done
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (r *reloadingCredential) load(cloud, credentialFile string) error {
	// read the file from the filesystem and update the current value we're holding on to if the certificate we read is newer, making sure to not step on the toes of anyone calling GetToken()
	byteValue, err := os.ReadFile(credentialFile)
	if err != nil {
		return err
	}

	var nestedCreds swagger.NestedCredentialsObject
	err = nestedCreds.UnmarshalJSON(byteValue)
	if err != nil {
		return err
	}

	var newCertValue *azidentity.ClientCertificateCredential
	newCertValue, err = getClientCertificateCredential(nestedCreds, cloud)
	if err != nil {
		return err
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.currentValue = newCertValue

	return nil
}
