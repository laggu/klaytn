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
// This file is derived from cmd/bootnode/main.go (2018/06/04).
// Modified and improved for the klaytn development.

package main

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/api/debug"
	"github.com/ground-x/klaytn/cmd/utils"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"github.com/ground-x/klaytn/networks/p2p/nat"
	"github.com/ground-x/klaytn/networks/p2p/netutil"
	"gopkg.in/urfave/cli.v1"
	"net"
	"os"
)

type bootnodeConfig struct {
	// Parameter variables
	addr         string
	genKeyPath   string
	nodeKeyFile  string
	nodeKeyHex   string
	natFlag      string
	netrestrict  string
	writeAddress bool

	// Context
	restrictList *netutil.Netlist
	nodeKey      *ecdsa.PrivateKey
	natm         nat.Interface
	listenAddr   string
}

var (
	logger = log.NewModuleLogger(log.CMDKBN)
)

const (
	generateNodeKeySpecified = iota
	noPrivateKeyPathSpecified
	nodeKeyDuplicated
	writeOutAddress
	goodToGo
)

func checkCMDState(ctx bootnodeConfig) int {
	if ctx.genKeyPath != "" {
		return generateNodeKeySpecified
	}
	if ctx.nodeKeyFile == "" && ctx.nodeKeyHex == "" {
		return noPrivateKeyPathSpecified
	}
	if ctx.nodeKeyFile != "" && ctx.nodeKeyHex != "" {
		return nodeKeyDuplicated
	}
	if ctx.writeAddress {
		return writeOutAddress
	}
	return goodToGo
}

func generateNodeKey(path string) {
	nodeKey, err := crypto.GenerateKey()
	if err != nil {
		utils.Fatalf("could not generate key: %v", err)
	}
	if err = crypto.SaveECDSA(path, nodeKey); err != nil {
		utils.Fatalf("%v", err)
	}
	os.Exit(0)
}

func doWriteOutAddress(ctx *bootnodeConfig) {
	err := readNodeKey(ctx)
	if err != nil {
		utils.Fatalf("Failed to read node key: %v", err)
	}
	fmt.Printf("%v\n", discover.PubkeyID(&(ctx.nodeKey).PublicKey))
	os.Exit(0)
}

func readNodeKey(ctx *bootnodeConfig) error {
	var err error
	if ctx.nodeKeyFile != "" {
		ctx.nodeKey, err = crypto.LoadECDSA(ctx.nodeKeyFile)
		return err
	}
	if ctx.nodeKeyHex != "" {
		ctx.nodeKey, err = crypto.LoadECDSA(ctx.nodeKeyHex)
		return err
	}
	return nil
}

func validateNetworkParameter(ctx *bootnodeConfig) error {
	var err error
	if ctx.natFlag != "" {
		ctx.natm, err = nat.Parse(ctx.natFlag)
		if err != nil {
			return err
		}
	}

	if ctx.netrestrict != "" {
		ctx.restrictList, err = netutil.ParseNetlist(ctx.netrestrict)
		if err != nil {
			return err
		}
	}

	if ctx.addr[0] != ':' {
		ctx.listenAddr = ":" + ctx.addr
	} else {
		ctx.listenAddr = ctx.addr
	}

	return nil
}

func bootnode(c *cli.Context) error {
	var (
		// Local variables
		err error
		ctx = bootnodeConfig{
			// Config variables
			addr:         c.GlobalString(utils.AddrFlag.Name),
			genKeyPath:   c.GlobalString(utils.GenKeyFlag.Name),
			nodeKeyFile:  c.GlobalString(utils.NodeKeyFileFlag.Name),
			nodeKeyHex:   c.GlobalString(utils.NodeKeyHexFlag.Name),
			natFlag:      c.GlobalString(utils.NATFlag.Name),
			netrestrict:  c.GlobalString(utils.NetrestrictFlag.Name),
			writeAddress: c.GlobalBool(utils.WriteAddressFlag.Name),
		}
	)

	// Check exit condition
	switch checkCMDState(ctx) {
	case generateNodeKeySpecified:
		generateNodeKey(ctx.genKeyPath)
	case noPrivateKeyPathSpecified:
		return errors.New("Use --nodekey or --nodekeyhex to specify a private key")
	case nodeKeyDuplicated:
		return errors.New("Options --nodekey and --nodekeyhex are mutually exclusive")
	case writeOutAddress:
		doWriteOutAddress(&ctx)
	default:
		err = readNodeKey(&ctx)
		if err != nil {
			return err
		}
	}

	err = validateNetworkParameter(&ctx)
	if err != nil {
		return err
	}

	addr, err := net.ResolveUDPAddr("udp", ctx.listenAddr)
	if err != nil {
		utils.Fatalf("Failed to ResolveUDPAddr: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		utils.Fatalf("Failed to ListenUDP: %v", err)
	}

	realaddr := conn.LocalAddr().(*net.UDPAddr)
	if ctx.natm != nil {
		if !realaddr.IP.IsLoopback() {
			go nat.Map(ctx.natm, nil, "udp", realaddr.Port, realaddr.Port, "Klaytn node discovery")
		}
		// TODO: react to external IP changes over time.
		if ext, err := ctx.natm.ExternalIP(); err == nil {
			realaddr = &net.UDPAddr{IP: ext, Port: realaddr.Port}
		}
	}

	cfg := discover.Config{
		PrivateKey:   ctx.nodeKey,
		AnnounceAddr: realaddr,
		NetRestrict:  ctx.restrictList,
	}
	if _, err := discover.ListenUDP(conn, cfg); err != nil {
		utils.Fatalf("%v", err)
	}

	select {}
}

func main() {
	var (
		cliFlags = []cli.Flag{
			utils.GenKeyFlag,
			utils.NodeKeyFileFlag,
			utils.NodeKeyHexFlag,
			utils.WriteAddressFlag,
			utils.AddrFlag,
			utils.NATFlag,
			utils.NetrestrictFlag,
		}
	)
	// TODO-Klaytn: remove `help` command
	app := utils.NewApp("", "the klaytn's bootnode command line interface")
	app.Name = "kbn"
	app.Copyright = "Copyright 2018 The klaytn Authors"
	app.UsageText = app.Name + " [global options] [commands]"
	app.Flags = append(app.Flags, cliFlags...)
	app.Flags = append(app.Flags, debug.Flags...)

	app.Action = bootnode
	app.Before = func(c *cli.Context) error {
		if err := debug.Setup(c); err != nil {
			return err
		}
		return nil
	}

	app.After = func(c *cli.Context) error {
		debug.Exit()
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
