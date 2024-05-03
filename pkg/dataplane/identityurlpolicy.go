package dataplane

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// injectIdentityURLPolicy injects the msi url to be used when calling the MSI dataplane swagger api client
type injectIdentityURLPolicy struct {
	nextForTest func(req *policy.Request) (*http.Response, error)
	validator   validator
}

func (t *injectIdentityURLPolicy) Do(req *policy.Request) (*http.Response, error) {
	// The Context has the identity url that we need to append with the apiVersion
	apiVersion := req.Raw().URL.Query().Get(apiVersionParameter)
	if err := t.validator.validateApiVersion(apiVersion); err != nil {
		return nil, errAPIVersion
	}

	rawIdentityURL := req.Raw().Context().Value(IdentityURLKey).(string)
	msiURL, err := url.Parse(rawIdentityURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidURL, err)
	}

	if err := t.validator.validateIdentityUrl(msiURL); err != nil {
		return nil, fmt.Errorf("MSI identity URL: %q is invalid: %w", msiURL, err)
	}

	// Append URL with version and set the IdentityURL to the modified value
	appendAPIVersion(msiURL, apiVersion)
	req.Raw().URL = msiURL
	req.Raw().Host = msiURL.Host

	return t.next(req)
}

func appendAPIVersion(u *url.URL, version string) {
	q := u.Query()
	q.Set(apiVersionParameter, version)
	u.RawQuery = q.Encode()
}

// allows to fake the response in test
func (t *injectIdentityURLPolicy) next(req *policy.Request) (*http.Response, error) {
	if t.nextForTest == nil {
		return req.Next()
	}
	return t.nextForTest(req)
}
