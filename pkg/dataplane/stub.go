package dataplane

import (
	"bytes"
	"hash/fnv"
	"io"
	"net/http"

	"github.com/Azure/msi-dataplane/internal/swagger"
)

// Hash table to store identities for dataplane
// Key is a hash of certain fields in the credentials object
type identityHashMap map[uint64]*CredentialsObject

// Stub is a transport that mocks the Managed Identity Dataplane API
// TODO - add support for system-assigned managed identities
type stub struct {
	// Key is a hash of the resource IDs
	userAssignedIdentities map[uint64]*CredentialsObject
}

func NewStub(creds []*CredentialsObject) *stub {
	userAssignedIdentities := make(map[uint64]*CredentialsObject)
	for _, identity := range creds {
		if identity != nil && identity.IsUserAssigned() {
			resourceIDs := make([]string, 0)
			for _, uaMSI := range identity.ExplicitIdentities {
				if uaMSI != nil && uaMSI.ResourceID != nil {
					resourceIDs = append(resourceIDs, *uaMSI.ResourceID)
				}
			}
			if len(resourceIDs) != 0 {
				hash := hashResourceIDs(resourceIDs)
				userAssignedIdentities[hash] = identity
			}
		}
	}
	return &stub{userAssignedIdentities: userAssignedIdentities}
}

// Implement transport interface
// Per MSI team's documentation, POST is for user-assigned MSI and GET is for system-assigned MSI
// https://eng.ms/docs/products/arm/rbac/managed_identities/msionboardinguserassigned
func (s stub) Do(req *http.Request) (*http.Response, error) {
	var response *http.Response
	var err error

	switch req.Method {
	case http.MethodPost:
		// User-assigned managed identities request
		response, err = s.post(req)
	default:
		response = &http.Response{
			StatusCode: http.StatusNotImplemented,
			Body:       io.NopCloser(bytes.NewBufferString("")),
		}
	}

	return response, err
}

func (s stub) post(req *http.Request) (*http.Response, error) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return &http.Response{StatusCode: http.StatusInternalServerError}, err
	}

	credRequestDefinition := &swagger.CredRequestDefinition{}
	if err := credRequestDefinition.UnmarshalJSON(bodyBytes); err != nil {
		return &http.Response{StatusCode: http.StatusInternalServerError}, err
	}

	identityList := credRequestDefinition.IdentityIDs
	resourceIds := make([]string, 0)
	for _, identity := range identityList {
		if identity != nil {
			resourceIds = append(resourceIds, *identity)
		}
	}

	hash := hashResourceIDs(resourceIds)
	creds, ok := s.userAssignedIdentities[hash]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	}

	// Marshal the credentials object into the body
	credBytes, err := creds.MarshalJSON()
	if err != nil {
		return &http.Response{StatusCode: http.StatusInternalServerError}, err
	}
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(credBytes)),
	}

	return response, nil
}

func hashResourceIDs(resourceIDs []string) uint64 {
	h := fnv.New64a()

	for _, id := range resourceIDs {
		h.Write([]byte(id))
	}

	return h.Sum64()
}

/*
func NewStub(cloud string) (*ManagedIdentityStub, error) {
	plOpts := runtime.PipelineOptions{
		PerCall: []policy.Policy{
			&injectIdentityURLPolicy{
				msiHost: getMsiHost(cloud),
			},
		},
	}

	clientOpts := &policy.ClientOptions{
		Transport: &stubTransporter{},
	}

	azCoreClient, err := azcore.NewClient("managedidentitydataplane.APIClient", moduleVersion, plOpts, clientOpts)
	if err != nil {
		return nil, err
	}
	swaggerClient := swagger.NewSwaggerClient(azCoreClient)
	stub := &ManagedIdentityStub{
		swaggerClient: swaggerClient,
	}
	return stub, nil
}

func (s *ManagedIdentityStub) GetUserAssignedIdentities(ctx context.Context, request UserAssignedMSIRequest) (*UserAssignedIdentities, error) {
	ctx = context.WithValue(ctx, identityURLKey, request.IdentityURL)
	_, err := s.swaggerClient.Getcreds(ctx, swagger.CredRequestDefinition{
		IdentityIDs: []*string{test.StringPtr("test")},
	}, nil)

	return nil, err
}
*/
