// Copyright 2018 The klaytn Authors
// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
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
//
// This file is derived from eth/backend.go (2018/06/04).
// Modified and improved for the klaytn development.

package cn

import (
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/accounts"
	"github.com/ground-x/klaytn/api"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/bloombits"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/vm"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/hexutil"
	"github.com/ground-x/klaytn/consensus"
	"github.com/ground-x/klaytn/consensus/istanbul"
	istanbulBackend "github.com/ground-x/klaytn/consensus/istanbul/backend"
	"github.com/ground-x/klaytn/contracts/reward"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/datasync/downloader"
	"github.com/ground-x/klaytn/event"
	"github.com/ground-x/klaytn/governance"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/networks/rpc"
	"github.com/ground-x/klaytn/node"
	"github.com/ground-x/klaytn/node/cn/filters"
	"github.com/ground-x/klaytn/node/cn/gasprice"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/database"
	"github.com/ground-x/klaytn/work"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
)

type LesServer interface {
	Start(srvr p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *blockchain.ChainIndexer)
}

// CN implements the Klaytn consensus node service.
type CN struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the CN

	// Handlers
	txPool          *blockchain.TxPool
	blockchain      *blockchain.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDB database.DBManager // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *blockchain.ChainIndexer       // Bloom indexer operating during block imports

	APIBackend *CNAPIBackend

	miner    *work.Miner
	gasPrice *big.Int
	coinbase common.Address

	rewardbase common.Address

	networkId     uint64
	netRPCService *api.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (klay.g. gas price and coinbase)

	components []interface{}

	governance *governance.Governance
}

func (s *CN) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new CN object (including the
// initialisation of the common CN object)
func New(ctx *node.ServiceContext, config *Config) (*CN, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run cn.CN in light sync mode, use les.LightCN")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDB := CreateDB(ctx, config, "chaindata")

	chainConfig, genesisHash, genesisErr := blockchain.SetupGenesisBlock(chainDB, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}

	// NOTE-Klaytn Now we use ChainConfig.UnitPrice from genesis.json.
	//         So let's update cn.Config.GasPrice using ChainConfig.UnitPrice.
	config.GasPrice = new(big.Int).SetUint64(chainConfig.UnitPrice)

	logger.Info("Initialised chain configuration", "config", chainConfig)
	governance := governance.NewGovernance(chainConfig)

	cn := &CN{
		config:         config,
		chainDB:        chainDB,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, config, chainConfig, chainDB, governance, ctx.NodeType()),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		coinbase:       config.Gxbase,
		rewardbase:     config.Rewardbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDB, params.BloomBitsBlocks),
		governance:     governance,
	}

	// istanbul BFT. force to set the istanbul coinbase to node key address
	if chainConfig.Istanbul != nil {
		cn.coinbase = crypto.PubkeyToAddress(ctx.NodeKey().PublicKey)
		governance.SetNodeAddress(cn.coinbase)
	}

	logger.Info("Initialising Klaytn protocol", "versions", cn.engine.Protocol().Versions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := chainDB.ReadDatabaseVersion()
		if bcVersion != blockchain.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run klay upgradedb.\n", bcVersion, blockchain.BlockChainVersion)
		}
		chainDB.WriteDatabaseVersion(blockchain.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &blockchain.CacheConfig{StateDBCaching: config.StateDBCaching, ArchiveMode: config.NoPruning, CacheSize: config.TrieCacheSize, BlockInterval: config.TrieBlockInterval}
	)
	var err error
	cn.blockchain, err = blockchain.NewBlockChain(chainDB, cacheConfig, cn.chainConfig, cn.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		logger.Error("Rewinding chain to upgrade configuration", "err", compat)
		cn.blockchain.SetHead(compat.RewindTo)
		chainDB.WriteChainConfig(genesisHash, chainConfig)
	}
	cn.bloomIndexer.Start(cn.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	cn.txPool = blockchain.NewTxPool(config.TxPool, cn.chainConfig, cn.blockchain)

	if cn.protocolManager, err = NewProtocolManager(cn.chainConfig, config.SyncMode, config.NetworkId, cn.eventMux, cn.txPool, cn.engine, cn.blockchain, chainDB, ctx.NodeType()); err != nil {
		return nil, err
	}

	// Set AcceptTxs flag in 1CN case to receive tx propagation.
	if chainConfig.Istanbul != nil {
		istanbulExtra, err := types.ExtractIstanbulExtra(cn.blockchain.Genesis().Header())
		if err != nil {
			logger.Error("Failed to decode IstanbulExtra", "err", err)
		} else {
			if len(istanbulExtra.Validators) == 1 {
				atomic.StoreUint32(&cn.protocolManager.acceptTxs, 1)
			}
		}
	}

	cn.protocolManager.wsendpoint = config.WsEndpoint

	wallet, err := cn.RewardbaseWallet()
	if err != nil {
		logger.Error("find err", "err", err)
	} else {
		cn.protocolManager.SetRewardbaseWallet(wallet)
	}
	cn.protocolManager.SetRewardbase(cn.rewardbase)

	if chainConfig.Istanbul != nil && cn.chainConfig.Governance.Istanbul.ProposerPolicy == uint64(istanbul.WeightedRandom) {
		reward.Subscribe(cn.blockchain)
	}

	// TODO-Klaytn improve to handle drop transaction on network traffic in PN and EN
	cn.miner = work.New(cn, cn.chainConfig, cn.EventMux(), cn.engine, ctx.NodeType())
	// istanbul BFT
	cn.miner.SetExtra(makeExtraData(config.ExtraData))

	cn.APIBackend = &CNAPIBackend{cn, nil}

	gpoParams := config.GPO

	// NOTE-Klaytn Now we use ChainConfig.UnitPrice from genesis.json and updated config.GasPrice with same value.
	//         So let's override gpoParams.Default with config.GasPrice
	gpoParams.Default = config.GasPrice

	cn.APIBackend.gpo = gasprice.NewOracle(cn.APIBackend, gpoParams)
	//@TODO Klaytn add core component
	cn.addComponent(cn.blockchain)
	cn.addComponent(cn.txPool)

	return cn, nil
}

// add component which may be used in another service component
func (s *CN) addComponent(component interface{}) {
	s.components = append(s.components, component)
}

func (s *CN) Components() []interface{} {
	return s.components
}

func (s *CN) SetComponents(component []interface{}) {
	// do nothing
}

// istanbul BFT
func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"klay",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.GetMaximumExtraDataSize() {
		logger.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.GetMaximumExtraDataSize())
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) database.DBManager {
	dbc := &database.DBConfig{Dir: name, DBType: database.LevelDB, ParallelDBWrite: config.ParallelDBWrite, Partitioned: config.PartitionedDB,
		LevelDBCacheSize: config.LevelDBCacheSize, LevelDBHandles: config.DatabaseHandles,
		ChildChainIndexing: config.ChildChainIndexing}
	return ctx.OpenDatabase(dbc)
}

// CreateConsensusEngine creates the required type of consensus engine instance for a klaytn service
func CreateConsensusEngine(ctx *node.ServiceContext, config *Config, chainConfig *params.ChainConfig, db database.DBManager, gov *governance.Governance, nodetype p2p.ConnType) consensus.Engine {
	// Only istanbul  BFT is allowed in the main net. PoA is supported by service chain
	if chainConfig.Governance == nil {
		chainConfig.Governance = governance.GetDefaultGovernanceConfig(params.UseIstanbul)
	} else {
		if chainConfig.Governance.Istanbul != nil {
			config.Istanbul.Epoch = chainConfig.Governance.Istanbul.Epoch
			config.Istanbul.ProposerPolicy = istanbul.ProposerPolicy(chainConfig.Governance.Istanbul.ProposerPolicy)
			config.Istanbul.SubGroupSize = chainConfig.Governance.Istanbul.SubGroupSize
		} else {
			chainConfig.Governance.Istanbul = governance.GetDefaultIstanbulConfig()
		}
	}
	return istanbulBackend.New(config.Rewardbase, &config.Istanbul, ctx.NodeKey(), db, gov, nodetype)
}

// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *CN) APIs() []rpc.API {
	apis := api.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "klay",
			Version:   "1.0",
			Service:   NewPublicKlayAPI(s),
			Public:    true,
		}, {
			Namespace: "klay",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "klay",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "klay",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.APIBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		}, {
			Namespace: "governance",
			Version:   "1.0",
			Service:   governance.NewGovernanceAPI(s.governance),
			Public:    true,
		},
	}...)
}

func (s *CN) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *CN) Coinbase() (eb common.Address, err error) {
	s.lock.RLock()
	coinbase := s.coinbase
	s.lock.RUnlock()

	if coinbase != (common.Address{}) {
		return coinbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			coinbase := accounts[0].Address

			s.lock.Lock()
			s.coinbase = coinbase
			s.lock.Unlock()

			logger.Info("Coinbase automatically configured", "address", coinbase)
			return coinbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("coinbase must be explicitly specified")
}

func (s *CN) Rewardbase() (eb common.Address, err error) {
	s.lock.RLock()
	rewardbase := s.rewardbase
	s.lock.RUnlock()

	if rewardbase != (common.Address{}) {
		return rewardbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			rewardbase := accounts[0].Address

			s.lock.Lock()
			s.rewardbase = rewardbase
			s.lock.Unlock()

			logger.Info("Rewardbase automatically configured", "address", rewardbase)
			return rewardbase, nil
		}
	}

	return common.Address{}, fmt.Errorf("rewardbase must be explicitly specified")
}

func (s *CN) RewardbaseWallet() (accounts.Wallet, error) {
	coinbase, err := s.Rewardbase()
	if err != nil {
		return nil, err
	}

	account := accounts.Account{Address: coinbase}
	wallet, err := s.AccountManager().Find(account)
	if err != nil {
		logger.Error("find err", "err", err)
		return nil, err
	}
	return wallet, nil
}

// SetRewardbase sets the mining reward address.
func (s *CN) SetCoinbase(coinbase common.Address) {
	s.lock.Lock()
	// istanbul BFT
	if _, ok := s.engine.(consensus.Istanbul); ok {
		logger.Error("Cannot set coinbase in Istanbul consensus")
		return
	}
	s.coinbase = coinbase
	s.lock.Unlock()

	s.miner.SetCoinbase(coinbase)
}

func (s *CN) SetRewardbase(rewardbase common.Address) {
	s.lock.Lock()
	s.rewardbase = rewardbase
	s.lock.Unlock()
	wallet, err := s.RewardbaseWallet()
	if err != nil {
		logger.Error("find err", "err", err)
	}
	s.protocolManager.SetRewardbase(rewardbase)
	s.protocolManager.SetRewardbaseWallet(wallet)
}

func (s *CN) StartMining(local bool) error {
	eb, err := s.Coinbase()
	if eb == (common.Address{}) {
		// TODO-Klaytn This zero address is only for the test code that uses gxhash.
		//             Remove this when cleaning up the gxhash code and its test code.
		eb = common.HexToAddress("0x0000000000000000000000000000000000000000")
	} else if err != nil {
		return fmt.Errorf("error on getting coinbase: %v", err)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so none will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *CN) StopMining()        { s.miner.Stop() }
func (s *CN) IsMining() bool     { return s.miner.Mining() }
func (s *CN) Miner() *work.Miner { return s.miner }

func (s *CN) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *CN) BlockChain() *blockchain.BlockChain { return s.blockchain }
func (s *CN) TxPool() *blockchain.TxPool         { return s.txPool }
func (s *CN) EventMux() *event.TypeMux           { return s.eventMux }
func (s *CN) Engine() consensus.Engine           { return s.engine }
func (s *CN) ChainDB() database.DBManager        { return s.chainDB }
func (s *CN) IsListening() bool                  { return true } // Always listening
func (s *CN) ProtocolVersion() int               { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *CN) NetVersion() uint64                 { return s.networkId }
func (s *CN) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *CN) ReBroadcastTxs(transactions types.Transactions) {
	s.protocolManager.ReBroadcastTxs(transactions)
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *CN) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Klaytn protocol implementation.
func (s *CN) Start(srvr p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = api.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers()
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Klaytn protocol.
func (s *CN) Stop() error {
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDB.Close()
	close(s.shutdownChan)

	return nil
}
