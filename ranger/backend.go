package ranger

import (
	"math/big"
	"github.com/ground-x/go-gxplatform/internal/gxapi"
	"sync"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/core"
	"github.com/ground-x/go-gxplatform/gxdb"
	"github.com/ground-x/go-gxplatform/gxp"
	"github.com/ground-x/go-gxplatform/event"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/core/bloombits"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/miner"
	"github.com/ground-x/go-gxplatform/node"
	"github.com/ground-x/go-gxplatform/rpc"
	"fmt"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/core/rawdb"
	"github.com/ground-x/go-gxplatform/core/vm"
	"github.com/ground-x/go-gxplatform/gxp/downloader"
	"github.com/ground-x/go-gxplatform/consensus"
	"github.com/ground-x/go-gxplatform/gxpclient"
	"github.com/ground-x/go-gxplatform/p2p"
	"github.com/ground-x/go-gxplatform/common/bitutil"
	"github.com/hashicorp/golang-lru"
)

const (
	bloomServiceThreads = 16
	peerCacheLimit      = 5
)

// GXP implements the Ranger service.
type Ranger struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan chan bool // Channel for shutting down the GXP

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *gxp.ProtocolManager

	// DB interfaces
	chainDb gxdb.Database // Block chain database

	eventMux       *event.TypeMux
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	APIBackend *RangerAPIBackend

	miner    *miner.Miner
	gasPrice *big.Int
	coinbase common.Address

	networkId     uint64
	netRPCService *gxapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (klay.g. gas price and coinbase)

	// consensus node url
	consUrl string
	cnClient *gxpclient.Client

	// consensus
	engine consensus.Engine

	proofFeed event.Feed
	proofCh chan NewProofEvent
	proofSub event.Subscription

	peerCache  *lru.Cache
}

// New creates a new GXP object (including the
// initialisation of the common GXP object)
func New(ctx *node.ServiceContext, config *Config) (*Ranger, error) {


	peerCache, _ := lru.New(peerCacheLimit)

	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	ranger := &Ranger{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		shutdownChan:   make(chan bool),
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		coinbase:       config.Gxbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   gxp.NewBloomIndexer(chainDb, params.BloomBitsBlocks),
		consUrl:        config.ConsensusURL,
		proofCh:        make(chan NewProofEvent),
		peerCache:      peerCache,
	}

	ranger.engine = &RangerEngine{proofFeed:&ranger.proofFeed}

	// istanbul BFT. force to set the istanbul coinbase to node key address
	//if chainConfig.Istanbul != nil {
	//	ranger.coinbase = crypto.PubkeyToAddress(ctx.NodeKey().PublicKey)
	//}

	log.Info("Initialising GXP protocol" , "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := rawdb.ReadDatabaseVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run klay upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		rawdb.WriteDatabaseVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	ranger.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, ranger.chainConfig, ranger.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		ranger.blockchain.SetHead(compat.RewindTo)
		rawdb.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	ranger.bloomIndexer.Start(ranger.blockchain)

	ranger.cnClient, err = gxpclient.Dial(ranger.consUrl)
	if err != nil {
		log.Error("Fail to connect consensus node","err",err)
	}

	ranger.txPool = core.NewTxPool(core.DefaultTxPoolConfig , ranger.chainConfig, ranger.blockchain)

	if ranger.protocolManager, err = gxp.NewRangerPM(ranger.chainConfig, config.SyncMode, config.NetworkId, ranger.eventMux, ranger.engine, ranger.blockchain, chainDb); err != nil {
		return nil, err
	}

	ranger.miner = miner.New(ranger, ranger.chainConfig, ranger.EventMux(), ranger.engine)
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
		_ , err := s.accountManager.Find(account)
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

			log.Info("Coinbase automatically configured", "address", coinbase)
			return coinbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("coinbase must be explicitly specified")
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (gxdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*gxdb.LDBDatabase); ok {
		db.Meter("klay/db/chaindata/")
	}
	return db, nil
}

func (s *Ranger) StopMining()         { s.miner.Stop() }
func (s *Ranger) IsMining() bool      { return s.miner.Mining() }
func (s *Ranger) Miner() *miner.Miner { return s.miner }

func (s *Ranger) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Ranger) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Ranger) TxPool() *core.TxPool               { return s.txPool }
func (s *Ranger) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Ranger) Engine() consensus.Engine           { return s.engine }
func (s *Ranger) ChainDb() gxdb.Database             { return s.chainDb }
func (s *Ranger) IsListening() bool                  { return true } // Always listening
func (s *Ranger) GxpVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Ranger) NetVersion() uint64                 { return s.networkId }
func (s *Ranger) Downloader() *downloader.Downloader { return s.protocolManager.Downloader() }


// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Ranger) APIs() []rpc.API {
	apis := gxapi.GetAPIs(s.APIBackend)

	// Append all the local APIs and return
	return append(apis,[]rpc.API{
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
	s.netRPCService = gxapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// GXP protocol.
func (s *Ranger) Stop() error {
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()

	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
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
						head := rawdb.ReadCanonicalHash(rn.chainDb, (section+1)*params.BloomBitsBlocks-1)
						if compVector, err := rawdb.ReadBloomBits(rn.chainDb, task.Bit, section, head); err == nil {
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
