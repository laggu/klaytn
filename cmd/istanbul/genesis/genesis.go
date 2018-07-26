package genesis

import (
	"math/big"
	"time"
	"path/filepath"
	"encoding/json"
	"github.com/ground-x/go-gxplatform/core"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"github.com/ground-x/go-gxplatform/core/types"
	"io/ioutil"
	"github.com/ground-x/go-gxplatform/cmd/istanbul/common"
	"github.com/ground-x/go-gxplatform/log"
)

const (
	FileName       = "genesis.json"
	InitGasLimit   = 9000000000000
	InitDifficulty = 1
)

func New(options ...Option) *core.Genesis {
	genesis := &core.Genesis{
		Timestamp:  uint64(time.Now().Unix()),
		GasLimit:   InitGasLimit,
		Difficulty: big.NewInt(InitDifficulty),
		Alloc:      make(core.GenesisAlloc),
		Config: &params.ChainConfig{
			ChainId:        big.NewInt(2017),
			HomesteadBlock: big.NewInt(1),
			EIP155Block:    big.NewInt(3),
			Istanbul: &params.IstanbulConfig{
				ProposerPolicy: uint64(istanbul.DefaultConfig.ProposerPolicy),
				Epoch:          istanbul.DefaultConfig.Epoch,
				SubGroupSize:   istanbul.DefaultConfig.SubGroupSize,

			},
		},
		Mixhash: types.IstanbulDigest,
	}

	for _, opt := range options {
		opt(genesis)
	}

	return genesis
}

func NewFileAt(dir string, isQuorum bool, options ...Option) string {
	genesis := New(options...)
	if err := Save(dir, genesis, isQuorum); err != nil {
		log.Error("Failed to save genesis", "dir", dir, "err", err)
		return ""
	}

	return filepath.Join(dir, FileName)
}

func NewFile(isQuorum bool, options ...Option) string {
	dir, _ := common.GenerateRandomDir()
	return NewFileAt(dir, isQuorum, options...)
}

func Save(dataDir string, genesis *core.Genesis, isQuorum bool) error {
	filePath := filepath.Join(dataDir, FileName)

	var raw []byte
	var err error
	if isQuorum {
		raw, err = json.Marshal(ToBFT(genesis, true))
	} else {
		raw, err = json.Marshal(genesis)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, raw, 0600)
}

