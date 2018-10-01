package main

import (
	"fmt"
	"os"

	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/accounts/keystore"
	"github.com/ground-x/go-gxplatform/cmd/utils"
	"github.com/ground-x/go-gxplatform/console"
	gxp2 "github.com/ground-x/go-gxplatform/node/cn"
	"github.com/ground-x/go-gxplatform/client"
	"github.com/ground-x/go-gxplatform/api/debug"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/metrics"
	"github.com/ground-x/go-gxplatform/metrics/prometheus"
	"github.com/ground-x/go-gxplatform/node"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/urfave/cli.v1"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	clientIdentifier = "klay" // Client identifier to advertise over the network
)

var (
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// Git tag (set via linker flags if exists)
	gitTag = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(gitCommit, "the go-klaytn command line interface")

	// flags that configure the node
	nodeFlags = []cli.Flag{
		utils.IdentityFlag,
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.BootnodesFlag,
		utils.BootnodesV4Flag,
		utils.BootnodesV5Flag,
		utils.DbTypeFlag,
		utils.DataDirFlag,
		utils.KeyStoreDirFlag,
		utils.NoUSBFlag,
		utils.DashboardEnabledFlag,
		utils.EthashCacheDirFlag,
		utils.EthashCachesInMemoryFlag,
		utils.EthashCachesOnDiskFlag,
		utils.EthashDatasetDirFlag,
		utils.EthashDatasetsInMemoryFlag,
		utils.EthashDatasetsOnDiskFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
		utils.FastSyncFlag,
		utils.LightModeFlag,
		utils.SyncModeFlag,
		utils.GCModeFlag,
		utils.LightServFlag,
		utils.LightPeersFlag,
		utils.LightKDFFlag,
		utils.LevelDBCacheSizeFlag,
		utils.TrieMemoryCacheSizeFlag,
		utils.TrieCacheGenFlag,
		utils.CacheTypeFlag,
		utils.CacheScaleFlag,
		utils.ListenPortFlag,
		utils.NodeTypeFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.CoinbaseFlag,
		utils.RewardbaseFlag,
		utils.RewardContractFlag,
		utils.GasPriceFlag,
		utils.MinerThreadsFlag,
		utils.MiningEnabledFlag,
		utils.TargetGasLimitFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
		utils.DeveloperFlag,
		utils.DeveloperPeriodFlag,
		utils.TestnetFlag,
		utils.RinkebyFlag,
		utils.VMEnableDebugFlag,
		utils.NetworkIdFlag,
		utils.RPCCORSDomainFlag,
		utils.RPCVirtualHostsFlag,
		utils.EthStatsURLFlag,
		utils.MetricsEnabledFlag,
		utils.PrometheusExporterFlag,
		utils.PrometheusExporterPortFlag,
		utils.FakePoWFlag,
		utils.NoCompactionFlag,
		utils.GpoBlocksFlag,
		utils.GpoPercentileFlag,
		utils.ExtraDataFlag,
		utils.SrvTypeFlag,
		configFileFlag,
	}

	rpcFlags = []cli.Flag{
		utils.RPCEnabledFlag,
		utils.RPCListenAddrFlag,
		utils.RPCPortFlag,
		utils.RPCApiFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
	}
)

func init() {
	// Initialize the CLI app and start Klay
	app.Action = klaytn
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2013-2018 The go-klaytn Authors"
	app.Commands = []cli.Command{
		// See chaincmd.go:
		initCommand,
		accountCommand,
		walletCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		// See versioncmd.go
		versionCommand,

		dumpConfigCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)
	app.Flags = append(app.Flags, debug.Flags...)

	app.Before = func(ctx *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		if err := debug.Setup(ctx); err != nil {
			return err
		}

		// Start prometheus exporter
		if metrics.Enabled {
			log.Info("Enabling metrics collection")
			if metrics.EnabledPrometheusExport {
				log.Info("Enabling Prometheus Exporter")
				pClient := prometheusmetrics.NewPrometheusProvider(metrics.DefaultRegistry, "klaytn",
					"", prometheus.DefaultRegisterer, 3*time.Second)
				go pClient.UpdatePrometheusMetrics()
				http.Handle("/metrics", promhttp.Handler())
				port := ctx.GlobalInt(metrics.PrometheusExporterPortFlag)

				go func() {
					err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
					if err != nil {
						log.Error("PrometheusExporter starting failed:", "port", port, "err", err)
					}
				}()
			}
		}

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(3 * time.Second)

		utils.SetupNetwork(ctx)
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		console.Stdin.Close() // Resets terminal mode.
		return nil
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// klay is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func klaytn(ctx *cli.Context) error {
	node := makeFullNode(ctx)
	startNode(ctx, node)
	node.Wait()
	return nil
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces and the
// miner.
func startNode(ctx *cli.Context, stack *node.Node) {
	debug.Memsize.Add("node", stack)

	// Start up the node itself
	utils.StartNode(stack)

	// Unlock any account specifically requested
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)

	passwords := utils.MakePasswordList(ctx)
	unlocks := strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	for i, account := range unlocks {
		if trimmed := strings.TrimSpace(account); trimmed != "" {
			unlockAccount(ctx, ks, trimmed, i, passwords)
		}
	}
	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	go func() {
		// Create a chain state reader for self-derivation
		rpcClient, err := stack.Attach()
		if err != nil {
			utils.Fatalf("Failed to attach to self: %v", err)
		}
		stateReader := client.NewClient(rpcClient)

		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				if event.Wallet.URL().Scheme == "ledger" {
					event.Wallet.SelfDerive(accounts.DefaultLedgerBaseDerivationPath, stateReader)
				} else {
					event.Wallet.SelfDerive(accounts.DefaultBaseDerivationPath, stateReader)
				}

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()

	// Start auxiliary services if enabled
	if ctx.GlobalBool(utils.MiningEnabledFlag.Name) || ctx.GlobalBool(utils.DeveloperFlag.Name) {

		var gxp *gxp2.GXP
		if err := stack.Service(&gxp); err != nil {
			utils.Fatalf("Klaytn service not running: %v", err)
		}
		// Use a reduced number of threads if requested
		if threads := ctx.GlobalInt(utils.MinerThreadsFlag.Name); threads > 0 {
			type threaded interface {
				SetThreads(threads int)
			}
			if th, ok := gxp.Engine().(threaded); ok {
				th.SetThreads(threads)
			}
		}
		if err := gxp.StartMining(true); err != nil {
			utils.Fatalf("Failed to start mining: %v", err)
		}
	} else {
		// istanbul BFT
		var gxp *gxp2.GXP
		if err := stack.Service(&gxp); err != nil {
			utils.Fatalf("Klaytn service not running: %v", err)
		}
	}
}
