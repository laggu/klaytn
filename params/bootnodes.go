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

type bootnodesByTypes struct {
	Addrs []string
}

// MainnetBootnodes are the kni URLs of the P2P bootstrap nodes running on
// the main klaytn network.
var MainnetBootnodes = []string{
	// TODO-Klaytn-Bootnode : Klaytn BootNode should be set. Now for only test.
	//"kni://a979fb575495b8d6db44f750317d0f4622bf4c2aa3365d6af7c284339968eef29b69ad0dce72a4d8db5ebb4968de0e3bec910127f134779fbcb0cb6d3331163c@52.16.188.185:30303", // IE
	//"kni://3f1d12044546b76342d59d4a05532c14b85aa669704bfe1f864fe079415aa2c02d743e03218e57a33fb94523adb54032871a6c51b2cc5514cb7c7e35b3ed0a99@13.93.211.84:30303",  // US-WEST
}

// TODO-Klaytn-Bootnode: below consts are derived from `node` package due to importing `node` package occurs cyclic import issue
const (
	CONSENSUSNODE = iota
	ENDPOINTNODE
	PROXYNODE
)

// BaobabBootnodes are the kni URLs of the PN's P2P bootstrap nodes running on the
// Baobab test network.
var BaobabBootnodes = map[int]bootnodesByTypes{
	// TODO-Klaytn-Bootnode: realize bootnode URLs and domains
	CONSENSUSNODE: {
		// CN bootnodes
		[]string{
			"kni://5549df14326d6af1272c4a4375c8b7aec3f6eed3a359e390aeed882bddb215837bb73490230f35973c39e5298b1c130c2a557f4b88e3462a89bab39ca8de3adf@permissoned.baobab.kr.klaytn.net:32323?discport=32323", // Imaginary (KR) bootnode for CN
			"kni://572eac675ad859034958570313f48e2de532a9d83717fbc257bdecd1e01250369fab5adbd9d14bc513b1844e5048df163efac161d878eb61cb033b830b017054@permissoned.baobab.jp.klaytn.net:32323?discport=32323", // Imaginary (JP) bootnode for CN
		},
	},
	ENDPOINTNODE: {
		// EN bootnodes
		[]string{
			"kni://559df4b2fa31c1650d7b0343156c726026613f79966e6af5c9c3c0fd65fc652827f6bf03a690de99a2d82b1acce64fbcc7b1471ee3d994d987a538c1907dec96@13.209.140.55:32323?discport=32323",
			"kni://5bcd28169dc7444e22c6aa7cf428086e77689aea5fc820e3c973167d4131d6a9efc6a38a93e75a887ac99734c440d3c3c90215ce1b349bd64ed11ee18e4dc265@52.78.141.146:32323?discport=32323",
		},
	},
	PROXYNODE: {
		// PN (formerly known as BN) bootnodes
		[]string{
			"kni://11eb3d77843914f4b78c8b814b343e0825fe1adc0ec2df001bec9cce6ff0fd8ae5c36dec31ce71f00b00de0c7230d22c54507520fd449986f6ad062510d5c9d9@bridge.baobab.kr.klaytn.net:32323?discport=32323", // Imaginary (KR) bootnode for PN
			"kni://b171d2ccc5ee35451d766401ba38392ed12d71f87459b371955a94caadc1c2228859169c44d817b52d03b12f3d91b9ebb204592ce8b290ee5f0824eafabf292b@bridge.baobab.jp.klaytn.net:32323?discport=32323", // Imaginary (JP) bootnode for PN
		},
	},
}
