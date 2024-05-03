//go:build unit

package dataplane

import (
	"testing"
)

func TestGetValidator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		cloud           string
		expectedMsiHost string
	}{
		{
			name:            "AzurePublicCloud",
			cloud:           AzurePublicCloud,
			expectedMsiHost: publicMSIEndpoint,
		},
		{
			name:            "AzureUSGovCloud",
			cloud:           AzureUSGovCloud,
			expectedMsiHost: usGovMSIEndpoint,
		},
		{
			name:            "Default",
			cloud:           "",
			expectedMsiHost: publicMSIEndpoint,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v := getValidator(tc.cloud)
			if v.msiHost == "" {
				t.Fatalf("msiHost is empty")
			}
			if v.hostRegexp == nil {
				t.Errorf("hostRegexp is nil")
			}
		})
	}
}
