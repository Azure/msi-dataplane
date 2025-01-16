package dataplane

import (
	"context"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
)

type reloadingCredential struct {
	currentValue *azidentity.ClientCertificateCredential
	cloud        string
	lock         *sync.RWMutex
	logger       *logr.Logger
}

type Option func(*reloadingCredential)

func WithLogger(logger logr.Logger) Option {
	return func(c *reloadingCredential) {
		c.logger = &logger
	}
}

func NewUserAssignedIdentityCredential(ctx context.Context, cloud string, credentialPath string, opts ...Option) (azcore.TokenCredential, error) {
	credential := &reloadingCredential{
		cloud: cloud,
		lock:  &sync.RWMutex{},
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

func (r *reloadingCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.currentValue.GetToken(ctx, options)
}

func (r *reloadingCredential) start(ctx context.Context, cloud, credentialFile string) {
	// set up the file watcher, call load() when we see events or on some timer in case no events are delivered
	fileWatcher, err := fsnotify.NewWatcher()
	if err != nil && r.logger != nil {
		r.logger.Error(err, "failed to create file watcher")
	}
	// we close the file watcher if adding the file to watch fails.
	// this will also close the new go routine created to watch the file
	err = fileWatcher.Add(credentialFile)
	if err != nil {
		fileWatcher.Close()
		r.logger.Error(err, "failed to add credentialFile to file watcher")
		return
	}

	go func() {
		for {
			select {
			case event, ok := <-fileWatcher.Events:
				if !ok {
					return
				}
				if event.Op.Has(fsnotify.Write) {
					if err := r.load(cloud, credentialFile); err != nil && r.logger != nil {
						r.logger.Error(err, "failed to load credentials from file")
					}
				}
			case err, ok := <-fileWatcher.Errors:
				if !ok {
					return
				}
				r.logger.Error(err, "failed to load credentials from file")
			}
		}
	}()

	// Keep the function running until the context is done
	<-ctx.Done()
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
	r.currentValue = newCertValue

	return nil
}
