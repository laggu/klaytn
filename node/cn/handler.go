// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
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
// This file is derived from eth/handler.go (2018/06/04).
// Modified and improved for the klaytn development.

package cn

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/accounts"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/consensus"
	"github.com/ground-x/klaytn/crypto"
	"github.com/ground-x/klaytn/datasync/downloader"
	"github.com/ground-x/klaytn/datasync/fetcher"
	"github.com/ground-x/klaytn/event"
	"github.com/ground-x/klaytn/networks/p2p"
	"github.com/ground-x/klaytn/networks/p2p/discover"
	"github.com/ground-x/klaytn/node"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/database"
	"math"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	softResponseLimit = 2 * 1024 * 1024 // Target maximum size of returned blocks, headers or node data.
	estHeaderRlpSize  = 500             // Approximate size of an RLP encoded block header

	// txChanSize is the size of channel listening to NewTxsEvent.
	// The number is referenced from the size of tx pool.
	txChanSize = 4096

	concurrentPerPeer  = 3
	channelSizePerPeer = 20
)

var (
	daoChallengeTimeout = 15 * time.Second // Time allowance for a node to reply to the DAO handshake challenge
)

// errIncompatibleConfig is returned if the requested protocols and configs are
// not compatible (low protocol version restrictions and high requirements).
var errIncompatibleConfig = errors.New("incompatible configuration")

func errResp(code errCode, format string, v ...interface{}) error {
	return fmt.Errorf("%v - %v", code, fmt.Sprintf(format, v...))
}

type ProtocolManager struct {
	networkId uint64

	fastSync  uint32 // Flag whether fast sync is enabled (gets disabled if we already have blocks)
	acceptTxs uint32 // Flag whether we're considered synchronised (enables transaction processing)

	txpool      txPool
	blockchain  *blockchain.BlockChain
	chainconfig *params.ChainConfig
	maxPeers    int

	downloader *downloader.Downloader
	fetcher    *fetcher.Fetcher
	peers      *peerSet

	SubProtocols []p2p.Protocol

	eventMux      *event.TypeMux
	txsCh         chan blockchain.NewTxsEvent
	txsSub        event.Subscription
	minedBlockSub *event.TypeMuxSubscription

	// channels for fetcher, syncer, txsyncLoop
	newPeerCh   chan Peer
	txsyncCh    chan *txsync
	quitSync    chan struct{}
	noMorePeers chan struct{}

	// wait group is used for graceful shutdowns during downloading
	// and processing
	wg sync.WaitGroup
	// istanbul BFT
	engine consensus.Engine

	rewardcontract common.Address
	rewardbase     common.Address
	rewardwallet   accounts.Wallet

	wsendpoint string

	nodetype p2p.ConnType

	scpm ServiceChainProtocolManager
}

// NewProtocolManager returns a new klaytn sub protocol manager. The klaytn sub protocol manages peers capable
// with the klaytn network.
func NewProtocolManager(config *params.ChainConfig, mode downloader.SyncMode, networkId uint64, mux *event.TypeMux, txpool txPool, engine consensus.Engine, blockchain *blockchain.BlockChain, chainDB database.DBManager, nodetype p2p.ConnType, scc *ServiceChainConfig) (*ProtocolManager, error) {
	// Create the protocol maanger with the base fields
	manager := &ProtocolManager{
		networkId:   networkId,
		eventMux:    mux,
		txpool:      txpool,
		blockchain:  blockchain,
		chainconfig: config,
		peers:       newPeerSet(),
		newPeerCh:   make(chan Peer),
		noMorePeers: make(chan struct{}),
		txsyncCh:    make(chan *txsync),
		quitSync:    make(chan struct{}),
		engine:      engine,
		nodetype:    nodetype,
		scpm:        NewServiceChainProtocolManager(scc),
	}

	// istanbul BFT
	if handler, ok := engine.(consensus.Handler); ok {
		handler.SetBroadcaster(manager, manager.nodetype)
	}

	// Figure out whether to allow fast sync or not
	if mode == downloader.FastSync && blockchain.CurrentBlock().NumberU64() > 0 {
		logger.Error("Blockchain not empty, fast sync disabled")
		mode = downloader.FullSync
	}
	if mode == downloader.FastSync {
		manager.fastSync = uint32(1)
	}
	// istanbul BFT
	protocol := engine.Protocol()
	// Initiate a sub-protocol for every implemented version we can handle
	manager.SubProtocols = make([]p2p.Protocol, 0, len(protocol.Versions))
	for i, version := range protocol.Versions {
		// Skip protocol version if incompatible with the mode of operation
		if mode == downloader.FastSync && version < klay63 {
			continue
		}
		// Compatible; initialise the sub-protocol
		version := version
		manager.SubProtocols = append(manager.SubProtocols, p2p.Protocol{
			Name:    protocol.Name,
			Version: version,
			Length:  protocol.Lengths[i],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				peer := manager.newPeer(int(version), p, rw)
				pubKey, err := p.ID().Pubkey()
				if err != nil {
					if p.ConnType() == node.CONSENSUSNODE {
						return err
					}
					peer.SetAddr(common.Address{})
				} else {
					addr := crypto.PubkeyToAddress(*pubKey)
					peer.SetAddr(addr)
				}
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					return manager.handle(peer)
				case <-manager.quitSync:
					return p2p.DiscQuitting
				}
			},
			RunWithRWs: func(p *p2p.Peer, rws []p2p.MsgReadWriter) error {
				peer, err := manager.newPeerWithRWs(int(version), p, rws)
				if err != nil {
					return err
				}
				pubKey, err := p.ID().Pubkey()
				if err != nil {
					if p.ConnType() == node.CONSENSUSNODE {
						return err
					}
					peer.SetAddr(common.Address{})
				} else {
					addr := crypto.PubkeyToAddress(*pubKey)
					peer.SetAddr(addr)
				}
				select {
				case manager.newPeerCh <- peer:
					manager.wg.Add(1)
					defer manager.wg.Done()
					return peer.Handle(manager)
				case <-manager.quitSync:
					return p2p.DiscQuitting
				}
			},
			NodeInfo: func() interface{} {
				return manager.NodeInfo()
			},
			PeerInfo: func(id discover.NodeID) interface{} {
				if p := manager.peers.Peer(fmt.Sprintf("%x", id[:8])); p != nil {
					return p.Info()
				}
				return nil
			},
		})
	}

	if len(manager.SubProtocols) == 0 {
		return nil, errIncompatibleConfig
	}
	// Construct the different synchronisation mechanisms
	manager.downloader = downloader.New(mode, chainDB, manager.eventMux, blockchain, nil, manager.removePeer)

	validator := func(header *types.Header) error {
		return engine.VerifyHeader(blockchain, header, true)
	}
	heighter := func() uint64 {
		return blockchain.CurrentBlock().NumberU64()
	}
	inserter := func(blocks types.Blocks) (int, error) {
		// If fast sync is running, deny importing weird blocks
		if atomic.LoadUint32(&manager.fastSync) == 1 {
			logger.Warn("Discarded bad propagated block", "number", blocks[0].Number(), "hash", blocks[0].Hash())
			return 0, nil
		}
		atomic.StoreUint32(&manager.acceptTxs, 1) // Mark initial sync done on any fetcher import
		return manager.blockchain.InsertChain(blocks)
	}
	manager.fetcher = fetcher.New(blockchain.GetBlockByHash, validator, manager.BroadcastBlock, heighter, inserter, manager.removePeer)

	return manager, nil
}

// istanbul BFT
func (pm *ProtocolManager) RegisterValidator(conType p2p.ConnType, validator p2p.PeerTypeValidator) {
	pm.peers.validator[conType] = validator
}

func (pm *ProtocolManager) getWSEndPoint() string {
	return pm.wsendpoint
}

func (pm *ProtocolManager) SetRewardContract(addr common.Address) {
	pm.rewardcontract = addr
}

func (pm *ProtocolManager) SetRewardbase(addr common.Address) {
	pm.rewardbase = addr
}

func (pm *ProtocolManager) SetRewardbaseWallet(wallet accounts.Wallet) {
	pm.rewardwallet = wallet
}

func (pm *ProtocolManager) removePeer(id string) {
	// Short circuit if the peer was already removed
	peer := pm.peers.Peer(id)
	if peer == nil {
		return
	}
	logger.Debug("Removing klaytn peer", "peer", id)

	// Unregister the peer from the downloader and peer set
	pm.downloader.UnregisterPeer(id)
	if err := pm.peers.Unregister(id); err != nil {
		logger.Error("Peer removal failed", "peer", id, "err", err)
	}
	// Hard disconnect at the networking layer
	if peer != nil {
		peer.GetP2PPeer().Disconnect(p2p.DiscUselessPeer)
	}
}

// getChainID returns the current chain id.
func (pm *ProtocolManager) getChainID() *big.Int {
	return pm.blockchain.Config().ChainID
}

func (pm *ProtocolManager) Start(maxPeers int) {
	pm.maxPeers = maxPeers

	// broadcast transactions
	pm.txsCh = make(chan blockchain.NewTxsEvent, txChanSize)
	pm.txsSub = pm.txpool.SubscribeNewTxsEvent(pm.txsCh)
	go pm.txBroadcastLoop()

	// broadcast mined blocks
	pm.minedBlockSub = pm.eventMux.Subscribe(blockchain.NewMinedBlockEvent{})
	go pm.minedBroadcastLoop()

	// start sync handlers
	go pm.syncer()
	go pm.txsyncLoop()
}

func (pm *ProtocolManager) Stop() {
	logger.Info("Stopping klaytn protocol")

	pm.txsSub.Unsubscribe()        // quits txBroadcastLoop
	pm.minedBlockSub.Unsubscribe() // quits blockBroadcastLoop

	// Quit the sync loop.
	// After this send has completed, no new peers will be accepted.
	pm.noMorePeers <- struct{}{}

	// Quit fetcher, txsyncLoop.
	close(pm.quitSync)

	// Disconnect existing sessions.
	// This also closes the gate for any new registrations on the peer set.
	// sessions which are already established but not added to pm.peers yet
	// will exit when they try to register.
	pm.peers.Close()

	// Wait for all peer handler goroutines and the loops to come down.
	pm.wg.Wait()

	logger.Info("klaytn protocol stopped")
}

func (pm *ProtocolManager) newPeer(pv int, p *p2p.Peer, rw p2p.MsgReadWriter) Peer {
	return newPeer(pv, p, newMeteredMsgWriter(rw))
}

// newPeerWithRWs creates a new Peer object with a slice of p2p.MsgReadWriter.
func (pm *ProtocolManager) newPeerWithRWs(pv int, p *p2p.Peer, rws []p2p.MsgReadWriter) (Peer, error) {
	meteredRWs := make([]p2p.MsgReadWriter, 0, len(rws))
	for _, rw := range rws {
		meteredRWs = append(meteredRWs, newMeteredMsgWriter(rw))
	}
	return newPeerWithRWs(pv, p, meteredRWs)
}

// handle is the callback invoked to manage the life cycle of a Klaytn peer. When
// this function terminates, the peer is disconnected.
func (pm *ProtocolManager) handle(p Peer) error {
	// Ignore maxPeers if this is a trusted peer
	if pm.peers.Len() >= pm.maxPeers && !p.GetP2PPeer().Info().Network.Trusted {
		return p2p.DiscTooManyPeers
	}
	p.GetP2PPeer().Log().Debug("klaytn peer connected", "name", p.GetP2PPeer().Name())

	// Execute the handshake
	var (
		genesis = pm.blockchain.Genesis()
		head    = pm.blockchain.CurrentHeader()
		hash    = head.Hash()
		number  = head.Number.Uint64()
		td      = pm.blockchain.GetTd(hash, number)
	)

	peerIsOnParentChain := p.GetP2PPeer().OnParentChain() // If the peer is on parent chain, this node is on child chain for the peer.
	onTheSameChain, err := p.Handshake(pm.networkId, pm.getChainID(), td, hash, genesis.Hash(), peerIsOnParentChain)
	if err != nil {
		p.GetP2PPeer().Log().Debug("klaytn peer handshake failed", "err", err)
		return err
	}
	if rw, ok := p.GetRW().(*meteredMsgReadWriter); ok {
		rw.Init(p.GetVersion())
	}

	if onTheSameChain {
		// Register the peer locally
		if err := pm.peers.Register(p); err != nil {
			// if starting node with unlock account, can't register peer until finish unlock
			p.GetP2PPeer().Log().Info("klaytn peer registration failed", "err", err)
			return err
		}
		defer pm.removePeer(p.GetID())

		// Register the peer in the downloader. If the downloader considers it banned, we disconnect
		if err := pm.downloader.RegisterPeer(p.GetID(), p.GetVersion(), p); err != nil {
			return err
		}
		// Propagate existing transactions. new transactions appearing
		// after this will be sent via broadcasts.
		pm.syncTransactions(p)
	} else {
		// Register the peer according to their role.
		if peerIsOnParentChain {
			if err := pm.scpm.getParentChainPeers().Register(p); err != nil {
				return err
			}
			defer pm.scpm.removeParentPeer(p.GetID())
		} else {
			if err := pm.scpm.getChildChainPeers().Register(p); err != nil {
				return err
			}
			defer pm.scpm.removeChildPeer(p.GetID())
		}
	}

	p.GetP2PPeer().Log().Info("Added a P2P Peer", "peerID", p.GetP2PPeerID(), "onTheSameChain", onTheSameChain, "onParentChain", peerIsOnParentChain)

	pubKey, err := p.GetP2PPeerID().Pubkey()
	if err != nil {
		return err
	}
	addr := crypto.PubkeyToAddress(*pubKey)

	// TODO-Klaytn check global worker and peer worker
	messageChannel := make(chan p2p.Msg, channelSizePerPeer)
	defer close(messageChannel)
	errChannel := make(chan error, channelSizePerPeer)
	for w := 1; w <= concurrentPerPeer; w++ {
		go pm.processMsg(messageChannel, p, addr, errChannel)
	}

	// main loop. handle incoming messages.
	for {
		msg, err := p.GetRW().ReadMsg()
		if err != nil {
			p.GetP2PPeer().Log().Debug("ProtocolManager failed to read msg", "err", err)
			return err
		}
		if msg.Size > ProtocolMaxMsgSize {
			err := errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
			p.GetP2PPeer().Log().Debug("ProtocolManager over max msg size", "err", err)
			return err
		}

		messageChannel <- msg

		select {
		case err := <-errChannel:
			return err
		default:
		}
		//go pm.handleMsg(p, addr, msg)

		//if err := pm.handleMsg(p); err != nil {
		//	p.Log().Debug("klaytn message handling failed", "err", err)
		//	return err
		//}
	}
}

func (pm *ProtocolManager) processMsg(msgCh <-chan p2p.Msg, p Peer, addr common.Address, errCh chan<- error) {
	for msg := range msgCh {
		if err := pm.handleMsg(p, addr, msg); err != nil {
			p.GetP2PPeer().Log().Debug("ProtocolManager failed to handle message", "msg", msg, "err", err)
			errCh <- err
			return
		}
		msg.Discard()
	}
	p.GetP2PPeer().Log().Debug("ProtocolManager.processMsg closed", "PeerName", p.GetP2PPeer().Name())
}

// handleMsg is invoked whenever an inbound message is received from a remote
// peer. The remote connection is torn down upon returning any error.
func (pm *ProtocolManager) handleMsg(p Peer, addr common.Address, msg p2p.Msg) error {
	// Below message size checking is done by handle().
	// Read the next message from the remote peer, and ensure it's fully consumed
	//msg, err := p.rw.ReadMsg()
	//if err != nil {
	//	return err
	//}
	//if msg.Size > ProtocolMaxMsgSize {
	//	return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	//}
	//defer msg.Discard()

	// istanbul BFT
	if handler, ok := pm.engine.(consensus.Handler); ok {
		//pubKey, err := p.ID().Pubkey()
		//if err != nil {
		//	return err
		//}
		//addr := crypto.PubkeyToAddress(*pubKey)
		handled, err := handler.HandleMsg(addr, msg)
		// if msg is istanbul msg, handled is true and err is nil if handle msg is successful.
		if handled {
			return err
		}
	}

	// Handle the message depending on its contents
	switch {
	case msg.Code == StatusMsg:
		// Status messages should never arrive after the handshake
		return errResp(ErrExtraStatusMsg, "uncontrolled status message")

		// Block header query, collect the requested headers and reply
	case msg.Code == GetBlockHeadersMsg:
		// Decode the complex header query
		var query getBlockHeadersData
		if err := msg.Decode(&query); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		hashMode := query.Origin.Hash != (common.Hash{})

		// Gather headers until the fetch or network limits is reached
		var (
			bytes   common.StorageSize
			headers []*types.Header
			unknown bool
		)
		for !unknown && len(headers) < int(query.Amount) && bytes < softResponseLimit && len(headers) < downloader.MaxHeaderFetch {
			// Retrieve the next header satisfying the query
			var origin *types.Header
			if hashMode {
				origin = pm.blockchain.GetHeaderByHash(query.Origin.Hash)
			} else {
				origin = pm.blockchain.GetHeaderByNumber(query.Origin.Number)
			}
			if origin == nil {
				break
			}
			number := origin.Number.Uint64()
			headers = append(headers, origin)
			bytes += estHeaderRlpSize

			// Advance to the next header of the query
			switch {
			case query.Origin.Hash != (common.Hash{}) && query.Reverse:
				// Hash based traversal towards the genesis block
				for i := 0; i < int(query.Skip)+1; i++ {
					if header := pm.blockchain.GetHeader(query.Origin.Hash, number); header != nil {
						query.Origin.Hash = header.ParentHash
						number--
					} else {
						unknown = true
						break
					}
				}
			case query.Origin.Hash != (common.Hash{}) && !query.Reverse:
				// Hash based traversal towards the leaf block
				var (
					current = origin.Number.Uint64()
					next    = current + query.Skip + 1
				)
				if next <= current {
					infos, _ := json.MarshalIndent(p.GetP2PPeer().Info(), "", "  ")
					p.GetP2PPeer().Log().Warn("GetBlockHeaders skip overflow attack", "current", current, "skip", query.Skip, "next", next, "attacker", infos)
					unknown = true
				} else {
					if header := pm.blockchain.GetHeaderByNumber(next); header != nil {
						if pm.blockchain.GetBlockHashesFromHash(header.Hash(), query.Skip+1)[query.Skip] == query.Origin.Hash {
							query.Origin.Hash = header.Hash()
						} else {
							unknown = true
						}
					} else {
						unknown = true
					}
				}
			case query.Reverse:
				// Number based traversal towards the genesis block
				if query.Origin.Number >= query.Skip+1 {
					query.Origin.Number -= query.Skip + 1
				} else {
					unknown = true
				}

			case !query.Reverse:
				// Number based traversal towards the leaf block
				query.Origin.Number += query.Skip + 1
			}
		}
		return p.SendBlockHeaders(headers)

	case msg.Code == BlockHeadersMsg:
		// A batch of headers arrived to one of our previous requests
		var headers []*types.Header
		if err := msg.Decode(&headers); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Filter out any explicitly requested headers, deliver the rest to the downloader
		filter := len(headers) == 1
		if filter {
			// Irrelevant of the fork checks, send the header to the fetcher just in case
			headers = pm.fetcher.FilterHeaders(p.GetID(), headers, time.Now())
		}
		if len(headers) > 0 || !filter {
			err := pm.downloader.DeliverHeaders(p.GetID(), headers)
			if err != nil {
				logger.Debug("Failed to deliver headers", "err", err)
			}
		}

	case msg.Code == GetBlockBodiesMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather blocks until the fetch or network limits is reached
		var (
			hash   common.Hash
			bytes  int
			bodies []rlp.RawValue
		)
		for bytes < softResponseLimit && len(bodies) < downloader.MaxBlockFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block body, stopping if enough was found
			if data := pm.blockchain.GetBodyRLP(hash); len(data) != 0 {
				bodies = append(bodies, data)
				bytes += len(data)
			}
		}
		return p.SendBlockBodiesRLP(bodies)

	case msg.Code == BlockBodiesMsg:
		// A batch of block bodies arrived to one of our previous requests
		var request blockBodiesData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver them all to the downloader for queuing
		transactions := make([][]*types.Transaction, len(request))
		uncles := make([][]*types.Header, len(request))

		for i, body := range request {
			transactions[i] = body.Transactions
			uncles[i] = body.Uncles
		}
		// Filter out any explicitly requested bodies, deliver the rest to the downloader
		filter := len(transactions) > 0 || len(uncles) > 0
		if filter {
			transactions, uncles = pm.fetcher.FilterBodies(p.GetID(), transactions, uncles, time.Now())
		}
		if len(transactions) > 0 || len(uncles) > 0 || !filter {
			err := pm.downloader.DeliverBodies(p.GetID(), transactions, uncles)
			if err != nil {
				logger.Debug("Failed to deliver bodies", "err", err)
			}
		}

	case p.GetVersion() >= klay63 && msg.Code == GetNodeDataMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash  common.Hash
			bytes int
			data  [][]byte
		)
		for bytes < softResponseLimit && len(data) < downloader.MaxStateFetch {
			// Retrieve the hash of the next state entry
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested state entry, stopping if enough was found
			if entry, err := pm.blockchain.TrieNode(hash); err == nil {
				data = append(data, entry)
				bytes += len(entry)
			}
		}
		return p.SendNodeData(data)

	case p.GetVersion() >= klay63 && msg.Code == NodeDataMsg:
		// A batch of node state data arrived to one of our previous requests
		var data [][]byte
		if err := msg.Decode(&data); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverNodeData(p.GetID(), data); err != nil {
			logger.Debug("Failed to deliver node state data", "err", err)
		}

	case p.GetVersion() >= klay63 && msg.Code == GetReceiptsMsg:
		// Decode the retrieval message
		msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
		if _, err := msgStream.List(); err != nil {
			return err
		}
		// Gather state data until the fetch or network limits is reached
		var (
			hash     common.Hash
			bytes    int
			receipts []rlp.RawValue
		)
		for bytes < softResponseLimit && len(receipts) < downloader.MaxReceiptFetch {
			// Retrieve the hash of the next block
			if err := msgStream.Decode(&hash); err == rlp.EOL {
				break
			} else if err != nil {
				return errResp(ErrDecode, "msg %v: %v", msg, err)
			}
			// Retrieve the requested block's receipts, skipping if unknown to us
			results := pm.blockchain.GetReceiptsByBlockHash(hash)
			if results == nil {
				if header := pm.blockchain.GetHeaderByHash(hash); header == nil || header.ReceiptHash != types.EmptyRootHash {
					continue
				}
			}
			// If known, encode and queue for response packet
			if encoded, err := rlp.EncodeToBytes(results); err != nil {
				logger.Error("Failed to encode receipt", "err", err)
			} else {
				receipts = append(receipts, encoded)
				bytes += len(encoded)
			}
		}
		return p.SendReceiptsRLP(receipts)

	case p.GetVersion() >= klay63 && msg.Code == ReceiptsMsg:
		// A batch of receipts arrived to one of our previous requests
		var receipts [][]*types.Receipt
		if err := msg.Decode(&receipts); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Deliver all to the downloader
		if err := pm.downloader.DeliverReceipts(p.GetID(), receipts); err != nil {
			logger.Debug("Failed to deliver receipts", "err", err)
		}

	case msg.Code == NewBlockHashesMsg:
		var announces newBlockHashesData
		if err := msg.Decode(&announces); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		// Mark the hashes as present at the remote node
		// Schedule all the unknown hashes for retrieval
		for _, block := range announces {
			p.AddToKnownBlocks(block.Hash)

			if !pm.blockchain.HasBlock(block.Hash, block.Number) {
				pm.fetcher.Notify(p.GetID(), block.Hash, block.Number, time.Now(), p.RequestOneHeader, p.RequestBodies)
			}
		}

	case msg.Code == NewBlockMsg:
		// Retrieve and decode the propagated block
		var request newBlockData
		if err := msg.Decode(&request); err != nil {
			return errResp(ErrDecode, "%v: %v", msg, err)
		}
		request.Block.ReceivedAt = msg.ReceivedAt
		request.Block.ReceivedFrom = p

		// Mark the peer as owning the block and schedule it for import
		p.AddToKnownBlocks(request.Block.Hash())
		pm.fetcher.Enqueue(p.GetID(), request.Block)

		// Assuming the block is importable by the peer, but possibly not yet done so,
		// calculate the head hash and TD that the peer truly must have.
		var (
			trueHead = request.Block.ParentHash()
			trueTD   = new(big.Int).Sub(request.TD, request.Block.Difficulty())
		)
		// Update the peers total difficulty if better than the previous
		if _, td := p.Head(); trueTD.Cmp(td) > 0 {
			p.SetHead(trueHead, trueTD)

			// Schedule a sync if above ours. Note, this will not fire a sync for a gap of
			// a singe block (as the true TD is below the propagated block), however this
			// scenario should easily be covered by the fetcher.
			currentBlock := pm.blockchain.CurrentBlock()
			if trueTD.Cmp(pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64())) > 0 {
				go pm.synchronise(p)
			}
		}

	case msg.Code == TxMsg:
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
			break
		}
		// Transactions can be processed, parse all of them and deliver to the pool
		var txs []*types.Transaction
		if err := msg.Decode(&txs); err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Only valid txs should be pushed into the pool.
		validTxs := make([]*types.Transaction, 0, len(txs))
		var err error
		for i, tx := range txs {
			// Validate and mark the remote transaction
			if tx == nil {
				err = errResp(ErrDecode, "transaction %d is nil", i)
				continue
			}
			p.AddToKnownTxs(tx.Hash())
			validTxs = append(validTxs, tx)
		}
		pm.txpool.AddRemotes(validTxs)
		return err

	// ServiceChain related messages
	case msg.Code == ServiceChainTxsMsg:
		scLogger.Debug("received ServiceChainTxsMsg")
		// Transactions arrived, make sure we have a valid and fresh chain to handle them
		if atomic.LoadUint32(&pm.acceptTxs) == 0 {
			break
		}
		if err := handleServiceChainTxDataMsg(pm, p, msg); err != nil {
			return err
		}

	case msg.Code == ServiceChainParentChainInfoRequestMsg:
		scLogger.Debug("received ServiceChainParentChainInfoRequestMsg")
		if err := handleServiceChainParentChainInfoRequestMsg(pm, p, msg); err != nil {
			return err
		}

	case msg.Code == ServiceChainParentChainInfoResponseMsg:
		scLogger.Debug("received ServiceChainParentChainInfoResponseMsg")
		if err := handleServiceChainParentChainInfoResponseMsg(pm, msg); err != nil {
			return err
		}

	case msg.Code == ServiceChainReceiptResponseMsg:
		scLogger.Debug("received ServiceChainReceiptResponseMsg")
		if err := handleServiceChainReceiptResponseMsg(pm, msg); err != nil {
			return err
		}

	case msg.Code == ServiceChainReceiptRequestMsg:
		scLogger.Debug("received ServiceChainReceiptRequestMsg")
		if err := handleServiceChainReceiptRequestMsg(pm, p, msg); err != nil {
			return err
		}

	default:
		return errResp(ErrInvalidMsgCode, "%v", msg.Code)
	}
	return nil
}

// handleServiceChainTxDataMsg handles service chain transactions from child chain.
// It will return an error if given tx is not TxTypeChainDataPegging type.
func handleServiceChainTxDataMsg(pm *ProtocolManager, p Peer, msg p2p.Msg) error {
	//pm.txMsgLock.Lock()
	// Transactions can be processed, parse all of them and deliver to the pool
	var txs []*types.Transaction
	if err := msg.Decode(&txs); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}

	// Only valid txs should be pushed into the pool.
	validTxs := make([]*types.Transaction, 0, len(txs))
	var err error
	for i, tx := range txs {
		if tx == nil {
			err = errResp(ErrDecode, "tx %d is nil", i)
			continue
		}
		if tx.Type() != types.TxTypeChainDataPegging {
			err = errResp(ErrUnexpectedTxType, "tx %d should be TxTypeChainDataPegging, but %s", i, tx.Type())
			continue
		}
		p.AddToKnownTxs(tx.Hash())
		validTxs = append(validTxs, tx)
	}
	pm.txpool.AddRemotes(validTxs)
	return err
}

// handleServiceChainParentChainInfoRequestMsg handles parent chain info request message from child chain.
// It will send the nonce of the account and its gas price to the child chain peer who requested.
func handleServiceChainParentChainInfoRequestMsg(pm *ProtocolManager, p Peer, msg p2p.Msg) error {
	var addr common.Address
	if err := msg.Decode(&addr); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	stateDB, err := pm.blockchain.State()
	if err != nil {
		// TODO-Klaytn-ServiceChain This error and inefficient balance error should be transferred.
		return errResp(ErrFailedToGetStateDB, "failed to get stateDB, err: %v", err)
	} else {
		pcInfo := parentChainInfo{stateDB.GetNonce(addr), pm.blockchain.Config().UnitPrice}
		p.SendServiceChainInfoResponse(&pcInfo)
		scLogger.Debug("SendServiceChainInfoResponse", "addr", addr, "nonce", pcInfo.Nonce, "gasPrice", pcInfo.GasPrice)
	}
	return nil
}

// handleServiceChainParentChainInfoResponseMsg handles parent chain info response message from parent chain.
// It will update the remoteNonce and remoteGasPrice of ServiceChainProtocolManager.
func handleServiceChainParentChainInfoResponseMsg(pm *ProtocolManager, msg p2p.Msg) error {
	var pcInfo parentChainInfo
	if err := msg.Decode(&pcInfo); err != nil {
		scLogger.Error("failed to decode", "err", err)
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	if pm.scpm.getRemoteNonce() > pcInfo.Nonce {
		// If received nonce is bigger than the current one, just leave a log and do nothing.
		scLogger.Warn("local nonce is bigger than the parent chain nonce.", "localNonce", pm.scpm.getRemoteNonce(), "remoteNonce", pcInfo.Nonce)
		return nil
	}
	pm.scpm.setRemoteNonce(pcInfo.Nonce)
	pm.scpm.setNonceSynced(true)
	pm.scpm.setRemoteGasPrice(pcInfo.GasPrice)
	scLogger.Debug("ServiceChainNonceResponse", "nonce", pcInfo.Nonce, "gasPrice", pcInfo.GasPrice)
	return nil
}

// handleServiceChainReceiptResponseMsg handles receipt response message from parent chain.
// It will store the received receipts and remove corresponding transaction in the resending list.
func handleServiceChainReceiptResponseMsg(pm *ProtocolManager, msg p2p.Msg) error {
	// TODO-Klaytn-ServiceChain Need to add an option, not to write receipts.
	// Decode the retrieval message
	var receipts []*types.ReceiptForStorage
	if err := msg.Decode(&receipts); err != nil && err != rlp.EOL {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	// Stores receipt and remove tx from sentServiceChainTxs only if the tx is successfully executed.
	pm.scpm.writeServiceChainTxReceipts(pm.blockchain, receipts)
	return nil
}

// handleServiceChainReceiptRequestMsg handles receipt request message from child chain.
// It will find and send corresponding receipts with given transaction hashes.
func handleServiceChainReceiptRequestMsg(pm *ProtocolManager, p Peer, msg p2p.Msg) error {
	// Decode the retrieval message
	msgStream := rlp.NewStream(msg.Payload, uint64(msg.Size))
	if _, err := msgStream.List(); err != nil {
		return err
	}
	// Gather state data until the fetch or network limits is reached
	var (
		hash               common.Hash
		receiptsForStorage []*types.ReceiptForStorage
	)
	for len(receiptsForStorage) < downloader.MaxReceiptFetch {
		// Retrieve the hash of the next block
		if err := msgStream.Decode(&hash); err == rlp.EOL {
			break
		} else if err != nil {
			return errResp(ErrDecode, "msg %v: %v", msg, err)
		}
		// Retrieve the receipt of requested service chain tx, skip if unknown.
		receipt := pm.blockchain.GetReceiptByTxHash(hash)
		if receipt == nil {
			continue
		}

		receiptsForStorage = append(receiptsForStorage, (*types.ReceiptForStorage)(receipt))
	}
	return p.SendServiceChainReceiptResponse(receiptsForStorage)
}

// BroadcastBlock will either propagate a block to a subset of it's peers, or
// will only announce it's availability (depending what's requested).
func (pm *ProtocolManager) BroadcastBlock(block *types.Block, propagate bool) {
	// If it is connected to a parent chain, broadcast service chain block.
	if propagate {
		pm.scpm.BroadcastServiceChainTxAndReceiptRequest(block)
	}

	hash := block.Hash()
	var peers []Peer
	if pm.nodetype == node.CONSENSUSNODE {
		peers = pm.peers.PeersWithoutBlock(hash)
	} else {
		peers = pm.peers.AnotherTypePeersWithoutBlock(hash, node.CONSENSUSNODE)
	}

	// If propagation is requested, send to a subset of the peer
	if propagate {
		// Calculate the TD of the block (it's not imported yet, so block.Td is not valid)
		var td *big.Int
		if parent := pm.blockchain.GetBlock(block.ParentHash(), block.NumberU64()-1); parent != nil {
			td = new(big.Int).Add(block.Difficulty(), pm.blockchain.GetTd(block.ParentHash(), block.NumberU64()-1))
		} else {
			logger.Error("Propagating dangling block", "number", block.Number(), "hash", hash)
			return
		}

		// TODO-Klaytn only send all validators + sub(peer) except subset for this block
		// Send the block to a subset of our peers
		//transfer := peers[:int(math.Sqrt(float64(len(peers))))]
		transfer := pm.subPeers(peers, int(math.Sqrt(float64(len(peers)))))
		for _, peer := range transfer {
			//peer.SendNewBlock(block, td)
			peer.AsyncSendNewBlock(block, td)
		}
		logger.Trace("Propagated block", "hash", hash, "recipients", len(transfer), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
		return
	}
	// Otherwise if the block is indeed in out own chain, announce it
	if pm.blockchain.HasBlock(hash, block.NumberU64()) {
		for _, peer := range peers {
			//peer.SendNewBlockHashes([]common.Hash{hash}, []uint64{block.NumberU64()})
			peer.AsyncSendNewBlockHash(block)
		}
		logger.Trace("Announced block", "hash", hash, "recipients", len(peers), "duration", common.PrettyDuration(time.Since(block.ReceivedAt)))
	}
}

// BroadcastTxs will propagate a batch of transactions to all peers which are not known to
// already have the given transaction.
func (pm *ProtocolManager) BroadcastTxs(txs types.Transactions) {
	// Broadcast transactions to a batch of peers not knowing about it
	switch pm.nodetype {
	case node.CONSENSUSNODE:
		pm.broadcastCNTx(txs)
	default:
		pm.broadcastNoCNTx(txs, false)
	}
}

func (pm *ProtocolManager) ReBroadcastTxs(txs types.Transactions) {
	if pm.nodetype != node.CONSENSUSNODE {
		pm.broadcastNoCNTx(txs, true)
	}
}

func (pm *ProtocolManager) broadcastCNTx(txs types.Transactions) {
	var txset = make(map[Peer]types.Transactions)
	for _, tx := range txs {
		peers := pm.peers.CNWithoutTx(tx.Hash())
		if len(peers) == 0 {
			logger.Trace("No peer to broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
			continue
		}

		// TODO-Klaytn Code Check
		//peers = peers[:int(math.Sqrt(float64(len(peers))))]
		half := (len(peers) / 2) + 2
		peers = pm.subPeers(peers, half)
		for _, peer := range peers {
			txset[peer] = append(txset[peer], tx)
		}
		logger.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
	}

	// FIXME include this again: peers = peers[:int(math.Sqrt(float64(len(peers))))]
	for peer, txs := range txset {
		//peer.SendTransactions(txs)
		peer.AsyncSendTransactions(txs)
	}
}

func (pm *ProtocolManager) broadcastNoCNTx(txs types.Transactions, resend bool) {
	var cntxset = make(map[Peer]types.Transactions)
	var txset = make(map[Peer]types.Transactions)
	for _, tx := range txs {
		// TODO-Klaytn drop or missing tx
		if resend {
			var peers []Peer
			if pm.nodetype == node.RANGERNODE || pm.nodetype == node.GENERALNODE {
				peers = pm.peers.TypePeers(node.BRIDGENODE)
			} else {
				peers = pm.peers.TypePeers(node.CONSENSUSNODE)
			}
			// TODO-Klaytn need to tuning pickSize. currently 3 is for availability and efficiency
			peers = pm.subPeers(peers, 3)
			for _, peer := range peers {
				txset[peer] = append(txset[peer], tx)
			}
		} else {
			peers := pm.peers.CNWithoutTx(tx.Hash())
			if len(peers) > 0 {
				// TODO-Klaytn optimize pickSize or propagation way
				peers = pm.subPeers(peers, 2)
				for _, peer := range peers {
					cntxset[peer] = append(cntxset[peer], tx)
				}
				logger.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
			}
			if pm.nodetype == node.RANGERNODE || pm.nodetype == node.GENERALNODE {
				peers = pm.peers.TypePeersWithoutTx(tx.Hash(), node.BRIDGENODE)
				for _, peer := range peers {
					txset[peer] = append(txset[peer], tx)
				}
			}
			peers = pm.peers.TypePeersWithoutTx(tx.Hash(), pm.nodetype)

			for _, peer := range peers {
				txset[peer] = append(txset[peer], tx)
			}
			logger.Trace("Broadcast transaction", "hash", tx.Hash(), "recipients", len(peers))
		}
	}

	if resend {
		for peer, txs := range txset {
			err := peer.ReSendTransactions(txs)
			if err != nil {
				logger.Error("peer.ReSendTransactions", "peer", peer.GetAddr(), "numTxs", len(txs))
			}
		}
	} else {
		for peer, txs := range cntxset {
			// TODO-Klaytn Handle network-failed txs
			//peer.AsyncSendTransactions(txs)
			err := peer.SendTransactions(txs)
			if err != nil {
				logger.Error("peer.SendTransactions", "peer", peer.GetAddr(), "numTxs", len(txs))
			}
		}
		for peer, txs := range txset {
			err := peer.SendTransactions(txs)
			if err != nil {
				logger.Error("peer.SendTransactions", "peer", peer.GetAddr(), "numTxs", len(txs))
			}
			//peer.AsyncSendTransactions(txs)
		}
	}
}

func (pm *ProtocolManager) subPeers(peers []Peer, pickSize int) []Peer {
	if len(peers) < pickSize {
		return peers
	}

	picker := rand.New(rand.NewSource(time.Now().Unix()))
	peerCount := len(peers)
	for i := 0; i < peerCount; i++ {
		randIndex := picker.Intn(peerCount)
		peers[i], peers[randIndex] = peers[randIndex], peers[i]
	}

	return peers[:pickSize]
}

// Mined broadcast loop
func (pm *ProtocolManager) minedBroadcastLoop() {
	// automatically stops if unsubscribe
	for obj := range pm.minedBlockSub.Chan() {
		switch ev := obj.Data.(type) {
		case blockchain.NewMinedBlockEvent:
			pm.BroadcastBlock(ev.Block, true)  // First propagate block to peers
			pm.BroadcastBlock(ev.Block, false) // Only then announce to the rest
		}
	}
}

func (pm *ProtocolManager) txBroadcastLoop() {
	for {
		select {
		case event := <-pm.txsCh:

			pm.BroadcastTxs(event.Txs)

			// Err() channel will be closed when unsubscribing.
		case <-pm.txsSub.Err():
			return
		}
	}
}

// NodeInfo represents a short summary of the Ethereum sub-protocol metadata
// known about the host peer.
type NodeInfo struct {
	Network    uint64              `json:"network"`    // Ethereum network ID (1=Frontier, 2=Morden, Ropsten=3, Rinkeby=4)
	Difficulty *big.Int            `json:"difficulty"` // Total difficulty of the host's blockchain
	Genesis    common.Hash         `json:"genesis"`    // SHA3 hash of the host's genesis block
	Config     *params.ChainConfig `json:"config"`     // Chain configuration for the fork rules
	Head       common.Hash         `json:"head"`       // SHA3 hash of the host's best owned block
}

// NodeInfo retrieves some protocol metadata about the running host node.
func (pm *ProtocolManager) NodeInfo() *NodeInfo {
	currentBlock := pm.blockchain.CurrentBlock()
	return &NodeInfo{
		Network:    pm.networkId,
		Difficulty: pm.blockchain.GetTd(currentBlock.Hash(), currentBlock.NumberU64()),
		Genesis:    pm.blockchain.Genesis().Hash(),
		Config:     pm.blockchain.Config(),
		Head:       currentBlock.Hash(),
	}
}

// istanbul BFT
func (pm *ProtocolManager) Enqueue(id string, block *types.Block) {
	pm.fetcher.Enqueue(id, block)
}

func (pm *ProtocolManager) FindPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	for _, p := range pm.peers.Peers() {
		addr := p.GetAddr()
		if addr == (common.Address{}) {
			pubKey, err := p.GetP2PPeerID().Pubkey()
			if err != nil {
				continue
			}
			addr = crypto.PubkeyToAddress(*pubKey)
		} else {
			addr = p.GetAddr()
		}

		if targets[addr] {
			m[addr] = p
		}
	}
	return m
}

func (pm *ProtocolManager) GetCNPeers() map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	for addr, p := range pm.peers.CNPeers() {
		m[addr] = p
	}
	return m
}

func (pm *ProtocolManager) FindCNPeers(targets map[common.Address]bool) map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	for addr, p := range pm.peers.CNPeers() {
		if targets[addr] {
			m[addr] = p
		}
	}
	return m
}

func (pm *ProtocolManager) GetRNPeers() map[common.Address]consensus.Peer {
	m := make(map[common.Address]consensus.Peer)
	for addr, p := range pm.peers.RNPeers() {
		m[addr] = p
	}
	return m
}

func (pm *ProtocolManager) GetPeers() []common.Address {
	addrs := make([]common.Address, 0)
	for _, p := range pm.peers.Peers() {
		addr := p.GetAddr()
		if addr == (common.Address{}) {
			pubKey, err := p.GetP2PPeerID().Pubkey()
			if err != nil {
				continue
			}
			addr = crypto.PubkeyToAddress(*pubKey)
		} else {
			addr = p.GetAddr()
		}
		addrs = append(addrs, addr)
	}
	return addrs
}

// GetChainAddr returns an address of an account used for service chain in string format.
// If given as a parameter, it will use it. If not given, it will use the address of the public key
// derived from chainKey.
func (pm *ProtocolManager) GetChainAddr() string {
	return pm.scpm.GetChainAddr().String()
}

// GetChainTxPeriod returns the period (in child chain blocks) of sending service chain transaction
// from child chain to parent chain.
func (pm *ProtocolManager) GetChainTxPeriod() uint64 {
	return pm.scpm.GetChainTxPeriod()
}

// GetSentChainTxsLimit returns the maximum number of stored  chain transactions
// in child chain node, which is for resending.
func (pm *ProtocolManager) GetSentChainTxsLimit() uint64 {
	return pm.scpm.GetSentChainTxsLimit()
}
