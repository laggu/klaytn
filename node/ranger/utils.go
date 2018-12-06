package ranger

import (
	"github.com/ground-x/go-gxplatform/common/hexutil"
	"github.com/ground-x/go-gxplatform/crypto"
)

// PrivateAdminAPI is the collection of administrative API methods exposed only
// over a secure RPC channel.
type PrivateAdminAPI struct {
	ranger *Ranger // Node interfaced by this API
}

// NewPrivateAdminAPI creates a new API definition for the private admin methods
// of the node itself.
func NewPrivateAdminAPI(rn *Ranger) *PrivateAdminAPI {
	return &PrivateAdminAPI{ranger: rn}
}

// PublicAdminAPI is the collection of administrative API methods exposed over
// both secure and unsecure RPC channels.
type PublicAdminAPI struct {
	ranger *Ranger // Node interfaced by this API
}

// NewPublicAdminAPI creates a new API definition for the public admin methods
// of the node itself.
func NewPublicAdminAPI(rn *Ranger) *PublicAdminAPI {
	return &PublicAdminAPI{ranger: rn}
}

// PublicDebugAPI is the collection of debugging related API methods exposed over
// both secure and unsecure RPC channels.
type PublicDebugAPI struct {
	ranger *Ranger // Node interfaced by this API
}

// NewPublicDebugAPI creates a new API definition for the public debug methods
// of the node itself.
func NewPublicDebugAPI(rn *Ranger) *PublicDebugAPI {
	return &PublicDebugAPI{ranger: rn}
}

// PublicWeb3API offers helper utils
type PublicWeb3API struct {
	stack *Ranger
}

// NewPublicWeb3API creates a new Web3Service instance
func NewPublicWeb3API(stack *Ranger) *PublicWeb3API {
	return &PublicWeb3API{stack}
}

// ClientVersion returns the node name
func (s *PublicWeb3API) ClientVersion() string {
	return "RangerNode v1.0"
}

// Sha3 applies the ethereum sha3 implementation on the input.
// It assumes the input is hex encoded.
func (s *PublicWeb3API) Sha3(input hexutil.Bytes) hexutil.Bytes {
	return crypto.Keccak256(input)
}
