// Copyright 2018 The go-klaytn Authors
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
// Modified and improved for the go-klaytn development.

package cn

import (
	"errors"
	"fmt"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/api"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/blockchain/bloombits"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/blockchain/vm"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/common/hexutil"
	"github.com/ground-x/go-gxplatform/consensus"
	"github.com/ground-x/go-gxplatform/consensus/gxhash"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	istanbulBackend "github.com/ground-x/go-gxplatform/consensus/istanbul/backend"
	"github.com/ground-x/go-gxplatform/crypto"
	"github.com/ground-x/go-gxplatform/datasync/downloader"
	"github.com/ground-x/go-gxplatform/event"
	"github.com/ground-x/go-gxplatform/networks/p2p"
	"github.com/ground-x/go-gxplatform/networks/rpc"
	"github.com/ground-x/go-gxplatform/node"
	"github.com/ground-x/go-gxplatform/node/cn/filters"
	"github.com/ground-x/go-gxplatform/node/cn/gasprice"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"github.com/ground-x/go-gxplatform/storage/database"
	"github.com/ground-x/go-gxplatform/work"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *blockchain.ChainIndexer)
}

// CN implements the Klaytn consensus node service.
type GXP struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the GXP

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

	APIBackend *GxpAPIBackend

	miner    *work.Miner
	gasPrice *big.Int
	coinbase common.Address

	rewardbase     common.Address
	rewardcontract common.Address

	networkId     uint64
	netRPCService *api.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (klay.g. gas price and coinbase)
}

func (s *GXP) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new GXP object (including the
// initialisation of the common GXP object)
func New(ctx *node.ServiceContext, config *Config) (*GXP, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run cn.GXP in light sync mode, use les.LightGXP")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDB, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := blockchain.SetupGenesisBlock(chainDB, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}

	// NOTE-GX Now we use ChainConfig.UnitPrice from genesis.json.
	//         So let's update cn.Config.GasPrice using ChainConfig.UnitPrice.
	config.GasPrice = new(big.Int).SetUint64(chainConfig.UnitPrice)

	logger.Info("Initialised chain configuration", "config", chainConfig)

	gxp := &GXP{
		config:         config,
		chainDB:        chainDB,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, config, chainConfig, chainDB),
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		coinbase:       config.Gxbase,
		rewardbase:     config.Rewardbase,
		rewardcontract: config.RewardContract,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDB, params.BloomBitsBlocks),
	}

	// istanbul BFT. force to set the istanbul coinbase to node key address
	if chainConfig.Istanbul != nil {
		gxp.coinbase = crypto.PubkeyToAddress(ctx.NodeKey().PublicKey)
	}

	logger.Info("Initialising Klaytn protocol", "versions", gxp.engine.Protocol().Versions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := chainDB.ReadDatabaseVersion()
		if bcVersion != blockchain.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run klay upgradedb.\n", bcVersion, blockchain.BlockChainVersion)
		}
		chainDB.WriteDatabaseVersion(blockchain.BlockChainVersion)
	}
	var (
		vmConfig   = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		trieConfig = &blockchain.TrieConfig{Disabled: config.NoPruning, CacheSize: config.TrieCacheSize, BlockInterval: config.TrieBlockInterval}
	)
	gxp.blockchain, err = blockchain.NewBlockChain(chainDB, trieConfig, gxp.chainConfig, gxp.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		logger.Error("Rewinding chain to upgrade configuration", "err", compat)
		gxp.blockchain.SetHead(compat.RewindTo)
		chainDB.WriteChainConfig(genesisHash, chainConfig)
	}
	gxp.bloomIndexer.Start(gxp.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	gxp.txPool = blockchain.NewTxPool(config.TxPool, gxp.chainConfig, gxp.blockchain)

	if gxp.protocolManager, err = NewProtocolManager(gxp.chainConfig, config.SyncMode, config.NetworkId, gxp.eventMux, gxp.txPool, gxp.engine, gxp.blockchain, chainDB, ctx.NodeType()); err != nil {
		return nil, err
	}
	gxp.protocolManager.wsendpoint = config.WsEndpoint

	wallet, err := gxp.RewardbaseWallet()
	if err != nil {
		logger.Error("find err", "err", err)
	} else {
		gxp.protocolManager.SetRewardbaseWallet(wallet)
	}
	gxp.protocolManager.SetRewardbase(gxp.rewardbase)
	gxp.protocolManager.SetRewardContract(gxp.rewardcontract)

	// TODO-GX improve to handle drop transaction on network traffic in BN,GN,RN
	gxp.miner = work.New(gxp, gxp.chainConfig, gxp.EventMux(), gxp.engine, ctx.NodeType())
	// istanbul BFT
	gxp.miner.SetExtra(makeExtraData(config.ExtraData, gxp.chainConfig.IsBFT))

	gxp.APIBackend = &GxpAPIBackend{gxp, nil}

	gpoParams := config.GPO

	// NOTE-GX Now we use ChainConfig.UnitPrice from genesis.json and updated config.GasPrice with same value.
	//         So let's override gpoParams.Default with config.GasPrice
	gpoParams.Default = config.GasPrice

	gxp.APIBackend.gpo = gasprice.NewOracle(gxp.APIBackend, gpoParams)

	return gxp, nil
}

// istanbul BFT
func makeExtraData(extra []byte, isBFT bool) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"klay",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.GetMaximumExtraDataSize(isBFT) {
		logger.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.GetMaximumExtraDataSize(isBFT))
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (database.DBManager, error) {
	db, err := ctx.OpenDatabase(name, config.LevelDBCacheSize, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for a klaytn service
func CreateConsensusEngine(ctx *node.ServiceContext, config *Config, chainConfig *params.ChainConfig, db database.DBManager) consensus.Engine {
	// If proof-of-authority is requested, set it up
	//if chainConfig.Clique != nil {
	//	return clique.New(chainConfig.Clique, db)
	//}
	if chainConfig.Istanbul != nil {
		if chainConfig.Istanbul.Epoch != 0 {
			config.Istanbul.Epoch = chainConfig.Istanbul.Epoch
		}
		config.Istanbul.ProposerPolicy = istanbul.ProposerPolicy(chainConfig.Istanbul.ProposerPolicy)
		config.Istanbul.SubGroupSize = chainConfig.Istanbul.SubGroupSize
		return istanbulBackend.New(config.Rewardbase, config.RewardContract, &config.Istanbul, ctx.NodeKey(), db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.Gxhash.PowMode == gxhash.ModeFake:
		logger.Debug("Gxhash used in fake mode")
		return gxhash.NewFaker()
	case config.Gxhash.PowMode == gxhash.ModeTest:
		logger.Debug("Gxhash used in test mode")
		return gxhash.NewTester()
	case config.Gxhash.PowMode == gxhash.ModeShared:
		logger.Debug("Gxhash used in shared mode")
		return gxhash.NewShared()
	default:
		engine := gxhash.New(gxhash.Config{
			CacheDir:       ctx.ResolvePath(config.Gxhash.CacheDir),
			CachesInMem:    config.Gxhash.CachesInMem,
			CachesOnDisk:   config.Gxhash.CachesOnDisk,
			DatasetDir:     config.Gxhash.DatasetDir,
			DatasetsInMem:  config.Gxhash.DatasetsInMem,
			DatasetsOnDisk: config.Gxhash.DatasetsOnDisk,
		})
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *GXP) APIs() []rpc.API {
	apis := api.GetAPIs(s.APIBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "klay",
			Version:   "1.0",
			Service:   NewPublicGXPAPI(s),
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
		},
	}...)
}

func (s *GXP) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *GXP) Coinbase() (eb common.Address, err error) {
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

func (s *GXP) Rewardbase() (eb common.Address, err error) {
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

func (s *GXP) RewardContract() (addr common.Address, err error) {
	s.lock.RLock()
	rewardcontract := s.rewardcontract
	s.lock.RUnlock()

	if rewardcontract != (common.Address{}) {
		return rewardcontract, nil
	}
	return common.Address{}, nil
}

func (s *GXP) RewardbaseWallet() (accounts.Wallet, error) {
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
func (s *GXP) SetCoinbase(coinbase common.Address) {
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

func (s *GXP) SetRewardbase(rewardbase common.Address) {
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

func (s *GXP) SetRewardContract(addr common.Address) {
	s.lock.Lock()
	s.rewardcontract = addr
	s.lock.Unlock()

	s.protocolManager.SetRewardContract(s.rewardcontract)
	//TODO-GX broadcast another CN with authentication rule
	//TODO-GX add governance feature
}

func (s *GXP) StartMining(local bool) error {
	eb, err := s.Coinbase()
	if err != nil {
		logger.Error("Cannot start mining without coinbase", "err", err)
		return fmt.Errorf("coinbase missing: %v", err)
	}
	//if clique, ok := s.engine.(*clique.Clique); ok {
	//	rewardwallet, err := s.accountManager.Find(accounts.Account{Address: eb})
	//	if rewardwallet == nil || err != nil {
	//		logger.Error("Coinbase account unavailable locally", "err", err)
	//		return fmt.Errorf("signer missing: %v", err)
	//	}
	//	clique.Authorize(eb, rewardwallet.SignHash)
	//}
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

func (s *GXP) StopMining()        { s.miner.Stop() }
func (s *GXP) IsMining() bool     { return s.miner.Mining() }
func (s *GXP) Miner() *work.Miner { return s.miner }

func (s *GXP) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *GXP) BlockChain() *blockchain.BlockChain { return s.blockchain }
func (s *GXP) TxPool() *blockchain.TxPool         { return s.txPool }
func (s *GXP) EventMux() *event.TypeMux           { return s.eventMux }
func (s *GXP) Engine() consensus.Engine           { return s.engine }
func (s *GXP) ChainDB() database.DBManager        { return s.chainDB }
func (s *GXP) IsListening() bool                  { return true } // Always listening
func (s *GXP) GxpVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *GXP) NetVersion() uint64                 { return s.networkId }
func (s *GXP) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *GXP) ReBroadcastTxs(transactions types.Transactions) {
	s.protocolManager.ReBroadcastTxs(transactions)
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *GXP) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// GXP protocol implementation.
func (s *GXP) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = api.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// GXP protocol.
func (s *GXP) Stop() error {
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
