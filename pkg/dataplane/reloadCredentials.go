package dataplane

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
)

const (
	// Errors returned when reloading credentials
	errCreateFileWatcher = "failed to create file watcher"
	errAddFileToWatcher  = "failed to add credentialFile to file watcher"
	errLoadCredentials   = "failed to load credentials from file"
)

type ReloadingCredential struct {
	ClientCertificateCredential *ClientCertificateCredential
	cloud                       string
	lock                        *sync.RWMutex
	logger                      *logr.Logger
	ticker                      *time.Ticker
}

type ClientCertificateCredential struct {
	currentValue *azidentity.ClientCertificateCredential
	NotBefore    string
}

type Option func(*ReloadingCredential)

// WithLogger sets a custom logger for the reloadingCredential.
// This can be useful for debugging or logging purposes.
func WithLogger(logger *logr.Logger) Option {
	return func(c *ReloadingCredential) {
		c.logger = logger
	}
}

// WithBackstopRefresh sets a custom timer for the reloadingCredential.
// This can be useful for loading credential file periodically.
func WithBackstopRefresh(d time.Duration) Option {
	return func(c *ReloadingCredential) {
		c.ticker = time.NewTicker(d)
	}
}

// NewUserAssignedIdentityCredential creates a new reloadingCredential for a user-assigned identity.
// ctx is used to manage the lifecycle of the reloader, allowing for cancellation if reloading is no longer needed.
// cloud specifies the cloud environment.
// credentialPath is the path to the credential file.
// opts allows for additional configuration, such as setting a custom logger, periodic reload time.
//
// The function ensures that a valid token is loaded before returning the credential.
// It also starts a background process to watch for changes to the credential file and reloads it as necessary.
func NewUserAssignedIdentityCredential(ctx context.Context, cloud string, credentialPath string, opts ...Option) (azcore.TokenCredential, error) {
	defaultLog := logr.FromSlogHandler(slog.NewTextHandler(os.Stdout, nil))
	credential := &ReloadingCredential{
		cloud:  cloud,
		lock:   &sync.RWMutex{},
		logger: &defaultLog,
		ticker: time.NewTicker(6 * time.Hour),
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
func (r *ReloadingCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.ClientCertificateCredential.currentValue.GetToken(ctx, options)
}

func (r *ReloadingCredential) start(ctx context.Context, cloud, credentialFile string) {
	// set up the file watcher, call load() when we see events or on some timer in case no events are delivered
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.logger.Error(err, errCreateFileWatcher)
	}
	// we close the file watcher if adding the file to watch fails.
	// this will also close the new go routine created to watch the file
	err = fileWatcher.Add(credentialFile)
	if err != nil {
		fileWatcher.Close()
		r.logger.Error(err, errAddFileToWatcher)
		return
	}

	go func() {
		defer fileWatcher.Close()
		defer r.ticker.Stop()
		for {
			select {
			case event, ok := <-fileWatcher.Events:
				if !ok {
					r.logger.Info("stopping credential reloader since file watcher has no events")
					return
				}
				if event.Op.Has(fsnotify.Write) {
					if err := r.load(cloud, credentialFile); err != nil {
						r.logger.Error(err, errLoadCredentials)
					}
				}
			case <-r.ticker.C:
				if err := r.load(cloud, credentialFile); err != nil {
					r.logger.Error(err, errLoadCredentials)
				}
			case err, ok := <-fileWatcher.Errors:
				if !ok {
					return
				}
				r.logger.Error(err, errLoadCredentials)
			case <-ctx.Done():
				r.logger.Info("user signaled context cancel, stopping credential reloader")
				return
			}
		}
	}()
}

func (r *ReloadingCredential) load(cloud, credentialFile string) error {
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

	err, ok := isLoadedCredentialNewer(*nestedCreds.NotBefore, r.ClientCertificateCredential.NotBefore)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	var newCertValue *azidentity.ClientCertificateCredential
	newCertValue, err = getClientCertificateCredential(nestedCreds, cloud)
	if err != nil {
		return err
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.ClientCertificateCredential.currentValue = newCertValue
	r.ClientCertificateCredential.NotBefore = *nestedCreds.NotBefore

	return nil
}

func isLoadedCredentialNewer(newCred string, currentCred string) (error, bool) {
	parsedNewCred, err := time.Parse(time.RFC3339, newCred)
	if err != nil {
		return err, false
	}

	parsedCurrentCred, err := time.Parse(time.RFC3339, currentCred)
	if err != nil {
		return err, false
	}

	return nil, parsedNewCred.After(parsedCurrentCred) || parsedNewCred.Equal(parsedCurrentCred)
}
