package ranger

import (
	"time"
	"math/big"
	"os"
	"os/user"
	"runtime"
	"path/filepath"
	"github.com/ground-x/go-gxplatform/datasync/downloader"
	"github.com/ground-x/go-gxplatform/consensus/gxhash"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/common/hexutil"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
)

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
	NetworkId:     1,
	DatabaseCache: 768,
	TrieCache:     256,
	TrieTimeout:   5 * time.Minute,
	GasPrice:      big.NewInt(18 * params.Ston), // TODO-GX-issue136 default gasPrice
	Istanbul: *istanbul.DefaultConfig,
	ConsensusURL:  "ws://localhost:8546",
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

	// Database options
	SkipBcVersionCheck bool `toml:"-"`
	DatabaseHandles    int  `toml:"-"`
	DatabaseCache      int
	TrieCache          int
	TrieTimeout        time.Duration

	// Mining-related options
	Gxbase       common.Address `toml:",omitempty"`
	MinerThreads int            `toml:",omitempty"`
	ExtraData    []byte         `toml:",omitempty"`
	GasPrice     *big.Int

	// Gxhash options
	Gxhash gxhash.Config

	// Miscellaneous options
	DocRoot string `toml:"-"`

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	ConsensusURL string

	// Istanbul options
	Istanbul istanbul.Config
}

type configMarshaling struct {
	ExtraData hexutil.Bytes
}
