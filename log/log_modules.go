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

package log

import (
	"strconv"
	"strings"
)

type ModuleID int

// When printID is true, log prints ModuleID instead of ModuleName.
// TODO-Klaytn This can be controlled by runtime configuration.
var printID = true

func GetModuleName(mi ModuleID) string {
	return moduleNames[int(mi)]
}

func GetModuleID(moduleName string) ModuleID {
	moduleName = strings.ToLower(moduleName)
	for i := 0; i < len(moduleNames); i++ {
		if moduleName == moduleNames[i] {
			return ModuleID(i)
		}
	}
	return ModuleNameLen
}

func (mi ModuleID) String() string {
	if printID {
		return strconv.Itoa(int(mi))
	}
	return moduleNames[mi]
}

const (
	// 0
	BaseLogger ModuleID = iota

	// 1~10
	AccountsAbiBind
	AccountsKeystore
	API
	APIDebug
	Blockchain
	BlockchainState
	BlockchainTypes
	BlockchainTypesAccount
	BlockchainTypesAccountKey
	CMDKlay

	// 11~20

	CMDUtils
	Common
	ConsensusGxhash
	ConsensusIstanbul
	ConsensusIstanbulBackend
	ConsensusIstanbulCore
	ConsensusIstanbulValidator
	Console
	DatasyncDownloader

	// 21~30
	DatasyncFetcher
	Metrics
	NetworksP2P
	NetworksP2PDiscover
	NetworksP2PNat
	NetworksP2PSimulations
	NetworksP2PSimulationsAdapters
	NetworksP2PSimulationsCnism
	NetworksRPC
	Node

	// 31~40
	NodeCN
	NodeCNFilters
	NodeCNTracers

	Reward
	ServiceChain
	StorageDatabase
	StorageStateDB
	VM
	Work

	// 41~50
	CMDKBN
	CMDUtilsNodeCMD
	CMDKCN
	CMDKPN
	CMDKEN
	NetworksGRPC
	CMDKSCN

	POA

	// ModuleNameLen should be placed at the end of the list.
	ModuleNameLen
)

var moduleNames = [ModuleNameLen]string{
	// 0
	"defaultLogger",

	// 1~10
	"accounts/abi/bind",
	"accounts/keystore",
	"api",
	"api/debug",
	"blockchain",
	"blockchain/state",
	"blockchain/types",
	"blockchain/types/account",
	"blockchain/types/accountkey",
	"cmd/klay",

	// 11~20

	"cmd/utils",
	"common",
	"consensus/gxhash",
	"consensus/istanbul",
	"consensus/istanbul/backend",
	"consensus/istanbul/core",
	"consensus/istanbul/validator",
	"console",
	"datasync/downloader",

	// 21~30
	"datasync/fetcher",
	"metrics",
	"networks/p2p",
	"networks/p2p/discover",
	"networks/p2p/nat",
	"networks/p2p/simulations",
	"networks/p2p/simulations/adapters",
	"networks/p2p/simulations/cnism",
	"networks/rpc",
	"node",

	// 31~40
	"node/cn",
	"node/cn/filters",
	"node/cn/tracers",

	"contracts/reward",
	"servicechain",
	"storage/database",
	"storage/statedb",
	"vm",
	"work",

	// 41~50
	"cmd/kbn",
	"cmd/utils/nodecmd",
	"cmd/kcn",
	"cmd/kpn",
	"cmd/ken",
	"networks/grpc",
	"cmd/kscn",

	"consensus/clique",
}
