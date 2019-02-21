// Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
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
// This file is derived from cmd/geth/config.go (2018/06/04).
// Modified and improved for the klaytn development.

package nodecmd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/cmd/utils"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/node"
	"github.com/ground-x/klaytn/node/cn"
	"github.com/ground-x/klaytn/node/sc"
	"github.com/ground-x/klaytn/params"
	"gopkg.in/urfave/cli.v1"
	"os"
	"reflect"
	"unicode"

	"github.com/naoina/toml"
	"io"
)

var (
	ConfigFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

type klayConfig struct {
	CN   cn.Config
	Node node.Config
}

// GetDumpConfigCommand returns cli.Command `dumpconfig` whose flags are initialized with nodeFlags and rpcFlags.
func GetDumpConfigCommand(nodeFlags, rpcFlags []cli.Flag) cli.Command {
	return cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(nodeFlags, rpcFlags...)),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}
}

func loadConfig(file string, cfg *klayConfig) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "klay", "shh")
	cfg.WSModules = append(cfg.WSModules, "klay", "shh")
	cfg.IPCPath = "klay.ipc"
	return cfg
}

func makeConfigNode(ctx *cli.Context) (*node.Node, klayConfig) {
	// TODO-Klaytn-Issue136 gasPrice
	// Load defaults.
	cfg := klayConfig{
		CN:   cn.DefaultConfig,
		Node: defaultNodeConfig(),
	}

	// Load config file.
	if file := ctx.GlobalString(ConfigFileFlag.Name); file != "" {
		if err := loadConfig(file, &cfg); err != nil {
			utils.Fatalf("%v", err)
		}
	}

	// Apply flags.
	utils.SetNodeConfig(ctx, &cfg.Node)
	stack, err := node.New(&cfg.Node)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetKlayConfig(ctx, stack, &cfg.CN)

	//utils.SetShhConfig(ctx, stack, &cfg.Shh)
	//utils.SetDashboardConfig(ctx, &cfg.Dashboard)

	return stack, cfg
}

func makeServiceChainConfig(ctx *cli.Context) (config sc.SCConfig) {
	cfg := sc.SCConfig{
		// TODO-Klaytn this value is temp for test
		NetworkId: 1,
		MaxPeer:   50,
	}

	// bridge service
	if ctx.GlobalBool(utils.EnabledBridgeFlag.Name) {
		cfg.EnabledBridge = true
		if ctx.GlobalBool(utils.IsMainBridgeFlag.Name) {
			cfg.IsMainBridge = true
		} else {
			cfg.IsMainBridge = false
		}
		if ctx.GlobalIsSet(utils.BridgeListenPortFlag.Name) {
			cfg.BridgePort = fmt.Sprintf(":%d", ctx.GlobalInt(utils.BridgeListenPortFlag.Name))
		}
		//// TODO-Klaytn-ServiceChain Add human-readable address once its implementation is introduced.
		if ctx.GlobalIsSet(utils.ChainAccountAddrFlag.Name) {
			tempStr := ctx.GlobalString(utils.ChainAccountAddrFlag.Name)
			if !common.IsHexAddress(tempStr) {
				logger.Crit("Given chainaddr does not meet hex format.", "chainaddr", tempStr)
			}
			tempAddr := common.StringToAddress(tempStr)
			cfg.ChainAccountAddr = &tempAddr
			logger.Info("A chain address is registered.", "chainAccountAddr", *cfg.ChainAccountAddr)
		}
		cfg.AnchoringPeriod = ctx.GlobalUint64(utils.AnchoringPeriodFlag.Name)
		cfg.SentChainTxsLimit = ctx.GlobalUint64(utils.SentChainTxsLimit.Name)
	} else {
		cfg.EnabledBridge = false
	}

	return cfg
}

func MakeFullNode(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigNode(ctx)

	utils.RegisterCNService(stack, &cfg.CN)

	scfg := makeServiceChainConfig(ctx)
	scfg.DataDir = cfg.Node.DataDir

	utils.RegisterService(stack, &scfg)

	return stack
}

func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigNode(ctx)
	comment := ""

	if cfg.CN.Genesis != nil {
		cfg.CN.Genesis = nil
		comment += "# Note: this config doesn't contain the genesis block.\n\n"
	}

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}
	io.WriteString(os.Stdout, comment)
	os.Stdout.Write(out)
	return nil
}
