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

package discover

// Test Scenarios
// 1. Unit tests
// When lookup method is called, it always wait for a doRequest. Test for this scenario is required.

import (
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/crypto"
	"net"
	"testing"
)

type preminedSimpleTestnet struct {
	target     NodeID
	targetSha  common.Hash
	knownNodes []*Node
}

func (tn *preminedSimpleTestnet) findnode(toid NodeID, toaddr *net.UDPAddr, target NodeID) ([]*Node, error) {
	// response all known nodes without udp transport
	return tn.knownNodes, nil
}

func (tn *preminedSimpleTestnet) close() {}

func (tn *preminedSimpleTestnet) waitping(from NodeID) error { return nil }

func (tn *preminedSimpleTestnet) ping(toID NodeID, toaddr *net.UDPAddr) error { return nil }

// This is the test network for the Lookup test.
var simpleLookupTestnet = &preminedSimpleTestnet{
	target:    MustHexID("166aea4f556532c6d34e8b740e5d314af7e9ac0ca79833bd751d6b665f12dfd38ec563c363b32f02aef4a80b44fd3def94612d497b99cb5f17fd24de454927ec"),
	targetSha: common.Hash{0x5c, 0x94, 0x4e, 0xe5, 0x1c, 0x5a, 0xe9, 0xf7, 0x2a, 0x95, 0xec, 0xcb, 0x8a, 0xed, 0x3, 0x74, 0xee, 0xcb, 0x51, 0x19, 0xd7, 0x20, 0xcb, 0xea, 0x68, 0x13, 0xe8, 0xe0, 0xd6, 0xad, 0x92, 0x61},
	knownNodes: []*Node{
		{
			ID:  MustHexID("f1ae93157cc48c2075dd5868fbf523e79e06caf4b8198f352f6e526680b78ff4227263de92612f7d63472bd09367bb92a636fff16fe46ccf41614f7a72495c2a"),
			IP:  net.ParseIP("10.10.2.1"),
			TCP: 12324,
			UDP: 12324,
		},
		{
			ID:  MustHexID("696ba1f0a9d55c59246f776600542a9e6432490f0cd78f8bb55a196918df2081a9b521c3c3ba48e465a75c10768807717f8f689b0b4adce00e1c75737552a178"),
			IP:  net.ParseIP("10.44.10.2"),
			TCP: 22222,
			UDP: 22324,
		},
		{
			ID:  MustHexID("d6d32178bdc38416f46ffb8b3ec9e4cb2cfff8d04dd7e4311a70e403cb62b10be1b447311b60b4f9ee221a8131fc2cbd45b96dd80deba68a949d467241facfa8"),
			IP:  net.ParseIP("192.111.22.3"),
			TCP: 32222,
			UDP: 34444,
		},
		{
			ID:  MustHexID("e3f88274d35cefdaabdf205afe0e80e936cc982b8e3e47a84ce664c413b29016a4fb4f3a3ebae0a2f79671f8323661ed462bf4390af94c424dc8ace0c301b90f"),
			IP:  net.ParseIP("192.168.128.2"),
			TCP: 42323,
			UDP: 42323,
		},
		{
			ID:  MustHexID("587f482d111b239c27c0cb89b51dd5d574db8efd8de14a2e6a1400c54d4567e77c65f89c1da52841212080b91604104768350276b6682f2f961cdaf4039581c7"),
			IP:  net.ParseIP("10.1.2.1"),
			TCP: 52326,
			UDP: 52323,
		},
		{
			ID:  MustHexID("0ddc736077da9a12ba410dc5ea63cbcbe7659dd08596485b2bff3435221f82c10d263efd9af938e128464be64a178b7cd22e19f400d5802f4c9df54bf89f2619"),
			IP:  net.ParseIP("172.126.0.1"),
			TCP: 62328,
			UDP: 62328,
		},
		{
			ID:  MustHexID("f253a2c354ee0e27cfcae786d726753d4ad24be6516b279a936195a487de4a59dbc296accf20463749ff55293263ed8c1b6365eecb248d44e75e9741c0d18205"),
			IP:  net.ParseIP("172.126.100.2"),
			TCP: 7333,
			UDP: 7332,
		},
		{
			ID:  MustHexID("f1168552c2efe541160f0909b0b4a9d6aeedcf595cdf0e9b165c97e3e197471a1ee6320e93389edfba28af6eaf10de98597ad56e7ab1b504ed762451996c3b98"),
			IP:  net.ParseIP("10.43.43.1"),
			TCP: 8333,
			UDP: 8333,
		},
		{
			ID:  MustHexID("7b9c1889ae916a5d5abcdfb0aaedcc9c6f9eb1c1a4f68d0c2d034fe79ac610ce917c3abc670744150fa891bfcd8ab14fed6983fca964de920aa393fa7b326748"),
			IP:  net.ParseIP("10.10.44.10"),
			TCP: 9044,
			UDP: 9044,
		},
		{
			ID:  MustHexID("3ea3d04a43a3dfb5ac11cffc2319248cf41b6279659393c2f55b8a0a5fc9d12581a9d97ef5d8ff9b5abf3321a290e8f63a4f785f450dc8a672aba3ba2ff4fdab"),
			IP:  net.ParseIP("192.16.8.2"),
			TCP: 10586,
			UDP: 10556,
		},
		{
			ID:  MustHexID("2fc897f05ae585553e5c014effd3078f84f37f9333afacffb109f00ca8e7a3373de810a3946be971cbccdfd40249f9fe7f322118ea459ac71acca85a1ef8b7f4"),
			IP:  net.ParseIP("127.0.0.1"),
			TCP: 12323,
			UDP: 12323,
		},
	},
}

func TestSimple_RetrieveNodes(t *testing.T) {
	self := nodeAtDistance(common.Hash{}, 0)
	conf := Config{
		udp:              simpleLookupTestnet,
		Id:               self.ID,
		Addr:             &net.UDPAddr{},
		Bootnodes:        nil,
		NodeDBPath:       "",
		MaxNeighborsNode: uint(len(simpleLookupTestnet.knownNodes)),
	}
	discv, _ := newSimple(&conf)
	s := discv.(*Simple)
	defer s.Close()

	// lookup on empty table returns no nodes
	if results := s.Lookup(simpleLookupTestnet.target); len(results) > 0 {
		t.Fatalf("lookup on empty table returned %d results: %#v", len(results), results)
	}
	// seed table with initial node (otherwise lookup will terminate immediately)
	seed := simpleLookupTestnet.knownNodes
	s.stuff(seed)
	sha := crypto.Keccak256Hash(simpleLookupTestnet.target[:])
	for i := range seed {
		results := s.RetrieveNodes(sha, i)
		// expect return i number of known nodes
		// TODO-Klaytn-Discovery: find a way to validate randomize return result
		if len(results) != s.bucketSize {
			t.Errorf("wrong number of results: got %d, want %d", len(results), i)
		}
		if hasDuplicates(results) {
			t.Errorf("result set contains duplicate entries")
		}
		// check result is contains the seed nodes
		for _, n := range results {
			if !contains(seed, n.ID) {
				t.Errorf("result set does not contain seed's node entrie")
			}
		}
	}

}

func TestSimple_bond(t *testing.T) {
	self := nodeAtDistance(common.Hash{}, 0)
	conf := Config{
		udp:              simpleLookupTestnet,
		Id:               self.ID,
		Addr:             &net.UDPAddr{},
		Bootnodes:        nil,
		NodeDBPath:       "",
		MaxNeighborsNode: uint(len(simpleLookupTestnet.knownNodes)),
	}
	discv, _ := newSimple(&conf)
	s := discv.(*Simple)
	defer s.Close()

	node := s.randNode(simpleLookupTestnet.knownNodes)
	n, err := s.bond(false, node.ID, &net.UDPAddr{}, 32323)
	if err != nil {
		t.Errorf("bond fail")
	}
	if n == nil {
		t.Errorf("empty result, want %s", node.ID)
	}
	if !contains(simpleLookupTestnet.knownNodes, n.ID) {
		t.Errorf("unknown node returned,")
	}

	if !s.hasBond(node.ID) {
		t.Error("wrong result: got false, want: true")
	}
}

func TestSimple_bondall(t *testing.T) {
	self := nodeAtDistance(common.Hash{}, 0)
	conf := Config{
		udp:              simpleLookupTestnet,
		Id:               self.ID,
		Addr:             &net.UDPAddr{},
		Bootnodes:        nil,
		NodeDBPath:       "",
		MaxNeighborsNode: uint(len(simpleLookupTestnet.knownNodes)),
	}
	discv, _ := newSimple(&conf)
	s := discv.(*Simple)
	defer s.Close()

	results := s.bondall(simpleLookupTestnet.knownNodes)

	if len(results) != len(simpleLookupTestnet.knownNodes) {
		t.Errorf("wrong result: got %d, want: %d", len(results), len(simpleLookupTestnet.knownNodes))
	}
	for _, n := range simpleLookupTestnet.knownNodes {
		if !s.hasBond(n.ID) {
			t.Errorf("wrong result, node %s has not bonded", n.ID)
		}
	}
}

func TestSimple_Resolve(t *testing.T) {
	self := nodeAtDistance(common.Hash{}, 0)
	conf := Config{
		udp:              simpleLookupTestnet,
		Id:               self.ID,
		Addr:             &net.UDPAddr{},
		Bootnodes:        nil,
		NodeDBPath:       "",
		MaxNeighborsNode: uint(len(simpleLookupTestnet.knownNodes)),
	}
	discv, _ := newSimple(&conf)
	s := discv.(*Simple)
	defer s.Close()

	seed := simpleLookupTestnet.knownNodes
	s.stuff(seed)

	// try to lookup random node from predefined known node
	targetNode := s.randNode(simpleLookupTestnet.knownNodes)

	result := s.Resolve(targetNode.ID)
	if result == nil {
		t.Fatalf("empty result: got empty, want: %s", targetNode.ID)
	}
	if result.ID != targetNode.ID {
		t.Errorf("wrong result: got %s, want: %s", result.ID, targetNode.ID)
	}
}

func TestSimple_Lookup(t *testing.T) {
	self := nodeAtDistance(common.Hash{}, 0)
	conf := Config{
		udp:              simpleLookupTestnet,
		Id:               self.ID,
		Addr:             &net.UDPAddr{},
		NodeDBPath:       "",
		MaxNeighborsNode: uint(len(simpleLookupTestnet.knownNodes) / 2),
		Bootnodes:        []*Node{simpleLookupTestnet.knownNodes[0]},
	}
	discv, _ := newSimple(&conf)
	s := discv.(*Simple)
	defer s.Close()

	seed := s.randNode(simpleLookupTestnet.knownNodes)
	s.stuff([]*Node{seed})

	results := s.Lookup(simpleLookupTestnet.target)
	if results == nil {
		t.Fatalf("empty result: got nil, want: %d length of slice", conf.MaxNeighborsNode)
	}
	if uint(len(results)) != conf.MaxNeighborsNode {
		t.Errorf("wrong result: got %d, want: %d", len(results), conf.MaxNeighborsNode)
	}
	if hasDuplicates(results) {
		t.Errorf("wrong result: found dupicated entry")
	}

}

// TODO-Klaytn-Discovery: add bump() unit test
// TODO-Klaytn-Discovery: add replacment and delete node unit test
