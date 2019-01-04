// Copyright 2018 The go-klaytn Authors
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ground-x/go-gxplatform/cmd/utils"
	"github.com/ground-x/go-gxplatform/node"
	rnpkg "github.com/ground-x/go-gxplatform/node/ranger"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/naoina/toml"
	"gopkg.in/urfave/cli.v1"
	"io"
	"os"
	"reflect"
	"unicode"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(append(nodeFlags, rpcFlags...)),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}

	configFileFlag = cli.StringFlag{
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

type rangerConfig struct {
	Gxp  rnpkg.Config
	Node node.Config
}

func loadConfig(file string, cfg *rangerConfig) error {
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

func defaultRangerConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit)
	cfg.HTTPModules = append(cfg.HTTPModules, "klay")
	cfg.WSModules = append(cfg.WSModules, "klay")
	cfg.IPCPath = "klay.ipc"
	return cfg
}

func makeConfigRanger(ctx *cli.Context) (*node.Node, rangerConfig) {
	// Load defaults.
	cfg := rangerConfig{
		Gxp:  rnpkg.DefaultConfig,
		Node: defaultRangerConfig(),
	}

	// Load config file.
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
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
	utils.SetRnConfig(ctx, stack, &cfg.Gxp)
	//if ctx.GlobalIsSet(utils.EthStatsURLFlag.Name) {
	//	cfg.Ethstats.URL = ctx.GlobalString(utils.EthStatsURLFlag.Name)
	//}
	//
	//utils.SetShhConfig(ctx, stack, &cfg.Shh)
	//utils.SetDashboardConfig(ctx, &cfg.Dashboard)

	return stack, cfg
}

func makeRanger(ctx *cli.Context) *node.Node {
	stack, cfg := makeConfigRanger(ctx)

	utils.RegisterRanagerService(stack, &cfg.Gxp)

	return stack
}

func dumpConfig(ctx *cli.Context) error {
	_, cfg := makeConfigRanger(ctx)
	comment := ""

	if cfg.Gxp.Genesis != nil {
		cfg.Gxp.Genesis = nil
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
