//go:build unit

package store

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/Azure/msi-dataplane/internal/test"
	"github.com/Azure/msi-dataplane/pkg/dataplane"
	"github.com/Azure/msi-dataplane/pkg/dataplane/swagger"
	mock "github.com/Azure/msi-dataplane/pkg/store/mock_kvclient"
	"go.uber.org/mock/gomock"
)

const mockSecretName = "test"

var errMock = errors.New("client error")

func TestDeleteCredentialsObject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		goMockCall    func(kvClient *mock.MockKeyVaultClient)
		expectedError error
	}{
		{
			name: "Returns success when kv client successfully deletes the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().DeleteSecret(gomock.Any(), mockSecretName, gomock.Any()).Return(azsecrets.DeleteSecretResponse{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Returns kv client error when kv client fails to delete the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().DeleteSecret(gomock.Any(), mockSecretName, gomock.Any()).Return(azsecrets.DeleteSecretResponse{}, errMock)
			},
			expectedError: errMock,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			kvClient := mock.NewMockKeyVaultClient(mockCtrl)
			tc.goMockCall(kvClient)

			kvStore := NewMsiKeyVaultStore(kvClient)
			if err := kvStore.DeleteSecret(context.Background(), mockSecretName); !errors.Is(err, tc.expectedError) {
				t.Errorf("Expected %s but got: %s", tc.expectedError, err)
			}
		})
	}
}

func TestGetCredentialsObject(t *testing.T) {
	t.Parallel()

	bogusValue := test.Bogus
	testCredentialsObject := dataplane.CredentialsObject{
		Values: swagger.CredentialsObject{
			ClientSecret: &bogusValue,
		},
	}
	testCredentialsObjectBuffer, err := testCredentialsObject.Values.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to encode test credentials object: %s", err)
	}
	testCredentialsObjectString := string(testCredentialsObjectBuffer)
	enabled := true
	expires := time.Now()
	notBefore := time.Now()
	testGetSecretResponse := azsecrets.GetSecretResponse{
		Secret: azsecrets.Secret{
			Value: &testCredentialsObjectString,
			Attributes: &azsecrets.SecretAttributes{
				Enabled:   &enabled,
				Expires:   &expires,
				NotBefore: &notBefore,
			},
		},
	}

	testCases := []struct {
		name          string
		goMockCall    func(kvClient *mock.MockKeyVaultClient)
		expectedError error
	}{
		{
			name: "Returns success when kv client successfully gets the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().GetSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(testGetSecretResponse, nil)
			},
			expectedError: nil,
		},
		{
			name: "Returns kv client error when kv client fails to get the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().GetSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(azsecrets.GetSecretResponse{}, errMock)
			},
			expectedError: errMock,
		},
		{
			name: "Returns error when secret value is nil",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				resp := testGetSecretResponse
				resp.Secret.Value = nil
				kvClient.EXPECT().GetSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			expectedError: errNilSecretValue,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			kvClient := mock.NewMockKeyVaultClient(mockCtrl)
			tc.goMockCall(kvClient)

			kvStore := NewMsiKeyVaultStore(kvClient)
			response, err := kvStore.GetCredentialsObject(context.Background(), mockSecretName)
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("Expected error %s but got: %s", tc.expectedError, err)
			}
			if err == nil {
				if !reflect.DeepEqual(testCredentialsObject, response.CredentialsObject) {
					t.Errorf("Expected credentials object %+v\n but got: %+v", testCredentialsObject, response.CredentialsObject)
				}
				if response.Properties.Enabled != enabled {
					t.Errorf("Expected enabled %t but got: %t", enabled, response.Properties.Enabled)
				}
				if !response.Properties.Expires.Equal(expires) {
					t.Errorf("Expected expires %s but got: %s", expires, response.Properties.Expires)
				}
				if !response.Properties.NotBefore.Equal(notBefore) {
					t.Errorf("Expected notBefore %s but got: %s", notBefore, response.Properties.NotBefore)
				}
			}
		})
	}
}

func TestDeletedGetCredentialsObject(t *testing.T) {
	t.Parallel()

	bogusValue := test.Bogus
	testCredentialsObject := dataplane.CredentialsObject{
		Values: swagger.CredentialsObject{
			ClientSecret: &bogusValue,
		},
	}
	testCredentialsObjectBuffer, err := testCredentialsObject.Values.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to encode test credentials object: %s", err)
	}
	testCredentialsObjectString := string(testCredentialsObjectBuffer)
	deletedDate := time.Now()
	recoveryLevel := "Purgable"
	testGetDeletedSecretResponse := azsecrets.GetDeletedSecretResponse{
		DeletedSecret: azsecrets.DeletedSecret{
			Value: &testCredentialsObjectString,
			Attributes: &azsecrets.SecretAttributes{
				RecoveryLevel: &recoveryLevel,
			},
			DeletedDate: &deletedDate,
		},
	}

	testCases := []struct {
		name          string
		goMockCall    func(kvClient *mock.MockKeyVaultClient)
		expectedError error
	}{
		{
			name: "Returns success when kv client successfully gets the deleted secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().GetDeletedSecret(gomock.Any(), gomock.Any(), gomock.Any()).Return(testGetDeletedSecretResponse, nil)
			},
			expectedError: nil,
		},
		{
			name: "Returns kv client error when kv client fails to get the deleted secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().GetDeletedSecret(gomock.Any(), gomock.Any(), gomock.Any()).Return(azsecrets.GetDeletedSecretResponse{}, errMock)
			},
			expectedError: errMock,
		},
		{
			name: "Returns error when secret value is nil",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				resp := testGetDeletedSecretResponse
				resp.DeletedSecret.Value = nil
				kvClient.EXPECT().GetDeletedSecret(gomock.Any(), gomock.Any(), gomock.Any()).Return(resp, nil)
			},
			expectedError: errNilSecretValue,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			kvClient := mock.NewMockKeyVaultClient(mockCtrl)
			tc.goMockCall(kvClient)

			kvStore := NewMsiKeyVaultStore(kvClient)
			response, err := kvStore.GetDeletedCredentialsObject(context.Background(), mockSecretName)
			if !errors.Is(err, tc.expectedError) {
				t.Errorf("Expected error %s but got: %s", tc.expectedError, err)
			}
			if err == nil {
				if !reflect.DeepEqual(testCredentialsObject, response.CredentialsObject) {
					t.Errorf("Expected credentials object %+v\n but got: %+v", testCredentialsObject, response.CredentialsObject)
				}
				if !response.Properties.DeletedDate.Equal(deletedDate) {
					t.Errorf("Expected deletedDate %s but got: %s", deletedDate, response.Properties.DeletedDate)
				}
			}
		})
	}
}

func TestNewDeletedListSecretsPager(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectedPager *runtime.Pager[azsecrets.ListDeletedSecretPropertiesResponse]
		goMockCall    func(kvClient *mock.MockKeyVaultClient)
	}{
		{
			name: "Returns a pager",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().NewListDeletedSecretPropertiesPager(gomock.Any()).Return(&runtime.Pager[azsecrets.ListDeletedSecretPropertiesResponse]{})
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			kvClient := mock.NewMockKeyVaultClient(mockCtrl)
			tc.goMockCall(kvClient)

			kvStore := NewMsiKeyVaultStore(kvClient)
			if pager := kvStore.GetDeletedSecretObjectPager(); pager == nil {
				t.Error("Expected pager but got nil")
			}
		})
	}
}

func TestNewListSecretsPager(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectedPager *runtime.Pager[azsecrets.ListSecretPropertiesResponse]
		goMockCall    func(kvClient *mock.MockKeyVaultClient)
	}{
		{
			name: "Returns a pager",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().NewListSecretPropertiesPager(gomock.Any()).Return(&runtime.Pager[azsecrets.ListSecretPropertiesResponse]{})
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			kvClient := mock.NewMockKeyVaultClient(mockCtrl)
			tc.goMockCall(kvClient)

			kvStore := NewMsiKeyVaultStore(kvClient)
			if pager := kvStore.GetSecretObjectPager(); pager == nil {
				t.Error("Expected pager but got nil")
			}
		})
	}
}

func TestPurgeDeletedSecret(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		goMockCall    func(kvClient *mock.MockKeyVaultClient)
		expectedError error
	}{
		{
			name: "Returns success when kv client successfully purges the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().PurgeDeletedSecret(gomock.Any(), mockSecretName, gomock.Any()).Return(azsecrets.PurgeDeletedSecretResponse{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Returns kv client error when kv client fails to purge the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().PurgeDeletedSecret(gomock.Any(), mockSecretName, gomock.Any()).Return(azsecrets.PurgeDeletedSecretResponse{}, errMock)
			},
			expectedError: errMock,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			kvClient := mock.NewMockKeyVaultClient(mockCtrl)
			tc.goMockCall(kvClient)

			kvStore := NewMsiKeyVaultStore(kvClient)
			if err := kvStore.PurgeDeletedSecretObject(context.Background(), mockSecretName); !errors.Is(err, tc.expectedError) {
				t.Errorf("Expected %s but got: %s", tc.expectedError, err)
			}
		})
	}
}

func TestSetCredentialsObject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		goMockCall    func(kvClient *mock.MockKeyVaultClient)
		expectedError error
	}{
		{
			name: "Returns success when kv client successfully sets the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().SetSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(azsecrets.SetSecretResponse{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Returns kv client error when kv client fails to set the secret",
			goMockCall: func(kvClient *mock.MockKeyVaultClient) {
				kvClient.EXPECT().SetSecret(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(azsecrets.SetSecretResponse{}, errMock)
			},
			expectedError: errMock,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			kvClient := mock.NewMockKeyVaultClient(mockCtrl)
			tc.goMockCall(kvClient)

			kvStore := NewMsiKeyVaultStore(kvClient)
			props := SecretProperties{
				Name: mockSecretName,
			}
			if err := kvStore.SetCredentialsObject(context.Background(), props, dataplane.CredentialsObject{}); !errors.Is(err, tc.expectedError) {
				t.Errorf("Expected %s but got: %s", tc.expectedError, err)
			}
		})
	}
}
