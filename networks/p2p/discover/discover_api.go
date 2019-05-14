// Copyright 2019 The klaytn Authors
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

package discover

import (
	"errors"
)

func (tab *Table) Name() string { return "TableDiscovery" }

// CreateUpdateNode inserts - potentially overwriting - a node into the peer database.
func (tab *Table) CreateUpdateNode(n *Node) error {
	return tab.db.updateNode(n)
}

// GetNode returns a node which has id in peer database.
func (tab *Table) GetNode(id NodeID) (*Node, error) {
	node := tab.db.node(id)
	if node == nil {
		return nil, errors.New("failed to retrieve the node with the given id")
	}
	return node, nil
}

// GetBucketEntries returns nodes in peer databases.
func (tab *Table) GetBucketEntries() []*Node {
	var ret []*Node
	tab.storagesMu.RLock()
	defer tab.storagesMu.RUnlock()
	for _, bs := range tab.storages {
		ret = append(ret, bs.getBucketEntries()...)
	}
	return ret
}

// GetBucketEntries returns nodes which place in replacements in peer databases.
func (tab *Table) GetReplacements() []*Node {
	var ret []*Node
	tab.storagesMu.RLock()
	defer tab.storagesMu.RUnlock()
	for _, bs := range tab.storages {
		if _, ok := bs.(*KademliaStorage); ok { // TODO-Klaytn-Node Are there any change SimpleStorage use this method?
			ret = append(ret, bs.(*KademliaStorage).getReplacements()...)
		}
	}
	return ret
}

// DeleteNode deletes node which has id in peer database.
func (tab *Table) DeleteNode(id NodeID) error {
	return tab.db.deleteNode(id)
}
