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
	"github.com/ground-x/klaytn/log"
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
			utils.PartitionedDBFlag,
			utils.DataDirFlag,
			utils.LightModeFlag,
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
	}
	// Open an initialise both full and light databases
	stack := MakeFullNode(ctx)

	parallelDBWrite := utils.IsParallelDBWrite(ctx)
	partitioned := utils.IsPartitionedDB(ctx)
	for _, name := range []string{"chaindata", "lightchaindata"} {
		dbc := &database.DBConfig{Dir: name, DBType: database.LevelDB, ParallelDBWrite: parallelDBWrite, Partitioned: partitioned,
			LevelDBCacheSize: 0, LevelDBHandles: 0, ChildChainIndexing: false}
		chaindb, err := stack.OpenDatabase(dbc)
		if err != nil {
			utils.Fatalf("Failed to open database: %v", err)
		}
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
