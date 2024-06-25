//go:build unit

package dataplane

import (
	"context"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/msi-dataplane/internal/test"
)

type fakeTransport struct {
	reqs  []*http.Request
	resps []*http.Response
}

func (ft *fakeTransport) Do(req *http.Request) (*http.Response, error) {
	ft.reqs = append(ft.reqs, req)
	return ft.resps[len(ft.reqs)-1], nil
}

func TestNewAuthenticatorPolicy(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name          string
		fakeTransport *fakeTransport
		validateRes   func(*WithT, *fakeTransport, *http.Response, error)
	}{
		{
			name: "Returns success when given a valid path",
			fakeTransport: &fakeTransport{
				resps: []*http.Response{
					{
						StatusCode: http.StatusUnauthorized,
						Header: http.Header{
							"Www-Authenticate": []string{
								`Bearer authorization="https://login.windows-ppe.net/5D929AE3-B37C-46AA-A3C8-C1558902F101"`,
							},
						},
						Body: http.NoBody,
					},
					{
						Body: http.NoBody,
					},
				},
			},
			validateRes: func(g *WithT, fakeTransport *fakeTransport, resp *http.Response, err error) {
				g.Expect(fakeTransport.reqs[0].Header).NotTo(HaveKey(headerAuthorization))
				g.Expect(fakeTransport.reqs[1].Header.Get(headerAuthorization)).To(Equal(
					"Bearer fake_token, tenantID 5d929ae3-b37c-46aa-a3c8-c1558902f101, " +
						"scopes [https://identity_url.com//.default]"))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(resp).To(Equal(fakeTransport.resps[1]))
			},
		},
		{
			name: "failure, authorization is not URL",
			fakeTransport: &fakeTransport{
				resps: []*http.Response{
					{
						StatusCode: http.StatusUnauthorized,
						Header: http.Header{
							"Www-Authenticate": []string{"Bearer authorization=\"\x00\""},
						},
						Body: http.NoBody,
					},
				},
			},
			validateRes: func(g *WithT, fakeTransport *fakeTransport, resp *http.Response, err error) {
				g.Expect(fakeTransport.reqs[0].Header).NotTo(HaveKey("Authorization"))
				g.Expect(err).To(MatchError(errInvalidAuthHeader))
				g.Expect(resp).To(Equal(fakeTransport.resps[0]))
			},
		},
		{
			name: "failure, authorization doesn't have tenant ID",
			fakeTransport: &fakeTransport{
				resps: []*http.Response{
					{
						StatusCode: http.StatusUnauthorized,
						Header: http.Header{
							"Www-Authenticate": []string{`Bearer authorization="https://localhost/"`},
						},
						Body: http.NoBody,
					},
				},
			},
			validateRes: func(g *WithT, fakeTransport *fakeTransport, resp *http.Response, err error) {
				g.Expect(fakeTransport.reqs[0].Header).NotTo(HaveKey("Authorization"))
				g.Expect(err).To(MatchError(errInvalidTenantID))
				g.Expect(resp).To(Equal(fakeTransport.resps[0]))
			},
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			pipeline := runtime.NewPipeline("", "", runtime.PipelineOptions{
				PerCall: []policy.Policy{
					NewAuthenticatorPolicy(&test.FakeCredential{}, "https://identity_url.com/"),
				},
			}, &policy.ClientOptions{
				Transport: tt.fakeTransport,
			})

			req, err := runtime.NewRequest(context.Background(), http.MethodGet, "https://localhost/")
			g.Expect(err).NotTo(HaveOccurred())

			resp, err := pipeline.Do(req)
			tt.validateRes(g, tt.fakeTransport, resp, err)
		})
	}
}
