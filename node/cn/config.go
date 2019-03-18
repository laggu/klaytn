// Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
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
// This file is derived from eth/config.go (2018/06/04).
// Modified and improved for the klaytn development.

package cn

import (
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/consensus/gxhash"
	"github.com/ground-x/klaytn/consensus/istanbul"
	"github.com/ground-x/klaytn/datasync/downloader"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/node/cn/gasprice"
	"github.com/ground-x/klaytn/params"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
)

var logger = log.NewModuleLogger(log.NodeCN)

// DefaultConfig contains default settings for use on the klaytn main net.
var DefaultConfig = Config{
	SyncMode: downloader.FullSync,
	Gxhash: gxhash.Config{
		CacheDir:       "gxhash",
		CachesInMem:    2,
		CachesOnDisk:   3,
		DatasetsInMem:  1,
		DatasetsOnDisk: 2,
	},
	NetworkId:         1,
	LevelDBCacheSize:  768,
	TrieCacheSize:     256,
	TrieTimeout:       5 * time.Minute,
	TrieBlockInterval: blockchain.DefaultBlockInterval,
	GasPrice:          big.NewInt(18 * params.Ston), // TODO-Klaytn-Issue136 default gasPrice

	TxPool: blockchain.DefaultTxPoolConfig,
	GPO: gasprice.Config{
		Blocks:     20,
		Percentile: 60,
	},
	WsEndpoint: "localhost:8546",

	Istanbul: *istanbul.DefaultConfig,
}

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		if user, err := user.Current(); err == nil {
			home = user.HomeDir
		}
	}
	if runtime.GOOS == "windows" {
		DefaultConfig.Gxhash.DatasetDir = filepath.Join(home, "AppData", "Gxhash")
	} else {
		DefaultConfig.Gxhash.DatasetDir = filepath.Join(home, ".gxhash")
	}
}

//go:generate gencodec -type Config -field-override configMarshaling -formats toml -out gen_config.go

type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the klaytn main net block is used.
	Genesis *blockchain.Genesis `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	SyncMode  downloader.SyncMode
	NoPruning bool

	// Service chain options
	ChainAccountAddr  *common.Address // A hex account address in parent chain used to sign service chain transaction.
	AnchoringPeriod   uint64          // Period when child chain sends an anchoring transaction to parent chain. Default value is 1.
	SentChainTxsLimit uint64          // Number of chain transactions stored for resending. Default value is 1000.

	// Light client options
	//LightServ  int `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	//LightPeers int `toml:",omitempty"` // Maximum number of LES client peers

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	PartitionedDB      bool
	LevelDBCacheSize   int
	TrieCacheSize      int
	TrieTimeout        time.Duration
	TrieBlockInterval  uint
	ChildChainIndexing bool
	ParallelDBWrite    bool
	StateDBCaching     bool

	// Mining-related options
	Gxbase    common.Address `toml:",omitempty"`
	ExtraData []byte         `toml:",omitempty"`
	GasPrice  *big.Int

	// Reward
	Rewardbase common.Address `toml:",omitempty"`

	// Gxhash options
	Gxhash gxhash.Config

	// Transaction pool options
	TxPool blockchain.TxPoolConfig

	// Gas Price Oracle options
	GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool
	// Istanbul options
	Istanbul istanbul.Config

	// Miscellaneous options
	DocRoot string `toml:"-"`

	WsEndpoint string `toml:",omitempty"`
}

type configMarshaling struct {
	ExtraData hexutil.Bytes
}
