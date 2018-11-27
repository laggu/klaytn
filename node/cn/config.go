package cn

import (
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/common/hexutil"
	"github.com/ground-x/go-gxplatform/consensus/gxhash"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/datasync/downloader"
	"github.com/ground-x/go-gxplatform/node/cn/gasprice"
	"github.com/ground-x/go-gxplatform/params"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"github.com/ground-x/go-gxplatform/log"
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
	NetworkId:        1,
	LightPeers:       100,
	LevelDBCacheSize: 768,
	TrieCacheSize:    256,
	TrieTimeout:      5 * time.Minute,
	TrieBlockInterval: blockchain.DefaultBlockInterval,
	GasPrice:         big.NewInt(18 * params.Ston), // TODO-GX-issue136 default gasPrice

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

	// Light client options
	LightServ  int `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	LightPeers int `toml:",omitempty"` // Maximum number of LES client peers

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	LevelDBCacheSize   int
	TrieCacheSize      int
	TrieTimeout        time.Duration
	TrieBlockInterval  uint

	// Mining-related options
	Gxbase       common.Address `toml:",omitempty"`
	MinerThreads int            `toml:",omitempty"`
	ExtraData    []byte         `toml:",omitempty"`
	GasPrice     *big.Int

	// Reward
	RewardContract common.Address `toml:",omitempty"`
	Rewardbase   common.Address `toml:",omitempty"`

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
