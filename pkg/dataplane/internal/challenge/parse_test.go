package challenge

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		input  http.Header
		output []Challenge
	}{
		{
			name:   "empty",
			input:  http.Header{},
			output: nil,
		},
		{
			name: "only a scheme",
			input: http.Header{
				http.CanonicalHeaderKey("WWW-Authenticate"): []string{`Basic`},
			},
			output: []Challenge{
				{Scheme: "Basic", Parameters: map[string]string{}},
			},
		},
		{
			name: "one scheme, some params",
			input: http.Header{
				http.CanonicalHeaderKey("WWW-Authenticate"): []string{`Basic realm="Dev", charset="UTF-8"`},
			},
			output: []Challenge{
				{Scheme: "Basic", Parameters: map[string]string{"realm": "Dev", "charset": "UTF-8"}},
			},
		},
		{
			name: "one scheme, some params",
			input: http.Header{
				http.CanonicalHeaderKey("WWW-Authenticate"): []string{`Digest realm="http-auth@example.org", qop="auth, auth-int", algorithm=SHA-256, nonce="7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v", opaque="FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"`},
			},
			output: []Challenge{
				{Scheme: "Digest", Parameters: map[string]string{
					"realm":     "http-auth@example.org",
					"qop":       "auth, auth-int",
					"algorithm": "SHA-256",
					"nonce":     "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v",
					"opaque":    "FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS",
				}},
			},
		},
		{
			name: "one scheme, multiple headers",
			input: http.Header{
				http.CanonicalHeaderKey("WWW-Authenticate"): []string{
					`Digest realm="http-auth@example.org", qop="auth, auth-int", algorithm=SHA-256, nonce="7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v", opaque="FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"`,
					`Digest realm="http-auth@example.org", qop="auth, auth-int", algorithm=MD5, nonce="7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v", opaque="FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS"`,
				},
			},
			output: []Challenge{
				{Scheme: "Digest", Parameters: map[string]string{
					"realm":     "http-auth@example.org",
					"qop":       "auth, auth-int",
					"algorithm": "SHA-256",
					"nonce":     "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v",
					"opaque":    "FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS",
				}},
				{Scheme: "Digest", Parameters: map[string]string{
					"realm":     "http-auth@example.org",
					"qop":       "auth, auth-int",
					"algorithm": "MD5",
					"nonce":     "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v",
					"opaque":    "FQhe/qaU925kfnzjCev0ciny7QMkPqMAFRtzCUYo5tdS",
				}},
			},
		},
		{
			name: "many schemes, one header",
			input: http.Header{
				http.CanonicalHeaderKey("WWW-Authenticate"): []string{
					`Basic realm="simple", Newauth realm="apps", type=1, title="Login to \"apps\""`,
				},
			},
			output: []Challenge{
				{Scheme: "Basic", Parameters: map[string]string{
					"realm": "simple",
				}},
				{Scheme: "Newauth", Parameters: map[string]string{
					"realm": "apps",
					"type":  "1",
					"title": `Login to "apps"`,
				}},
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			challenges, err := Parse(testCase.input)
			require.NoError(t, err)
			got, want := challenges, testCase.output
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("invalid parse: -want, +got:\n%s", diff)
			}
		})
	}
}
