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

package main

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ground-x/klaytn/cmd/utils"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"github.com/ground-x/klaytn/networks/p2p/nat"
	"github.com/ground-x/klaytn/networks/p2p/netutil"
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

func (ctx *bootnodeConfig) checkCMDState() int {
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

func (ctx *bootnodeConfig) generateNodeKey() {
	nodeKey, err := crypto.GenerateKey()
	if err != nil {
		utils.Fatalf("could not generate key: %v", err)
	}
	if err = crypto.SaveECDSA(ctx.genKeyPath, nodeKey); err != nil {
		utils.Fatalf("%v", err)
	}
	os.Exit(0)
}

func (ctx *bootnodeConfig) doWriteOutAddress() {
	err := ctx.readNodeKey()
	if err != nil {
		utils.Fatalf("Failed to read node key: %v", err)
	}
	fmt.Printf("%v\n", discover.PubkeyID(&(ctx.nodeKey).PublicKey))
	os.Exit(0)
}

func (ctx *bootnodeConfig) readNodeKey() error {
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

func (ctx *bootnodeConfig) validateNetworkParameter() error {
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
