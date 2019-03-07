// Copyright 2018 The klaytn Authors
// Copyright 2016 The go-ethereum Authors
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
// This file is derived from cmd/geth/main.go (2018/06/04).
// Modified and improved for the klaytn development.

package nodecmd

import (
	"github.com/ground-x/klaytn/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

// Common flags that configure the node
var CommonNodeFlags = []cli.Flag{
	utils.IdentityFlag,
	utils.UnlockedAccountFlag,
	utils.PasswordFileFlag,
	utils.BootnodesFlag,
	utils.BootnodesV4Flag,
	utils.DbTypeFlag,
	utils.DataDirFlag,
	utils.KeyStoreDirFlag,
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
	utils.SyncModeFlag,
	utils.GCModeFlag,
	utils.LightServFlag,
	utils.LightPeersFlag,
	utils.LightKDFFlag,
	utils.PartitionedDBFlag,
	utils.LevelDBCacheSizeFlag,
	utils.DBParallelDBWriteFlag,
	utils.TrieMemoryCacheSizeFlag,
	utils.TrieCacheGenFlag,
	utils.TrieBlockIntervalFlag,
	utils.CacheTypeFlag,
	utils.CacheScaleFlag,
	utils.CacheUsageLevelFlag,
	utils.MemorySizeFlag,
	utils.ChildChainIndexingFlag,
	utils.CacheWriteThroughFlag,
	utils.ListenPortFlag,
	utils.SubListenPortFlag,
	utils.MultiChannelUseFlag,
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
	utils.NetrestrictFlag,
	utils.NodeKeyFileFlag,
	utils.NodeKeyHexFlag,
	utils.VMEnableDebugFlag,
	utils.VMLogTargetFlag,
	utils.NetworkIdFlag,
	utils.RPCCORSDomainFlag,
	utils.RPCVirtualHostsFlag,
	utils.MetricsEnabledFlag,
	utils.PrometheusExporterFlag,
	utils.PrometheusExporterPortFlag,
	utils.ExtraDataFlag,
	utils.SrvTypeFlag,
	utils.ChainAccountAddrFlag,
	utils.AnchoringPeriodFlag,
	utils.SentChainTxsLimit,
	utils.BaobabFlag,
	ConfigFileFlag,
	utils.EnabledBridgeFlag,
	utils.IsMainBridgeFlag,
	utils.BridgeListenPortFlag,
	utils.ParentChainURLFlag,
	utils.EnableSBNFlag, //TODO-Klaytn-Node remove after the real bootnode is implemented
	utils.SBNAddrFlag,   //TODO-Klaytn-Node remove after the real bootnode is implemented
	utils.SBNPortFlag,   //TODO-Klaytn-Node remove after the real bootnode is implemented
}

// Common RPC flags
var CommonRPCFlags = []cli.Flag{
	utils.RPCEnabledFlag,
	utils.RPCListenAddrFlag,
	utils.RPCPortFlag,
	utils.RPCApiFlag,
	utils.WSEnabledFlag,
	utils.WSListenAddrFlag,
	utils.WSPortFlag,
	utils.GRPCEnabledFlag,
	utils.GRPCListenAddrFlag,
	utils.GRPCPortFlag,
	utils.WSApiFlag,
	utils.WSAllowedOriginsFlag,
	utils.IPCDisabledFlag,
	utils.IPCPathFlag,
}
