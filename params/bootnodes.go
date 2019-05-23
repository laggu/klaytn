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

// MainnetBootnodes are the URLs of bootnodes running on the Klaytn main network.
var MainnetBootnodes = map[p2p.ConnType]bootnodesByTypes{
	p2p.CONSENSUSNODE: {
		[]string{},
	},
	p2p.PROXYNODE: {
		[]string{
			"kni://18b36118cce093673499fc6e9aa196f047fe17a0de35b6f2a76a4557802f6abf9f89aa5e7330e93c9014b714b9df6378393611efe39aec9d3d831d6aa9d617ae@ston65.cypress.klaytn.net:32323?ntype=bn",
			"kni://63f1c96874da85140ecca3ce24875cb5ef28fa228bc3572e16f690db4a48fc8067502d2f6e8f0c66fb558276a5ada1e4906852c7ae42b0003e9f9f25d1e123b1@ston873.cypress.klaytn.net:32323?ntype=bn",
			"kni://94cc15e2014b86584908707de55800c0a2ea8a24dc5550dcb507043e4cf18ff04f21dc86ed17757dc63b1fa85bb418b901e5e24e4197ad4bbb0d96cd9389ed98@ston106.cypress.klaytn.net:32323?ntype=bn",
		},
	},
	p2p.ENDPOINTNODE: {
		[]string{
			"kni://18b36118cce093673499fc6e9aa196f047fe17a0de35b6f2a76a4557802f6abf9f89aa5e7330e93c9014b714b9df6378393611efe39aec9d3d831d6aa9d617ae@ston65.cypress.klaytn.net:32323?ntype=bn",
			"kni://63f1c96874da85140ecca3ce24875cb5ef28fa228bc3572e16f690db4a48fc8067502d2f6e8f0c66fb558276a5ada1e4906852c7ae42b0003e9f9f25d1e123b1@ston873.cypress.klaytn.net:32323?ntype=bn",
			"kni://94cc15e2014b86584908707de55800c0a2ea8a24dc5550dcb507043e4cf18ff04f21dc86ed17757dc63b1fa85bb418b901e5e24e4197ad4bbb0d96cd9389ed98@ston106.cypress.klaytn.net:32323?ntype=bn",
		},
	},
}

// BaobabBootnodes are the URLs of bootnodes running on the Baobab test network.
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
