// Copyright 2018 The klaytn Authors
// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
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

package setup

import "gopkg.in/urfave/cli.v1"

var fundingAddr string
var dockerImageId string
var outputPath string

var (
	cliqueFlag = cli.BoolFlag{
		Name:  "clique",
		Usage: "Use Clique consensus",
	}

	numOfCNsFlag = cli.IntFlag{
		Name:  "num",
		Usage: "Number of consensus nodes",
		Value: 4,
	}

	unitPriceFlag = cli.Uint64Flag{
		Name:  "unitPrice",
		Usage: "Price of unit",
		Value: 0,
	}

	deriveShaImplFlag = cli.IntFlag{
		Name:  "deriveShaImpl",
		Usage: "Implementation of DeriveSha [0:Original, 1:Simple, 2:Concat]",
		Value: 0,
	}

	outputPathFlag = cli.StringFlag{
		Name:        "output, o",
		Usage:       "istanbul's result saved at this output folder",
		Value:       "istanbul-output",
		Destination: &outputPath,
	}

	fundingAddrFlag = cli.StringFlag{
		Name:        "fundingAddr",
		Value:       "75a59b94889a05c03c66c3c84e9d2f8308ca4abd",
		Usage:       "Give initial fund to the given addr",
		Destination: &fundingAddr,
	}

	dockerImageIdFlag = cli.StringFlag{
		Name:        "docker-image-id",
		Value:       "428948643293.dkr.ecr.ap-northeast-2.amazonaws.com/gxp/client-go:latest",
		Usage:       "Base docker image ID (Image[:tag])",
		Destination: &dockerImageId,
	}

	subGroupSizeFlag = cli.IntFlag{
		Name:  "subgroup-size",
		Usage: "CN's Subgroup size",
		Value: 21,
	}

	fasthttpFlag = cli.BoolFlag{
		Name:  "fasthttp",
		Usage: "(docker only) Use High performance http module",
	}

	networkIdFlag = cli.IntFlag{
		Name:  "network-id",
		Usage: "(docker only) network identifier (default : 2018)",
		Value: 2018,
	}

	nografanaFlag = cli.BoolFlag{
		Name:  "no-grafana",
		Usage: "(docker only) Do not make grafana container",
	}

	numOfPNsFlag = cli.IntFlag{
		Name:  "pn-num",
		Usage: "(docker, deploy only) Number of proxy node",
		Value: 1,
	}

	useTxGenFlag = cli.BoolFlag{
		Name:  "txgen",
		Usage: "(docker only) Add txgen container",
	}

	txGenRateFlag = cli.IntFlag{
		Name:  "txgen-rate",
		Usage: "(docker only) txgen's rate option [default : 2000]",
		Value: 2000,
	}

	txGenConnFlag = cli.IntFlag{
		Name:  "txgen-conn",
		Usage: "(docker only) txgen's connection size option [default : 100]",
		Value: 100,
	}

	txGenDurFlag = cli.StringFlag{
		Name:  "txgen-dur",
		Usage: "(docker only) txgen's duration option [default : 1m]",
		Value: "1m",
	}

	txGenThFlag = cli.IntFlag{
		Name:  "txgen-th",
		Usage: "(docker-only) txgen's thread size option [default : 2]",
		Value: 2,
	}

	rpcPortFlag = cli.IntFlag{
		Name:  "rpc-port",
		Usage: "klay.conf - klaytn node's rpc port [default: 8551] ",
		Value: 8551,
	}

	wsPortFlag = cli.IntFlag{
		Name:  "ws-port",
		Usage: "klay.conf - klaytn node's ws port [default: 8552]",
		Value: 8552,
	}

	p2pPortFlag = cli.IntFlag{
		Name:  "p2p-port",
		Usage: "klay.conf - klaytn node's p2p port [default: 32323]",
		Value: 32323,
	}

	dataDirFlag = cli.StringFlag{
		Name:  "data-dir",
		Usage: "klay.conf - klaytn node's data directory path [default : /var/klay/data]",
		Value: "/var/klay/data",
	}

	logDirFlag = cli.StringFlag{
		Name:  "log-dir",
		Usage: "klay.conf - klaytn node's log directory path [default : /var/klay/log]",
		Value: "/var/klay/log",
	}
)
