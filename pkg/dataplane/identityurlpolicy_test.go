//go:build unit

package dataplane

import (
	"context"
	"errors"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

func TestInjectURLPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		ctx              context.Context
		finalInjectedURL string
		finalInjectHost  string
		cloud            string // to initialize what msi host to use
		apiVersion       string
		expectedError    error
	}{
		{
			name:             "SUCCESS - valid request with url injected - public cloud",
			ctx:              buildCtx("https://test.identity.azure.net/my-blah-url?k1=v1&k2=v2"),
			finalInjectedURL: "https://test.identity.azure.net/my-blah-url?api-version=2023-02-28&k1=v1&k2=v2",
			finalInjectHost:  "test.identity.azure.net",
			cloud:            AzurePublicCloud,
			apiVersion:       "?api-version=2023-02-28",
			expectedError:    nil,
		},
		{
			name:          "FAILURE - no api version",
			ctx:           buildCtx("https://test.identity.azure.net/my-blah-url?k1=v1&k2=v2"),
			cloud:         AzurePublicCloud,
			expectedError: errAPIVersion,
		},
		{
			name:          "FAILURE - non string context value",
			ctx:           buildCtx(123),
			cloud:         AzurePublicCloud,
			expectedError: errInvalidCtxValueType,
		},
		{
			name:          "FAILURE - invalid url",
			ctx:           buildCtx("https://test.identity.azure.net/my-blah-url\"\x00\""),
			cloud:         AzurePublicCloud,
			apiVersion:    "?api-version=2023-02-28",
			expectedError: errInvalidIdentityURL,
		},
		{
			name:          "FAILURE - non https",
			ctx:           buildCtx("http://test.identity.azure.net/my-blah-url"),
			cloud:         AzurePublicCloud,
			apiVersion:    "?api-version=2023-02-28",
			expectedError: errNotHTTPS,
		},
		{
			name:          "FAILURE - not the correct msi host",
			ctx:           buildCtx("https://bad.host.com/my-blah-url"),
			cloud:         AzurePublicCloud,
			apiVersion:    "?api-version=2023-02-28",
			expectedError: errInvalidDomain,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()
			g := NewWithT(tt)

			var nextRequest *http.Request
			injectURLPolicy := &injectIdentityURLPolicy{
				nextForTest: func(req *policy.Request) (*http.Response, error) {
					nextRequest = req.Raw()
					return &http.Response{}, nil
				},
				msiHost: getMsiHost(test.cloud),
			}

			endpoint := "https://management.azure.com" + test.apiVersion
			// MSI API client hardcodes the endpoint with API version, mimic that
			req, err := runtime.NewRequest(test.ctx, http.MethodGet, endpoint)
			g.Expect(err).ToNot(HaveOccurred())

			_, err = injectURLPolicy.Do(req)

			if test.expectedError != nil {
				g.Expect(err).To(HaveOccurred())
				g.Expect(errors.Is(err, test.expectedError))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(nextRequest).ToNot(BeNil())
				g.Expect(nextRequest.URL.String()).To(Equal(test.finalInjectedURL))
				g.Expect(nextRequest.URL.Host).To(Equal(test.finalInjectHost))
			}
		})
	}
}

func buildCtx(value any) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, IdentityURLKey, value)
}
