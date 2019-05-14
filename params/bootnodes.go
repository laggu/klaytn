// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
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
// This file is derived from params/bootnodes.go (2018/06/04).
// Modified and improved for the klaytn development.

package params

import "github.com/ground-x/klaytn/networks/p2p"

type bootnodesByTypes struct {
	Addrs []string
}

// MainnetBootnodes are the kni URLs of the P2P bootstrap nodes running on
// the Klaytn main network.
var MainnetBootnodes = []string{
	// TODO-Klaytn-Bootnode : Klaytn BootNode should be set. Now for only test.
	//"kni://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303", // IE
	//"kni://3f1d12044546b76342d59d4a05532c14b85aa669704bfe1f864fe079415aa2c02d743e03218e57a33fb94523adb54032871a6c51b2cc5514cb7c7e35b3ed0a99@13.93.211.84:30303",  // US-WEST
}

// BaobabBootnodes are the kni URLs of the PN's P2P bootstrap nodes running on the Baobab test network.
var BaobabBootnodes = map[p2p.ConnType]bootnodesByTypes{
	p2p.CONSENSUSNODE: {
		[]string{
			"kni://0b34bc04018ff4b4079d7734d2788cc6d73fc1e699321d09f7ad6f49825f054251fd35b0cbb4003b7fbf4825c6318e82cb1d1514d7b1d294de9f6f5f70e8eae9@bn1.baobab.klaytn.net:32323?discport=32323&ntype=bn",
		},
	},
	p2p.ENDPOINTNODE: {
		[]string{
			"kni://f22ebd1fc610686b5749a4e4ec4da9ba4647fd0bdd8b7058e1c58221e06d71686b519da522fdb930bcde1bf0339f73bdade429123b787b37199c6605f2efa025@bn2.baobab.klaytn.net:32323?discport=32323&ntype=bn",
		},
	},
	p2p.PROXYNODE: {
		[]string{
			"kni://f22ebd1fc610686b5749a4e4ec4da9ba4647fd0bdd8b7058e1c58221e06d71686b519da522fdb930bcde1bf0339f73bdade429123b787b37199c6605f2efa025@bn2.baobab.klaytn.net:32323?discport=32323&ntype=bn",
		},
	},
}
