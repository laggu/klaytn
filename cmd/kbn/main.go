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
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/api/debug"
	"github.com/ground-x/klaytn/cmd/utils"
	"github.com/ground-x/klaytn/cmd/utils/nodecmd"
	"github.com/ground-x/klaytn/log"
	"github.com/ground-x/klaytn/metrics"
	"github.com/ground-x/klaytn/metrics/prometheus"
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"github.com/ground-x/klaytn/networks/p2p/nat"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/urfave/cli.v1"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

func bootnode(c *cli.Context) error {
	var (
		// Local variables
		err error
		ctx = bootnodeConfig{
			// Config variables
			addr:         c.GlobalString(utils.BNAddrFlag.Name),
			genKeyPath:   c.GlobalString(utils.GenKeyFlag.Name),
			nodeKeyFile:  c.GlobalString(utils.NodeKeyFileFlag.Name),
			nodeKeyHex:   c.GlobalString(utils.NodeKeyHexFlag.Name),
			natFlag:      c.GlobalString(utils.NATFlag.Name),
			netrestrict:  c.GlobalString(utils.NetrestrictFlag.Name),
			writeAddress: c.GlobalBool(utils.WriteAddressFlag.Name),

			IPCPath:          "klay.ipc",
			DataDir:          c.GlobalString(utils.DataDirFlag.Name),
			HTTPPort:         DefaultHTTPPort,
			HTTPModules:      []string{"net"},
			HTTPVirtualHosts: []string{"localhost"},
			WSPort:           DefaultWSPort,
			WSModules:        []string{"net"},
			GRPCPort:         DefaultGRPCPort,

			Logger: log.NewModuleLogger(log.CMDKBN),
		}
	)

	setIPC(c, &ctx)
	// httptype is http or fasthttp
	if c.GlobalIsSet(utils.SrvTypeFlag.Name) {
		ctx.HTTPServerType = c.GlobalString(utils.SrvTypeFlag.Name)
	}
	setHTTP(c, &ctx)
	setWS(c, &ctx)
	setgRPC(c, &ctx)

	// Check exit condition
	switch ctx.checkCMDState() {
	case generateNodeKeySpecified:
		ctx.generateNodeKey()
	case noPrivateKeyPathSpecified:
		return errors.New("Use --nodekey or --nodekeyhex to specify a private key")
	case nodeKeyDuplicated:
		return errors.New("Options --nodekey and --nodekeyhex are mutually exclusive")
	case writeOutAddress:
		ctx.doWriteOutAddress()
	default:
		err = ctx.readNodeKey()
		if err != nil {
			return err
		}
	}

	err = ctx.validateNetworkParameter()
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

	node, err := New(&ctx)
	if err != nil {
		return err
	}
	if err := startNode(node); err != nil {
		return err
	}
	node.Wait()
	return nil
}

func startNode(node *Node) error {
	if err := node.Start(); err != nil {
		return err
	}
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		logger.Info("Got interrupt, shutting down...")
		go node.Stop()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				logger.Info("Already shutting down, interrupt more to panic.", "times", i-1)
			}
		}
	}()
	return nil
}

func main() {
	var (
		cliFlags = []cli.Flag{
			utils.SrvTypeFlag,
			utils.DataDirFlag,
			utils.GenKeyFlag,
			utils.NodeKeyFileFlag,
			utils.NodeKeyHexFlag,
			utils.WriteAddressFlag,
			utils.BNAddrFlag,
			utils.NATFlag,
			utils.NetrestrictFlag,
			utils.MetricsEnabledFlag,
			utils.PrometheusExporterFlag,
			utils.PrometheusExporterPortFlag,
		}
	)
	// TODO-Klaytn: remove `help` command
	app := utils.NewApp("", "the klaytn's bootnode command line interface")
	app.Name = "kbn"
	app.Copyright = "Copyright 2018 The klaytn Authors"
	app.UsageText = app.Name + " [global options] [commands]"
	app.Flags = append(app.Flags, cliFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, nodecmd.CommonRPCFlags...)
	app.Commands = []cli.Command{
		nodecmd.AttachCommand,
	}

	app.Action = bootnode
	app.Before = func(c *cli.Context) error {
		if err := debug.Setup(c); err != nil {
			return err
		}

		// Start prometheus exporter
		if metrics.Enabled {
			logger.Info("Enabling metrics collection")
			if metrics.EnabledPrometheusExport {
				logger.Info("Enabling Prometheus Exporter")
				pClient := prometheusmetrics.NewPrometheusProvider(metrics.DefaultRegistry, "klaytn",
					"", prometheus.DefaultRegisterer, 3*time.Second)
				go pClient.UpdatePrometheusMetrics()
				http.Handle("/metrics", promhttp.Handler())
				port := c.GlobalInt(metrics.PrometheusExporterPortFlag)

				go func() {
					err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
					if err != nil {
						logger.Error("PrometheusExporter starting failed:", "port", port, "err", err)
					}
				}()
			}
		}

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(3 * time.Second)

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
