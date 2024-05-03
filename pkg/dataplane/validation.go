package dataplane

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
)

var (
	errAPIVersion    = errors.New("the api-version parameter was not in MSI data plane request")
	errNotHTTPS      = errors.New("the scheme of the MSI URL is not https")
	errInvalidDomain = errors.New("the MSI URL was not the expected domain")
)

type validator struct {
	msiHost    string
	hostRegexp *regexp.Regexp
}

func getValidator(cloud string) validator {
	switch cloud {
	case AzurePublicCloud:
		return validator{
			msiHost:    publicMSIEndpoint,
			hostRegexp: regexp.MustCompile("(?i)^[^.]+[.]" + regexp.QuoteMeta(publicMSIEndpoint) + "$"),
		}
	case AzureUSGovCloud:
		return validator{
			msiHost:    usGovMSIEndpoint,
			hostRegexp: regexp.MustCompile("(?i)^[^.]+[.]" + regexp.QuoteMeta(usGovMSIEndpoint) + "$"),
		}
	default:
		return validator{
			msiHost:    publicMSIEndpoint,
			hostRegexp: regexp.MustCompile("(?i)^[^.]+[.]" + regexp.QuoteMeta(publicMSIEndpoint) + "$"),
		}
	}
}

func (v validator) validateApiVersion(version string) error {
	if version == "" {
		return errAPIVersion
	}
	return nil
}

func (v validator) validateIdentityUrl(u *url.URL) error {
	if u.Scheme != https {
		return fmt.Errorf("%w: %q", errNotHTTPS, u)
	}

	if !v.hostRegexp.MatchString(u.Host) {
		return fmt.Errorf("%w. Given: %q, Expected: %q", errInvalidDomain, u, v.msiHost)
	}

	return nil
}
