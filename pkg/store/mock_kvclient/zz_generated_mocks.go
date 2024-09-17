// Code generated by MockGen. DO NOT EDIT.
// Source: kvclient.go
//
// Generated by this command:
//
//	mockgen -destination=mock_kvclient/zz_generated_mocks.go -package=mock -source=kvclient.go
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	runtime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azsecrets "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	gomock "go.uber.org/mock/gomock"
)

// MockKeyVaultClient is a mock of KeyVaultClient interface.
type MockKeyVaultClient struct {
	ctrl     *gomock.Controller
	recorder *MockKeyVaultClientMockRecorder
}

// MockKeyVaultClientMockRecorder is the mock recorder for MockKeyVaultClient.
type MockKeyVaultClientMockRecorder struct {
	mock *MockKeyVaultClient
}

// NewMockKeyVaultClient creates a new mock instance.
func NewMockKeyVaultClient(ctrl *gomock.Controller) *MockKeyVaultClient {
	mock := &MockKeyVaultClient{ctrl: ctrl}
	mock.recorder = &MockKeyVaultClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeyVaultClient) EXPECT() *MockKeyVaultClientMockRecorder {
	return m.recorder
}

// DeleteSecret mocks base method.
func (m *MockKeyVaultClient) DeleteSecret(ctx context.Context, name string, options *azsecrets.DeleteSecretOptions) (azsecrets.DeleteSecretResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSecret", ctx, name, options)
	ret0, _ := ret[0].(azsecrets.DeleteSecretResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteSecret indicates an expected call of DeleteSecret.
func (mr *MockKeyVaultClientMockRecorder) DeleteSecret(ctx, name, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSecret", reflect.TypeOf((*MockKeyVaultClient)(nil).DeleteSecret), ctx, name, options)
}

// GetDeletedSecret mocks base method.
func (m *MockKeyVaultClient) GetDeletedSecret(ctx context.Context, name string, options *azsecrets.GetDeletedSecretOptions) (azsecrets.GetDeletedSecretResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeletedSecret", ctx, name, options)
	ret0, _ := ret[0].(azsecrets.GetDeletedSecretResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeletedSecret indicates an expected call of GetDeletedSecret.
func (mr *MockKeyVaultClientMockRecorder) GetDeletedSecret(ctx, name, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeletedSecret", reflect.TypeOf((*MockKeyVaultClient)(nil).GetDeletedSecret), ctx, name, options)
}

// GetSecret mocks base method.
func (m *MockKeyVaultClient) GetSecret(ctx context.Context, name, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecret", ctx, name, version, options)
	ret0, _ := ret[0].(azsecrets.GetSecretResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecret indicates an expected call of GetSecret.
func (mr *MockKeyVaultClientMockRecorder) GetSecret(ctx, name, version, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockKeyVaultClient)(nil).GetSecret), ctx, name, version, options)
}

// NewListDeletedSecretPropertiesPager mocks base method.
func (m *MockKeyVaultClient) NewListDeletedSecretPropertiesPager(options *azsecrets.ListDeletedSecretPropertiesOptions) *runtime.Pager[azsecrets.ListDeletedSecretPropertiesResponse] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewListDeletedSecretPropertiesPager", options)
	ret0, _ := ret[0].(*runtime.Pager[azsecrets.ListDeletedSecretPropertiesResponse])
	return ret0
}

// NewListDeletedSecretPropertiesPager indicates an expected call of NewListDeletedSecretPropertiesPager.
func (mr *MockKeyVaultClientMockRecorder) NewListDeletedSecretPropertiesPager(options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewListDeletedSecretPropertiesPager", reflect.TypeOf((*MockKeyVaultClient)(nil).NewListDeletedSecretPropertiesPager), options)
}

// NewListSecretPropertiesPager mocks base method.
func (m *MockKeyVaultClient) NewListSecretPropertiesPager(options *azsecrets.ListSecretPropertiesOptions) *runtime.Pager[azsecrets.ListSecretPropertiesResponse] {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewListSecretPropertiesPager", options)
	ret0, _ := ret[0].(*runtime.Pager[azsecrets.ListSecretPropertiesResponse])
	return ret0
}

// NewListSecretPropertiesPager indicates an expected call of NewListSecretPropertiesPager.
func (mr *MockKeyVaultClientMockRecorder) NewListSecretPropertiesPager(options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewListSecretPropertiesPager", reflect.TypeOf((*MockKeyVaultClient)(nil).NewListSecretPropertiesPager), options)
}

// PurgeDeletedSecret mocks base method.
func (m *MockKeyVaultClient) PurgeDeletedSecret(ctx context.Context, name string, options *azsecrets.PurgeDeletedSecretOptions) (azsecrets.PurgeDeletedSecretResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PurgeDeletedSecret", ctx, name, options)
	ret0, _ := ret[0].(azsecrets.PurgeDeletedSecretResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PurgeDeletedSecret indicates an expected call of PurgeDeletedSecret.
func (mr *MockKeyVaultClientMockRecorder) PurgeDeletedSecret(ctx, name, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PurgeDeletedSecret", reflect.TypeOf((*MockKeyVaultClient)(nil).PurgeDeletedSecret), ctx, name, options)
}

// SetSecret mocks base method.
func (m *MockKeyVaultClient) SetSecret(ctx context.Context, name string, parameters azsecrets.SetSecretParameters, options *azsecrets.SetSecretOptions) (azsecrets.SetSecretResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetSecret", ctx, name, parameters, options)
	ret0, _ := ret[0].(azsecrets.SetSecretResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetSecret indicates an expected call of SetSecret.
func (mr *MockKeyVaultClientMockRecorder) SetSecret(ctx, name, parameters, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSecret", reflect.TypeOf((*MockKeyVaultClient)(nil).SetSecret), ctx, name, parameters, options)
}
