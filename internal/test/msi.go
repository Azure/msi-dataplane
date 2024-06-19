package test

import "github.com/Azure/msi-dataplane/internal/swagger"

func GetTestMSI(placeHolder string) *swagger.NestedCredentialsObject {
	return &swagger.NestedCredentialsObject{
		AuthenticationEndpoint:     &placeHolder,
		CannotRenewAfter:           &placeHolder,
		ClientID:                   &placeHolder,
		ClientSecret:               &placeHolder,
		ClientSecretURL:            &placeHolder,
		CustomClaims:               &swagger.CustomClaims{},
		MtlsAuthenticationEndpoint: &placeHolder,
		NotAfter:                   &placeHolder,
		NotBefore:                  &placeHolder,
		ObjectID:                   &placeHolder,
		RenewAfter:                 &placeHolder,
		ResourceID:                 &placeHolder,
		TenantID:                   &placeHolder,
	}
}
