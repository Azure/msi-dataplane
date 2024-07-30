// Code generated by MockGen. DO NOT EDIT.
// Source: client.go
//
// Generated by this command:
//
//	mockgen -destination=mock_swagger_client/zz_generated_mocks.go -package=mock -source=client.go
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	swagger "github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
	gomock "go.uber.org/mock/gomock"
)

// MockmsiClient is a mock of msiClient interface.
type MockmsiClient struct {
	ctrl     *gomock.Controller
	recorder *MockmsiClientMockRecorder
}

// MockmsiClientMockRecorder is the mock recorder for MockmsiClient.
type MockmsiClientMockRecorder struct {
	mock *MockmsiClient
}

// NewMockmsiClient creates a new mock instance.
func NewMockmsiClient(ctrl *gomock.Controller) *MockmsiClient {
	mock := &MockmsiClient{ctrl: ctrl}
	mock.recorder = &MockmsiClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockmsiClient) EXPECT() *MockmsiClientMockRecorder {
	return m.recorder
}

// Getcreds mocks base method.
func (m *MockmsiClient) Getcreds(ctx context.Context, credRequest swagger.CredRequestDefinition, options *swagger.ManagedIdentityDataPlaneAPIClientGetcredsOptions) (swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Getcreds", ctx, credRequest, options)
	ret0, _ := ret[0].(swagger.ManagedIdentityDataPlaneAPIClientGetcredsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Getcreds indicates an expected call of Getcreds.
func (mr *MockmsiClientMockRecorder) Getcreds(ctx, credRequest, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Getcreds", reflect.TypeOf((*MockmsiClient)(nil).Getcreds), ctx, credRequest, options)
}
