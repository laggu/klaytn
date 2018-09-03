package ranger

import "github.com/ground-x/go-gxplatform/common"

// PublicGXPAPI provides an API to access GXPlatform full node-related
// information.
type PublicRangerAPI struct {
	ranger *Ranger
}

// NewPublicGXPAPI creates a new GXP protocol API for full nodes.
func NewPublicRangerAPI(e *Ranger) *PublicRangerAPI {
	return &PublicRangerAPI{e}
}

func(pr *PublicRangerAPI) Accounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range pr.ranger.accountManager.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// Coinbase is the address that mining rewards will be send to
func (pr *PublicRangerAPI) Coinbase() (common.Address, error) {
	return pr.ranger.Coinbase()
}
