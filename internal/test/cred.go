package test

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type FakeCredential struct{}

func (f *FakeCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token: fmt.Sprintf("fake_token, tenantID %s, scopes %v", opts.TenantID, opts.Scopes),
	}, nil
}
