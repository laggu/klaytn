// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from node/api.go (2018/06/04).
// Modified and improved for the klaytn development.

package node

// PrivateAdminAPI is the collection of administrative API methods exposed only
// over a secure RPC channel.
type PublicServiceChainAPI struct {
	node *Node // Node interfaced by this API
}

// NewPrivateAdminAPI creates a new API definition for the private admin methods
// of the node itself.
func NewPublicServiceChainAPI(node *Node) *PublicServiceChainAPI {
	return &PublicServiceChainAPI{node: node}
}

// AddPeerOnParentChain requests connecting to a remote parent chain node, and also maintaining the new
// connection at all times, even reconnecting if it is lost.
func (api *PublicServiceChainAPI) AddPeerOnParentChain(url string) (bool, error) {
	// Make sure the server is running, fail otherwise
	server := api.node.Server()
	if server == nil {
		return false, ErrNodeStopped
	}
	// TODO-Klaytn Refactoring this to check whether the url is valid or not by dialing and return it.
	if _, err := addPeerInternal(server, url, true); err != nil {
		return false, err
	} else {
		return true, nil
	}
}
