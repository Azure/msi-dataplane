package store

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/Azure/msi-dataplane/pkg/dataplane"
)

type SecretProperties struct {
	Enabled   bool
	Expires   time.Time
	Name      string
	NotBefore time.Time
}

type SecretResponse struct {
	CredentialsObject dataplane.CredentialsObject
	Properties        SecretProperties
}

type CredentialsObjectPager struct {
	pager          *runtime.Pager[azsecrets.ListSecretPropertiesResponse]
	propertiesList []*azsecrets.SecretProperties
}

// Returns the secret properties for the next secret in the pager
// If secret properties and error are nil, the list is exhausted
func (p *CredentialsObjectPager) NextSecretProperty(ctx context.Context) (*SecretProperties, error) {
	if len(p.propertiesList) == 0 {
		// Use the pager to update properties list
		if !p.pager.More() {
			return nil, nil
		}

		resp, err := p.pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		p.propertiesList = resp.SecretPropertiesListResult.Value
	}

	var name string
	if p.propertiesList[0].ID != nil {
		name = p.propertiesList[0].ID.Name()
	} else {
		return nil, fmt.Errorf("secret has no id")
	}

	if p.propertiesList[0].Attributes == nil {
		return nil, fmt.Errorf("secret %s has no attributes", name)
	}

	var enabled bool
	if p.propertiesList[0].Attributes != nil && p.propertiesList[0].Attributes.Enabled != nil {
		enabled = *p.propertiesList[0].Attributes.Enabled
	} else {
		return nil, fmt.Errorf("secret %s has no enabled attribute", name)
	}

	var expires time.Time
	if p.propertiesList[0].Attributes != nil && p.propertiesList[0].Attributes.Expires != nil {
		expires = *p.propertiesList[0].Attributes.Expires
	} else {
		return nil, fmt.Errorf("secret %s has no expires attribute", name)
	}

	var notBefore time.Time
	if p.propertiesList[0].Attributes != nil && p.propertiesList[0].Attributes.NotBefore != nil {
		notBefore = *p.propertiesList[0].Attributes.NotBefore
	} else {
		return nil, fmt.Errorf("secret %s has no notBefore attribute", name)
	}

	prop := &SecretProperties{
		Enabled:   enabled,
		Expires:   expires,
		Name:      name,
		NotBefore: notBefore,
	}

	p.propertiesList = p.propertiesList[1:]

	return prop, nil
}
