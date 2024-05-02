//go:build go1.18
// +build go1.18

// Code generated by Microsoft (R) AutoRest Code Generator (autorest: 3.10.2, generator: @autorest/go@4.0.0-preview.63)
// Changes may cause incorrect behavior and will be lost if the code is regenerated.
// Code generated by @autorest/go. DO NOT EDIT.

package dataplane

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"reflect"
)

// MarshalJSON implements the json.Marshaller interface for type CredRequestDefinition.
func (c CredRequestDefinition) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "customClaims", c.CustomClaims)
	populate(objectMap, "delegatedResources", c.DelegatedResources)
	populate(objectMap, "identityIds", c.IdentityIDs)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type CredRequestDefinition.
func (c *CredRequestDefinition) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", c, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "customClaims":
			err = unpopulate(val, "CustomClaims", &c.CustomClaims)
			delete(rawMsg, key)
		case "delegatedResources":
			err = unpopulate(val, "DelegatedResources", &c.DelegatedResources)
			delete(rawMsg, key)
		case "identityIds":
			err = unpopulate(val, "IdentityIDs", &c.IdentityIDs)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", c, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type CredentialsObject.
func (c CredentialsObject) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "authentication_endpoint", c.AuthenticationEndpoint)
	populate(objectMap, "cannot_renew_after", c.CannotRenewAfter)
	populate(objectMap, "client_id", c.ClientID)
	populate(objectMap, "client_secret", c.ClientSecret)
	populate(objectMap, "client_secret_url", c.ClientSecretURL)
	populate(objectMap, "custom_claims", c.CustomClaims)
	populate(objectMap, "delegated_resources", c.DelegatedResources)
	populate(objectMap, "delegation_url", c.DelegationURL)
	populate(objectMap, "explicit_identities", c.ExplicitIdentities)
	populate(objectMap, "internal_id", c.InternalID)
	populate(objectMap, "mtls_authentication_endpoint", c.MtlsAuthenticationEndpoint)
	populate(objectMap, "not_after", c.NotAfter)
	populate(objectMap, "not_before", c.NotBefore)
	populate(objectMap, "object_id", c.ObjectID)
	populate(objectMap, "renew_after", c.RenewAfter)
	populate(objectMap, "tenant_id", c.TenantID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type CredentialsObject.
func (c *CredentialsObject) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", c, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "authentication_endpoint":
			err = unpopulate(val, "AuthenticationEndpoint", &c.AuthenticationEndpoint)
			delete(rawMsg, key)
		case "cannot_renew_after":
			err = unpopulate(val, "CannotRenewAfter", &c.CannotRenewAfter)
			delete(rawMsg, key)
		case "client_id":
			err = unpopulate(val, "ClientID", &c.ClientID)
			delete(rawMsg, key)
		case "client_secret":
			err = unpopulate(val, "ClientSecret", &c.ClientSecret)
			delete(rawMsg, key)
		case "client_secret_url":
			err = unpopulate(val, "ClientSecretURL", &c.ClientSecretURL)
			delete(rawMsg, key)
		case "custom_claims":
			err = unpopulate(val, "CustomClaims", &c.CustomClaims)
			delete(rawMsg, key)
		case "delegated_resources":
			err = unpopulate(val, "DelegatedResources", &c.DelegatedResources)
			delete(rawMsg, key)
		case "delegation_url":
			err = unpopulate(val, "DelegationURL", &c.DelegationURL)
			delete(rawMsg, key)
		case "explicit_identities":
			err = unpopulate(val, "ExplicitIdentities", &c.ExplicitIdentities)
			delete(rawMsg, key)
		case "internal_id":
			err = unpopulate(val, "InternalID", &c.InternalID)
			delete(rawMsg, key)
		case "mtls_authentication_endpoint":
			err = unpopulate(val, "MtlsAuthenticationEndpoint", &c.MtlsAuthenticationEndpoint)
			delete(rawMsg, key)
		case "not_after":
			err = unpopulate(val, "NotAfter", &c.NotAfter)
			delete(rawMsg, key)
		case "not_before":
			err = unpopulate(val, "NotBefore", &c.NotBefore)
			delete(rawMsg, key)
		case "object_id":
			err = unpopulate(val, "ObjectID", &c.ObjectID)
			delete(rawMsg, key)
		case "renew_after":
			err = unpopulate(val, "RenewAfter", &c.RenewAfter)
			delete(rawMsg, key)
		case "tenant_id":
			err = unpopulate(val, "TenantID", &c.TenantID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", c, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type CustomClaims.
func (c CustomClaims) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "xms_az_nwperimid", c.XMSAzNwperimid)
	populate(objectMap, "xms_az_tm", c.XMSAzTm)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type CustomClaims.
func (c *CustomClaims) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", c, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "xms_az_nwperimid":
			err = unpopulate(val, "XMSAzNwperimid", &c.XMSAzNwperimid)
			delete(rawMsg, key)
		case "xms_az_tm":
			err = unpopulate(val, "XMSAzTm", &c.XMSAzTm)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", c, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type DelegatedResourceObject.
func (d DelegatedResourceObject) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "delegation_id", d.DelegationID)
	populate(objectMap, "delegation_url", d.DelegationURL)
	populate(objectMap, "explicit_identities", d.ExplicitIdentities)
	populate(objectMap, "implicit_identity", d.ImplicitIdentity)
	populate(objectMap, "internal_id", d.InternalID)
	populate(objectMap, "resource_id", d.ResourceID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type DelegatedResourceObject.
func (d *DelegatedResourceObject) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", d, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "delegation_id":
			err = unpopulate(val, "DelegationID", &d.DelegationID)
			delete(rawMsg, key)
		case "delegation_url":
			err = unpopulate(val, "DelegationURL", &d.DelegationURL)
			delete(rawMsg, key)
		case "explicit_identities":
			err = unpopulate(val, "ExplicitIdentities", &d.ExplicitIdentities)
			delete(rawMsg, key)
		case "implicit_identity":
			err = unpopulate(val, "ImplicitIdentity", &d.ImplicitIdentity)
			delete(rawMsg, key)
		case "internal_id":
			err = unpopulate(val, "InternalID", &d.InternalID)
			delete(rawMsg, key)
		case "resource_id":
			err = unpopulate(val, "ResourceID", &d.ResourceID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", d, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorResponse.
func (e ErrorResponse) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "error", e.Error)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorResponse.
func (e *ErrorResponse) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "error":
			err = unpopulate(val, "Error", &e.Error)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type ErrorResponseError.
func (e ErrorResponseError) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "code", e.Code)
	populate(objectMap, "message", e.Message)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type ErrorResponseError.
func (e *ErrorResponseError) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", e, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "code":
			err = unpopulate(val, "Code", &e.Code)
			delete(rawMsg, key)
		case "message":
			err = unpopulate(val, "Message", &e.Message)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", e, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type MoveIdentityResponse.
func (m MoveIdentityResponse) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "identityUrl", m.IdentityURL)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type MoveIdentityResponse.
func (m *MoveIdentityResponse) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", m, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "identityUrl":
			err = unpopulate(val, "IdentityURL", &m.IdentityURL)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", m, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type MoveRequestBodyDefinition.
func (m MoveRequestBodyDefinition) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "targetResourceId", m.TargetResourceID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type MoveRequestBodyDefinition.
func (m *MoveRequestBodyDefinition) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", m, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "targetResourceId":
			err = unpopulate(val, "TargetResourceID", &m.TargetResourceID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", m, err)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaller interface for type NestedCredentialsObject.
func (n NestedCredentialsObject) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]any)
	populate(objectMap, "authentication_endpoint", n.AuthenticationEndpoint)
	populate(objectMap, "cannot_renew_after", n.CannotRenewAfter)
	populate(objectMap, "client_id", n.ClientID)
	populate(objectMap, "client_secret", n.ClientSecret)
	populate(objectMap, "client_secret_url", n.ClientSecretURL)
	populate(objectMap, "custom_claims", n.CustomClaims)
	populate(objectMap, "mtls_authentication_endpoint", n.MtlsAuthenticationEndpoint)
	populate(objectMap, "not_after", n.NotAfter)
	populate(objectMap, "not_before", n.NotBefore)
	populate(objectMap, "object_id", n.ObjectID)
	populate(objectMap, "renew_after", n.RenewAfter)
	populate(objectMap, "resource_id", n.ResourceID)
	populate(objectMap, "tenant_id", n.TenantID)
	return json.Marshal(objectMap)
}

// UnmarshalJSON implements the json.Unmarshaller interface for type NestedCredentialsObject.
func (n *NestedCredentialsObject) UnmarshalJSON(data []byte) error {
	var rawMsg map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMsg); err != nil {
		return fmt.Errorf("unmarshalling type %T: %v", n, err)
	}
	for key, val := range rawMsg {
		var err error
		switch key {
		case "authentication_endpoint":
			err = unpopulate(val, "AuthenticationEndpoint", &n.AuthenticationEndpoint)
			delete(rawMsg, key)
		case "cannot_renew_after":
			err = unpopulate(val, "CannotRenewAfter", &n.CannotRenewAfter)
			delete(rawMsg, key)
		case "client_id":
			err = unpopulate(val, "ClientID", &n.ClientID)
			delete(rawMsg, key)
		case "client_secret":
			err = unpopulate(val, "ClientSecret", &n.ClientSecret)
			delete(rawMsg, key)
		case "client_secret_url":
			err = unpopulate(val, "ClientSecretURL", &n.ClientSecretURL)
			delete(rawMsg, key)
		case "custom_claims":
			err = unpopulate(val, "CustomClaims", &n.CustomClaims)
			delete(rawMsg, key)
		case "mtls_authentication_endpoint":
			err = unpopulate(val, "MtlsAuthenticationEndpoint", &n.MtlsAuthenticationEndpoint)
			delete(rawMsg, key)
		case "not_after":
			err = unpopulate(val, "NotAfter", &n.NotAfter)
			delete(rawMsg, key)
		case "not_before":
			err = unpopulate(val, "NotBefore", &n.NotBefore)
			delete(rawMsg, key)
		case "object_id":
			err = unpopulate(val, "ObjectID", &n.ObjectID)
			delete(rawMsg, key)
		case "renew_after":
			err = unpopulate(val, "RenewAfter", &n.RenewAfter)
			delete(rawMsg, key)
		case "resource_id":
			err = unpopulate(val, "ResourceID", &n.ResourceID)
			delete(rawMsg, key)
		case "tenant_id":
			err = unpopulate(val, "TenantID", &n.TenantID)
			delete(rawMsg, key)
		}
		if err != nil {
			return fmt.Errorf("unmarshalling type %T: %v", n, err)
		}
	}
	return nil
}

func populate(m map[string]any, k string, v any) {
	if v == nil {
		return
	} else if azcore.IsNullValue(v) {
		m[k] = nil
	} else if !reflect.ValueOf(v).IsNil() {
		m[k] = v
	}
}

func unpopulate(data json.RawMessage, fn string, v any) error {
	if data == nil || string(data) == "null" {
		return nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("struct field %s: %v", fn, err)
	}
	return nil
}
