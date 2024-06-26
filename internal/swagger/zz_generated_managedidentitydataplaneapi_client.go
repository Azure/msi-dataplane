//go:build go1.18
// +build go1.18

// Code generated by Microsoft (R) AutoRest Code Generator (autorest: 3.10.2, generator: @autorest/go@4.0.0-preview.63)
// Changes may cause incorrect behavior and will be lost if the code is regenerated.
// Code generated by @autorest/go. DO NOT EDIT.

package swagger

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"net/http"
)

// ManagedIdentityDataPlaneAPIClient contains the methods for the ManagedIdentityDataPlaneAPI group.
// Don't use this type directly, use a constructor function instead.
type ManagedIdentityDataPlaneAPIClient struct {
	internal *azcore.Client
}

// Deleteidentity -
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2024-01-01
//   - options - ManagedIdentityDataPlaneAPIClientDeleteidentityOptions contains the optional parameters for the ManagedIdentityDataPlaneAPIClient.Deleteidentity
//     method.
func (client *ManagedIdentityDataPlaneAPIClient) Deleteidentity(ctx context.Context, options *ManagedIdentityDataPlaneAPIClientDeleteidentityOptions) (ManagedIdentityDataPlaneAPIClientDeleteidentityResponse, error) {
	var err error
	req, err := client.deleteidentityCreateRequest(ctx, options)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientDeleteidentityResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientDeleteidentityResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK, http.StatusNoContent) {
		err = runtime.NewResponseError(httpResp)
		return ManagedIdentityDataPlaneAPIClientDeleteidentityResponse{}, err
	}
	return ManagedIdentityDataPlaneAPIClientDeleteidentityResponse{}, nil
}

// deleteidentityCreateRequest creates the Deleteidentity request.
func (client *ManagedIdentityDataPlaneAPIClient) deleteidentityCreateRequest(ctx context.Context, options *ManagedIdentityDataPlaneAPIClientDeleteidentityOptions) (*policy.Request, error) {
	req, err := runtime.NewRequest(ctx, http.MethodDelete, host)
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2024-01-01")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}

// Getcred -
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2024-01-01
//   - options - ManagedIdentityDataPlaneAPIClientGetcredOptions contains the optional parameters for the ManagedIdentityDataPlaneAPIClient.Getcred
//     method.
func (client *ManagedIdentityDataPlaneAPIClient) Getcred(ctx context.Context, options *ManagedIdentityDataPlaneAPIClientGetcredOptions) (ManagedIdentityDataPlaneAPIClientGetcredResponse, error) {
	var err error
	req, err := client.getcredCreateRequest(ctx, options)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientGetcredResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientGetcredResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK) {
		err = runtime.NewResponseError(httpResp)
		return ManagedIdentityDataPlaneAPIClientGetcredResponse{}, err
	}
	resp, err := client.getcredHandleResponse(httpResp)
	return resp, err
}

// getcredCreateRequest creates the Getcred request.
func (client *ManagedIdentityDataPlaneAPIClient) getcredCreateRequest(ctx context.Context, options *ManagedIdentityDataPlaneAPIClientGetcredOptions) (*policy.Request, error) {
	req, err := runtime.NewRequest(ctx, http.MethodGet, host)
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2024-01-01")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, nil
}

// getcredHandleResponse handles the Getcred response.
func (client *ManagedIdentityDataPlaneAPIClient) getcredHandleResponse(resp *http.Response) (ManagedIdentityDataPlaneAPIClientGetcredResponse, error) {
	result := ManagedIdentityDataPlaneAPIClientGetcredResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.CredentialsObject); err != nil {
		return ManagedIdentityDataPlaneAPIClientGetcredResponse{}, err
	}
	return result, nil
}

// Getcreds -
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2024-01-01
//   - options - ManagedIdentityDataPlaneAPIClientGetcredsOptions contains the optional parameters for the ManagedIdentityDataPlaneAPIClient.Getcreds
//     method.
func (client *ManagedIdentityDataPlaneAPIClient) Getcreds(ctx context.Context, credRequest CredRequestDefinition, options *ManagedIdentityDataPlaneAPIClientGetcredsOptions) (ManagedIdentityDataPlaneAPIClientGetcredsResponse, error) {
	var err error
	req, err := client.getcredsCreateRequest(ctx, credRequest, options)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK) {
		err = runtime.NewResponseError(httpResp)
		return ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, err
	}
	resp, err := client.getcredsHandleResponse(httpResp)
	return resp, err
}

// getcredsCreateRequest creates the Getcreds request.
func (client *ManagedIdentityDataPlaneAPIClient) getcredsCreateRequest(ctx context.Context, credRequest CredRequestDefinition, options *ManagedIdentityDataPlaneAPIClientGetcredsOptions) (*policy.Request, error) {
	req, err := runtime.NewRequest(ctx, http.MethodPost, host)
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2024-01-01")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	if err := runtime.MarshalAsJSON(req, credRequest); err != nil {
		return nil, err
	}
	return req, nil
}

// getcredsHandleResponse handles the Getcreds response.
func (client *ManagedIdentityDataPlaneAPIClient) getcredsHandleResponse(resp *http.Response) (ManagedIdentityDataPlaneAPIClientGetcredsResponse, error) {
	result := ManagedIdentityDataPlaneAPIClientGetcredsResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.CredentialsObject); err != nil {
		return ManagedIdentityDataPlaneAPIClientGetcredsResponse{}, err
	}
	return result, nil
}

// Moveidentity -
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2024-01-01
//   - options - ManagedIdentityDataPlaneAPIClientMoveidentityOptions contains the optional parameters for the ManagedIdentityDataPlaneAPIClient.Moveidentity
//     method.
func (client *ManagedIdentityDataPlaneAPIClient) Moveidentity(ctx context.Context, moveRequestBody MoveRequestBodyDefinition, options *ManagedIdentityDataPlaneAPIClientMoveidentityOptions) (ManagedIdentityDataPlaneAPIClientMoveidentityResponse, error) {
	var err error
	req, err := client.moveidentityCreateRequest(ctx, moveRequestBody, options)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientMoveidentityResponse{}, err
	}
	httpResp, err := client.internal.Pipeline().Do(req)
	if err != nil {
		return ManagedIdentityDataPlaneAPIClientMoveidentityResponse{}, err
	}
	if !runtime.HasStatusCode(httpResp, http.StatusOK) {
		err = runtime.NewResponseError(httpResp)
		return ManagedIdentityDataPlaneAPIClientMoveidentityResponse{}, err
	}
	resp, err := client.moveidentityHandleResponse(httpResp)
	return resp, err
}

// moveidentityCreateRequest creates the Moveidentity request.
func (client *ManagedIdentityDataPlaneAPIClient) moveidentityCreateRequest(ctx context.Context, moveRequestBody MoveRequestBodyDefinition, options *ManagedIdentityDataPlaneAPIClientMoveidentityOptions) (*policy.Request, error) {
	urlPath := "/proxy/move"
	req, err := runtime.NewRequest(ctx, http.MethodPost, runtime.JoinPaths(host, urlPath))
	if err != nil {
		return nil, err
	}
	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2024-01-01")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header["Accept"] = []string{"application/json"}
	if err := runtime.MarshalAsJSON(req, moveRequestBody); err != nil {
		return nil, err
	}
	return req, nil
}

// moveidentityHandleResponse handles the Moveidentity response.
func (client *ManagedIdentityDataPlaneAPIClient) moveidentityHandleResponse(resp *http.Response) (ManagedIdentityDataPlaneAPIClientMoveidentityResponse, error) {
	result := ManagedIdentityDataPlaneAPIClientMoveidentityResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.MoveIdentityResponse); err != nil {
		return ManagedIdentityDataPlaneAPIClientMoveidentityResponse{}, err
	}
	return result, nil
}
