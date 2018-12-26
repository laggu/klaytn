// Copyright 2018 The go-klaytn Authors
//
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

package ranger

import (
	"fmt"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/api"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/blockchain/bloombits"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/blockchain/vm"
	"github.com/ground-x/go-gxplatform/client"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/common/bitutil"
	"github.com/ground-x/go-gxplatform/consensus"
	"github.com/ground-x/go-gxplatform/datasync/downloader"
	"github.com/ground-x/go-gxplatform/event"
	"github.com/ground-x/go-gxplatform/networks/p2p"
	"github.com/ground-x/go-gxplatform/networks/rpc"
	"github.com/ground-x/go-gxplatform/node"
	"github.com/ground-x/go-gxplatform/node/cn"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/storage/database"
	"github.com/ground-x/go-gxplatform/work"
	"github.com/hashicorp/golang-lru"
	"math/big"
	"sync"
)

const (
	bloomServiceThreads = 16
	peerCacheLimit      = 5
)

// klaytn implements the Ranger service.
type Ranger struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the klaytn

	// Handlers
	txPool          *blockchain.TxPool
	blockchain      *blockchain.BlockChain
	protocolManager *cn.ProtocolManager

	// DB interfaces
	chainDB database.DBManager // Block chain database

	eventMux       *event.TypeMux
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *blockchain.ChainIndexer       // Bloom indexer operating during block imports

	APIBackend *RangerAPIBackend

	miner    *work.Miner
	gasPrice *big.Int
	coinbase common.Address

	networkId     uint64
	netRPCService *api.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and coinbase)

	// consensus node url
	consUrl  string
	cnClient *client.Client

	// consensus
	engine consensus.Engine

	proofFeed event.Feed
	proofCh   chan NewProofEvent
	proofSub  event.Subscription

	peerCache *lru.Cache
}

// New creates a new klaytn object (including the
// initialisation of the common klaytn object)
func New(ctx *node.ServiceContext, config *Config) (*Ranger, error) {

	peerCache, _ := lru.New(peerCacheLimit)

	chainDB, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := blockchain.SetupGenesisBlock(chainDB, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}

	// NOTE-GX Now we use ChainConfig.UnitPrice from genesis.json.
	//         So let's update ranger.Config.GasPrice using ChainConfig.UnitPrice.
	config.GasPrice = new(big.Int).SetUint64(chainConfig.UnitPrice)

	logger.Info("Initialised chain configuration", "config", chainConfig)

	ranger := &Ranger{
		config:         config,
		chainDB:        chainDB,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		coinbase:       config.Gxbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   cn.NewBloomIndexer(chainDB, params.BloomBitsBlocks),
		consUrl:        config.ConsensusURL,
		proofCh:        make(chan NewProofEvent),
		peerCache:      peerCache,
	}

	ranger.engine = &RangerEngine{proofFeed: &ranger.proofFeed}

	// istanbul BFT. force to set the istanbul coinbase to node key address
	//if chainConfig.Istanbul != nil {
	//	ranger.coinbase = crypto.PubkeyToAddress(ctx.NodeKey().PublicKey)
	//}

	logger.Info("Initialising klaytn protocol", "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := chainDB.ReadDatabaseVersion()
		if bcVersion != blockchain.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run ranger upgradedb.\n", bcVersion, blockchain.BlockChainVersion)
		}
		chainDB.WriteDatabaseVersion(blockchain.BlockChainVersion)
	}
	var (
		vmConfig   = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		trieConfig = &blockchain.TrieConfig{Disabled: config.NoPruning, CacheSize: config.TrieCacheSize, BlockInterval: config.TrieBlockInterval}
	)
	ranger.blockchain, err = blockchain.NewBlockChain(chainDB, trieConfig, ranger.chainConfig, ranger.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		logger.Error("Rewinding chain to upgrade configuration", "err", compat)
		ranger.blockchain.SetHead(compat.RewindTo)
		chainDB.WriteChainConfig(genesisHash, chainConfig)
	}
	ranger.bloomIndexer.Start(ranger.blockchain)

	ranger.cnClient, err = client.Dial(ranger.consUrl)
	if err != nil {
		logger.Error("Fail to connect consensus node", "err", err)
	}

	ranger.txPool = blockchain.NewTxPool(blockchain.DefaultTxPoolConfig, ranger.chainConfig, ranger.blockchain)

	if ranger.protocolManager, err = cn.NewRangerPM(ranger.chainConfig, config.SyncMode, config.NetworkId, ranger.eventMux, ranger.engine, ranger.blockchain, chainDB); err != nil {
		return nil, err
	}

	ranger.miner = work.New(ranger, ranger.chainConfig, ranger.EventMux(), ranger.engine, ctx.NodeType())
	ranger.APIBackend = &RangerAPIBackend{ranger}

	ranger.proofSub = ranger.proofFeed.Subscribe(ranger.proofCh)
	go ranger.proofReplication()

	return ranger, nil
}

func (s *Ranger) Coinbase() (eb common.Address, err error) {
	s.lock.RLock()
	coinbase := s.coinbase
	s.lock.RUnlock()

	if coinbase != (common.Address{}) {
		// validator address in istanbul bft isn't in keystore.
		account := accounts.Account{Address: coinbase}
		_, err := s.accountManager.Find(account)
		if err == nil {
			return coinbase, nil
		}
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

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (database.DBManager, error) {
	db, err := ctx.OpenDatabase(name, config.LevelDBCacheSize, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *Ranger) StopMining()        { s.miner.Stop() }
func (s *Ranger) IsMining() bool     { return s.miner.Mining() }
func (s *Ranger) Miner() *work.Miner { return s.miner }

func (s *Ranger) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Ranger) BlockChain() *blockchain.BlockChain { return s.blockchain }
func (s *Ranger) TxPool() *blockchain.TxPool         { return s.txPool }
func (s *Ranger) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Ranger) Engine() consensus.Engine           { return s.engine }
func (s *Ranger) ChainDB() database.DBManager        { return s.chainDB }
func (s *Ranger) IsListening() bool                  { return true } // Always listening
func (s *Ranger) GxpVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Ranger) NetVersion() uint64                 { return s.networkId }
func (s *Ranger) Downloader() *downloader.Downloader { return s.protocolManager.Downloader() }

// TODO-KLAYTN drop or missing tx
func (s *Ranger) ReBroadcastTxs(transactions types.Transactions) {
	s.protocolManager.ReBroadcastTxs(transactions)
}

// APIs returns the collection of RPC services the klaytn package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Ranger) APIs() []rpc.API {
	apis := api.GetAPIs(s.APIBackend)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "klay",
			Version:   "1.0",
			Service:   NewPublicRangerAPI(s),
			Public:    true,
		},
	}...)
}

func (s *Ranger) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

func (s *Ranger) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = api.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// klaytn protocol.
func (s *Ranger) Stop() error {
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()

	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDB.Close()
	close(s.shutdownChan)

	s.proofSub.Unsubscribe()
	s.peerCache.Purge()

	return nil
}

// startBloomHandlers starts a batch of goroutines to accept bloom bit database
// retrievals from possibly a range of filters and serving the data to satisfy.
func (rn *Ranger) startBloomHandlers() {
	for i := 0; i < bloomServiceThreads; i++ {
		go func() {
			for {
				select {
				case <-rn.shutdownChan:
					return

				case request := <-rn.bloomRequests:
					task := <-request
					task.Bitsets = make([][]byte, len(task.Sections))
					for i, section := range task.Sections {
						head := rn.chainDB.ReadCanonicalHash((section+1)*params.BloomBitsBlocks - 1)
						if compVector, err := rn.chainDB.ReadBloomBits(database.BloomBitsKey(task.Bit, section, head)); err == nil {
							if blob, err := bitutil.DecompressBytes(compVector, int(params.BloomBitsBlocks)/8); err == nil {
								task.Bitsets[i] = blob
							} else {
								task.Error = err
							}
						} else {
							task.Error = err
						}
					}
					request <- task
				}
			}
		}()
	}
}
