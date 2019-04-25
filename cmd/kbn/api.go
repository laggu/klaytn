// Copyright 2018 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"net"
)

type BootnodeAPI struct {
	bn *BN
}

func NewBootnodeAPI(b *BN) *BootnodeAPI {
	return &BootnodeAPI{bn: b}
}

func (api *BootnodeAPI) Name() string {
	return api.bn.Name()
}

func (api *BootnodeAPI) Resolve(target discover.NodeID) *discover.Node {
	return api.bn.Resolve(target)
}

func (api *BootnodeAPI) Lookup(target discover.NodeID) []*discover.Node {
	return api.bn.Lookup(target)
}

func (api *BootnodeAPI) ReadRandomNodes() []*discover.Node {
	var buf []*discover.Node
	api.bn.ReadRandomNodes(buf)
	return buf
}

func (api *BootnodeAPI) CreateUpdateNode(id discover.NodeID, ip net.IP, udpPort, tcpPort uint16) error {
	return api.bn.CreateUpdateNode(id, ip, udpPort, tcpPort)
}

func (api *BootnodeAPI) GetNode(id discover.NodeID) (*discover.Node, error) {
	return api.bn.GetNode(id)
}

func (api *BootnodeAPI) GetTableEntries() []*discover.Node {
	return api.bn.GetTableEntries()
}

func (api *BootnodeAPI) GetTableReplacements() []*discover.Node {
	return api.bn.GetTableReplacements()
}

func (api *BootnodeAPI) DeleteNode(id discover.NodeID) error {
	return api.bn.DeleteNode(id)
}
