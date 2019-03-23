// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from cmd/geth/chaincmd.go (2018/06/04).
// Modified and improved for the klaytn development.

package nodecmd

import (
	"encoding/json"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/cmd/utils"
	"github.com/ground-x/klaytn/governance"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/storage/database"
	"gopkg.in/urfave/cli.v1"
	"os"
)

var logger = log.NewModuleLogger(log.CMDUtilsNodeCMD)

var (
	InitCommand = cli.Command{
		Action:    utils.MigrateFlags(initGenesis),
		Name:      "init",
		Usage:     "Bootstrap and initialize a new genesis block",
		ArgsUsage: "<genesisPath>",
		Flags: []cli.Flag{
			utils.DbTypeFlag,
			utils.NoPartitionedDBFlag,
			utils.DataDirFlag,
		},
		Category: "BLOCKCHAIN COMMANDS",
		Description: `
The init command initializes a new genesis block and definition for the network.
This is a destructive action and changes the network in which you will be
participating.

It expects the genesis file as argument.`,
	}
)

// initGenesis will initialise the given JSON format genesis file and writes it as
// the zero'd block (i.e. genesis) or will fail hard if it can't succeed.
func initGenesis(ctx *cli.Context) error {
	// Make sure we have a valid genesis JSON
	genesisPath := ctx.Args().First()
	if len(genesisPath) == 0 {
		utils.Fatalf("Must supply path to genesis JSON file")
	}
	file, err := os.Open(genesisPath)
	if err != nil {
		utils.Fatalf("Failed to read genesis file: %v", err)
	}
	defer file.Close()

	genesis := new(blockchain.Genesis)
	if err := json.NewDecoder(file).Decode(genesis); err != nil {
		utils.Fatalf("invalid genesis file: %v", err)
		return err
	}

	genesis = checkGenesisAndFillDefaultIfNeeded(genesis)
	if err := governance.CheckGenesisValues(genesis.Config); err != nil {
		logger.Crit("Error in genesis json values", "err", err)
	}
	genesis.Governance, err = governance.MakeGovernanceData(genesis.Config.Governance)

	if genesis.Config.StakingUpdateInterval != 0 {
		params.SetStakingUpdateInterval(genesis.Config.StakingUpdateInterval)
	} else {
		genesis.Config.StakingUpdateInterval = params.StakingUpdateInterval()
	}

	if genesis.Config.ProposerUpdateInterval != 0 {
		params.SetProposerUpdateInterval(genesis.Config.ProposerUpdateInterval)
	} else {
		genesis.Config.ProposerUpdateInterval = params.ProposerUpdateInterval()
	}

	if err != nil {
		logger.Crit("Error in making governance data", "err", err)
	}

	// Open an initialise both full and light databases
	stack := MakeFullNode(ctx)

	parallelDBWrite := utils.IsParallelDBWrite(ctx)
	partitioned := utils.IsPartitionedDB(ctx)
	for _, name := range []string{"chaindata", "lightchaindata"} {
		dbc := &database.DBConfig{Dir: name, DBType: database.LevelDB, ParallelDBWrite: parallelDBWrite, Partitioned: partitioned,
			LevelDBCacheSize: 0, LevelDBHandles: 0, ChildChainIndexing: false}
		chaindb := stack.OpenDatabase(dbc)
		// Initialize DeriveSha implementation
		blockchain.InitDeriveSha(genesis.Config.DeriveShaImpl)

		_, hash, err := blockchain.SetupGenesisBlock(chaindb, genesis)
		if err != nil {
			utils.Fatalf("Failed to write genesis block: %v", err)
		}
		logger.Info("Successfully wrote genesis state", "database", name, "hash", hash)
		chaindb.Close()
	}
	return nil
}

func checkGenesisAndFillDefaultIfNeeded(genesis *blockchain.Genesis) *blockchain.Genesis {
	engine := params.UseIstanbul
	valueChanged := false

	// using Clique as a consensus engine
	if genesis.Config.Istanbul == nil && genesis.Config.Clique != nil {
		engine = params.UseClique
		if genesis.Config.Governance == nil {
			genesis.Config.Governance = governance.GetDefaultGovernanceConfig(engine)
		}
		valueChanged = true
	} else if genesis.Config.Istanbul == nil && genesis.Config.Clique == nil {
		engine = params.UseIstanbul
		genesis.Config.Istanbul = governance.GetDefaultIstanbulConfig()
		valueChanged = true
	} else if genesis.Config.Istanbul != nil && genesis.Config.Clique != nil {
		// Error case. Both istanbul and Clique exists
		logger.Crit("Both clique and istanbul configuration exists. Only one configuration can be applied. Exiting..")
	}

	// We have governance config
	if genesis.Config.Governance != nil {
		// and also we have istanbul config. Then use governance's value prior to istanbul's value
		if engine == params.UseIstanbul && genesis.Config.Governance.Istanbul != nil {
			if genesis.Config.Istanbul != nil {
				if genesis.Config.Istanbul.Epoch != genesis.Config.Governance.Istanbul.Epoch ||
					genesis.Config.Istanbul.ProposerPolicy != genesis.Config.Governance.Istanbul.ProposerPolicy ||
					genesis.Config.Istanbul.SubGroupSize != genesis.Config.Governance.Istanbul.SubGroupSize {
					valueChanged = true
				}
			} else {
				genesis.Config.Istanbul = new(params.IstanbulConfig)
			}

			genesis.Config.UnitPrice = genesis.Config.Governance.UnitPrice
			genesis.Config.Istanbul.Epoch = genesis.Config.Governance.Istanbul.Epoch
			genesis.Config.Istanbul.SubGroupSize = genesis.Config.Governance.Istanbul.SubGroupSize
			genesis.Config.Istanbul.ProposerPolicy = genesis.Config.Governance.Istanbul.ProposerPolicy
		}
	} else {
		genesis.Config.Governance = governance.GetDefaultGovernanceConfig(engine)

		// We don't have governance config and engine is istanbul
		if engine == params.UseIstanbul && genesis.Config.Istanbul != nil {
			genesis.Config.Governance.UnitPrice = genesis.Config.UnitPrice
			genesis.Config.Governance.Istanbul.Epoch = genesis.Config.Istanbul.Epoch
			genesis.Config.Governance.Istanbul.SubGroupSize = genesis.Config.Istanbul.SubGroupSize
			genesis.Config.Governance.Istanbul.ProposerPolicy = genesis.Config.Istanbul.ProposerPolicy
		}
	}

	if valueChanged {
		logger.Warn("Some input value of genesis.json have been set to default or changed")
	}
	return genesis
}
