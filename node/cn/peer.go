// Copyright 2018 The go-klaytn Authors
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
// This file is derived from eth/peer.go (2018/06/04).
// Modified and improved for the go-klaytn development.

package cn

import (
	"errors"
	"fmt"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/datasync/downloader"
	"github.com/ground-x/go-gxplatform/networks/p2p"
	"github.com/ground-x/go-gxplatform/networks/p2p/discover"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"gopkg.in/fatih/set.v0"
	"math/big"
	"sync"
	"time"

	"github.com/ground-x/go-gxplatform/node"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

const (
	maxKnownTxs    = 32768 // Maximum transactions hashes to keep in the known list (prevent DOS)
	maxKnownBlocks = 1024  // Maximum block hashes to keep in the known list (prevent DOS)

	// maxQueuedTxs is the maximum number of transaction lists to queue up before
	// dropping broadcasts. This is a sensitive number as a transaction list might
	// contain a single transaction, or thousands.
	maxQueuedTxs = 128

	// maxQueuedProps is the maximum number of block propagations to queue up before
	// dropping broadcasts. There's not much point in queueing stale blocks, so a few
	// that might cover uncles should be enough.
	maxQueuedProps = 4

	// maxQueuedAnns is the maximum number of block announcements to queue up before
	// dropping broadcasts. Similarly to block propagations, there's no point to queue
	// above some healthy uncle limit, so use that.
	maxQueuedAnns = 4

	handshakeTimeout = 5 * time.Second
)

// PeerInfo represents a short summary of the Ethereum sub-protocol metadata known
// about a connected peer.
type PeerInfo struct {
	Version    int      `json:"version"`    // Ethereum protocol version negotiated
	Difficulty *big.Int `json:"difficulty"` // Total difficulty of the peer's blockchain
	Head       string   `json:"head"`       // SHA3 hash of the peer's best owned block
}

// propEvent is a block propagation, waiting for its turn in the broadcast queue.
type propEvent struct {
	block *types.Block
	td    *big.Int
}

type Peer interface {
	// broadcast is a write loop that multiplexes block propagations, announcements
	// and transaction broadcasts into the remote peer. The goal is to have an async
	// writer that does not lock up node internals.
	Broadcast()

	// close signals the broadcast goroutine to terminate.
	Close()

	// Info gathers and returns a collection of metadata known about a peer.
	Info() *PeerInfo

	// SetHead updates the head hash and total difficulty of the peer.
	SetHead(hash common.Hash, td *big.Int)

	// AddToKnownBlocks adds a block to knownBlocks for the peer, ensuring that the block will
	// never be propagated to this particular peer.
	AddToKnownBlocks(hash common.Hash)

	// AddToKnownTxs adds a transaction to knownTxs for the peer, ensuring that it
	// will never be propagated to this particular peer.
	AddToKnownTxs(hash common.Hash)

	// istanbul BFT
	// Send writes an RLP-encoded message with the given code.
	// data should encode as an RLP list.
	Send(msgcode uint64, data interface{}) error

	// SendTransactions sends transactions to the peer and includes the hashes
	// in its transaction hash set for future reference.
	SendTransactions(txs types.Transactions) error

	// ReSendTransaction sends txs to a peer in order to prevent the txs from missing.
	ReSendTransactions(txs types.Transactions) error

	// AsyncSendTransactions sends transactions asynchronously to the peer
	AsyncSendTransactions(txs []*types.Transaction)

	// SendNewBlockHashes announces the availability of a number of blocks through
	// a hash notification.
	SendNewBlockHashes(hashes []common.Hash, numbers []uint64) error

	// AsyncSendNewBlockHash queues the availability of a block for propagation to a
	// remote peer. If the peer's broadcast queue is full, the event is silently
	// dropped.
	AsyncSendNewBlockHash(block *types.Block)

	// SendNewBlock propagates an entire block to a remote peer.
	SendNewBlock(block *types.Block, td *big.Int) error

	// AsyncSendNewBlock queues an entire block for propagation to a remote peer. If
	// the peer's broadcast queue is full, the event is silently dropped.
	AsyncSendNewBlock(block *types.Block, td *big.Int)

	// SendBlockHeaders sends a batch of block headers to the remote peer.
	SendBlockHeaders(headers []*types.Header) error

	// SendBlockBodies sends a batch of block contents to the remote peer.
	SendBlockBodies(bodies []*blockBody) error

	// SendBlockBodiesRLP sends a batch of block contents to the remote peer from
	// an already RLP encoded format.
	SendBlockBodiesRLP(bodies []rlp.RawValue) error

	// SendNodeDataRLP sends a batch of arbitrary internal data, corresponding to the
	// hashes requested.
	SendNodeData(data [][]byte) error

	// SendReceiptsRLP sends a batch of transaction receipts, corresponding to the
	// ones requested from an already RLP encoded format.
	SendReceiptsRLP(receipts []rlp.RawValue) error

	// RequestOneHeader is a wrapper around the header query functions to fetch a
	// single header. It is used solely by the fetcher.
	RequestOneHeader(hash common.Hash) error

	// Handshake executes the eth protocol handshake, negotiating version number,
	// network IDs, difficulties, head and genesis blocks.
	Handshake(network uint64, td *big.Int, head common.Hash, genesis common.Hash) error

	// ConnType returns the conntype of the peer.
	ConnType() p2p.ConnType

	// GetID returns the id of the peer.
	GetID() string

	// GetP2PPeerID returns the id of the p2p.Peer.
	GetP2PPeerID() discover.NodeID

	// GetAddr returns the address of the peer.
	GetAddr() common.Address

	// SetAddr sets the address of the peer.
	SetAddr(addr common.Address)

	// GetVersion returns the version of the peer.
	GetVersion() int

	// GetKnownBlocks returns the knownBlocks of the peer.
	GetKnownBlocks() *set.Set

	// GetKnownTxs returns the knownBlocks of the peer.
	GetKnownTxs() *set.Set

	// GetP2PPeer returns the p2p.
	GetP2PPeer() *p2p.Peer

	// GetRW returns the MsgReadWriter of the peer.
	GetRW() p2p.MsgReadWriter

	// Peer encapsulates the methods required to synchronise with a remote full peer.
	downloader.Peer
}

// basePeer is a common data structure used by implementation of Peer.
type basePeer struct {
	id string

	addr common.Address

	*p2p.Peer
	rw p2p.MsgReadWriter

	version  int         // Protocol version negotiated
	forkDrop *time.Timer // Timed connection dropper if forks aren't validated in time

	head common.Hash
	td   *big.Int
	lock sync.RWMutex

	knownTxs    *set.Set                  // Set of transaction hashes known to be known by this peer
	knownBlocks *set.Set                  // Set of block hashes known to be known by this peer
	queuedTxs   chan []*types.Transaction // Queue of transactions to broadcast to the peer
	queuedProps chan *propEvent           // Queue of blocks to broadcast to the peer
	queuedAnns  chan *types.Block         // Queue of blocks to announce to the peer
	term        chan struct{}             // Termination channel to stop the broadcaster
}

// NewPeer returns new Peer interface
func newPeer(version int, p *p2p.Peer, rw p2p.MsgReadWriter) Peer {
	id := p.ID()

	return &singleChannelPeer{
		basePeer: basePeer{
			Peer:        p,
			rw:          rw,
			version:     version,
			id:          fmt.Sprintf("%x", id[:8]),
			knownTxs:    set.New(),
			knownBlocks: set.New(),
			queuedTxs:   make(chan []*types.Transaction, maxQueuedTxs),
			queuedProps: make(chan *propEvent, maxQueuedProps),
			queuedAnns:  make(chan *types.Block, maxQueuedAnns),
			term:        make(chan struct{}),
		},
	}
}

// broadcast is a write loop that multiplexes block propagations, announcements
// and transaction broadcasts into the remote peer. The goal is to have an async
// writer that does not lock up node internals.
func (p *basePeer) Broadcast() {
	for {
		select {
		case txs := <-p.queuedTxs:
			if err := p.SendTransactions(txs); err != nil {
				logger.Error("fail to SendTransactions", "err", err)
				continue
				//return
			}
			p.Log().Trace("Broadcast transactions", "count", len(txs))

		case prop := <-p.queuedProps:
			if err := p.SendNewBlock(prop.block, prop.td); err != nil {
				logger.Error("fail to SendNewBlock", "err", err)
				continue
				//return
			}
			p.Log().Trace("Propagated block", "number", prop.block.Number(), "hash", prop.block.Hash(), "td", prop.td)

		case block := <-p.queuedAnns:
			if err := p.SendNewBlockHashes([]common.Hash{block.Hash()}, []uint64{block.NumberU64()}); err != nil {
				logger.Error("fail to SendNewBlockHashes", "err", err)
				continue
				//return
			}
			p.Log().Trace("Announced block", "number", block.Number(), "hash", block.Hash())

		case <-p.term:
			p.Log().Debug("Peer broadcast loop end")
			return
		}
	}
}

// close signals the broadcast goroutine to terminate.
func (p *basePeer) Close() {
	close(p.term)
}

// Info gathers and returns a collection of metadata known about a peer.
func (p *basePeer) Info() *PeerInfo {
	hash, td := p.Head()

	return &PeerInfo{
		Version:    p.version,
		Difficulty: td,
		Head:       hash.Hex(),
	}
}

// Head retrieves a copy of the current head hash and total difficulty of the
// peer.
func (p *basePeer) Head() (hash common.Hash, td *big.Int) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	copy(hash[:], p.head[:])
	return hash, new(big.Int).Set(p.td)
}

// SetHead updates the head hash and total difficulty of the peer.
func (p *basePeer) SetHead(hash common.Hash, td *big.Int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	copy(p.head[:], hash[:])
	p.td.Set(td)
}

// AddToKnownBlocks adds a block to knownBlocks for the peer, ensuring that the block will
// never be propagated to this particular peer.
func (p *basePeer) AddToKnownBlocks(hash common.Hash) {
	if !p.knownBlocks.Has(hash) {
		// If we reached the memory allowance, drop a previously known block hash
		for p.knownBlocks.Size() >= maxKnownBlocks {
			p.knownBlocks.Pop()
		}
		p.knownBlocks.Add(hash)
	}
}

// AddToKnownTxs adds a transaction to knownTxs for the peer, ensuring that it
// will never be propagated to this particular peer.
func (p *basePeer) AddToKnownTxs(hash common.Hash) {
	if !p.knownTxs.Has(hash) {
		// If we reached the memory allowance, drop a previously known transaction hash
		for p.knownTxs.Size() >= maxKnownTxs {
			p.knownTxs.Pop()
		}
		p.knownTxs.Add(hash)
	}
}

// istanbul BFT
// Send writes an RLP-encoded message with the given code.
// data should encode as an RLP list.
func (p *basePeer) Send(msgcode uint64, data interface{}) error {
	return p2p.Send(p.rw, msgcode, data)
}

// SendTransactions sends transactions to the peer and includes the hashes
// in its transaction hash set for future reference.
func (p *basePeer) SendTransactions(txs types.Transactions) error {
	for _, tx := range txs {
		p.AddToKnownTxs(tx.Hash())
	}
	return p2p.Send(p.rw, TxMsg, txs)
}

// ReSendTransaction sends txs to a peer in order to prevent the txs from missing.
func (p *basePeer) ReSendTransactions(txs types.Transactions) error {
	return p2p.Send(p.rw, TxMsg, txs)
}

func (p *basePeer) AsyncSendTransactions(txs []*types.Transaction) {
	select {
	case p.queuedTxs <- txs:
		for _, tx := range txs {
			p.AddToKnownTxs(tx.Hash())
		}
	default:
		p.Log().Debug("Dropping transaction propagation", "count", len(txs))
	}
}

// SendNewBlockHashes announces the availability of a number of blocks through
// a hash notification.
func (p *basePeer) SendNewBlockHashes(hashes []common.Hash, numbers []uint64) error {
	for _, hash := range hashes {
		p.knownBlocks.Add(hash)
	}
	request := make(newBlockHashesData, len(hashes))
	for i := 0; i < len(hashes); i++ {
		request[i].Hash = hashes[i]
		request[i].Number = numbers[i]
	}
	return p2p.Send(p.rw, NewBlockHashesMsg, request)
}

// AsyncSendNewBlockHash queues the availability of a block for propagation to a
// remote peer. If the peer's broadcast queue is full, the event is silently
// dropped.
func (p *basePeer) AsyncSendNewBlockHash(block *types.Block) {
	select {
	case p.queuedAnns <- block:
		p.knownBlocks.Add(block.Hash())
	default:
		p.Log().Debug("Dropping block announcement", "number", block.NumberU64(), "hash", block.Hash())
	}
}

// SendNewBlock propagates an entire block to a remote peer.
func (p *basePeer) SendNewBlock(block *types.Block, td *big.Int) error {
	p.knownBlocks.Add(block.Hash())
	return p2p.Send(p.rw, NewBlockMsg, []interface{}{block, td})
}

// AsyncSendNewBlock queues an entire block for propagation to a remote peer. If
// the peer's broadcast queue is full, the event is silently dropped.
func (p *basePeer) AsyncSendNewBlock(block *types.Block, td *big.Int) {
	select {
	case p.queuedProps <- &propEvent{block: block, td: td}:
		p.knownBlocks.Add(block.Hash())
	default:
		p.Log().Debug("Dropping block propagation", "number", block.NumberU64(), "hash", block.Hash())
	}
}

// SendBlockHeaders sends a batch of block headers to the remote peer.
func (p *basePeer) SendBlockHeaders(headers []*types.Header) error {
	return p2p.Send(p.rw, BlockHeadersMsg, headers)
}

// SendBlockBodies sends a batch of block contents to the remote peer.
func (p *basePeer) SendBlockBodies(bodies []*blockBody) error {
	return p2p.Send(p.rw, BlockBodiesMsg, blockBodiesData(bodies))
}

// SendBlockBodiesRLP sends a batch of block contents to the remote peer from
// an already RLP encoded format.
func (p *basePeer) SendBlockBodiesRLP(bodies []rlp.RawValue) error {
	return p2p.Send(p.rw, BlockBodiesMsg, bodies)
}

// SendNodeDataRLP sends a batch of arbitrary internal data, corresponding to the
// hashes requested.
func (p *basePeer) SendNodeData(data [][]byte) error {
	return p2p.Send(p.rw, NodeDataMsg, data)
}

// SendReceiptsRLP sends a batch of transaction receipts, corresponding to the
// ones requested from an already RLP encoded format.
func (p *basePeer) SendReceiptsRLP(receipts []rlp.RawValue) error {
	return p2p.Send(p.rw, ReceiptsMsg, receipts)
}

// RequestOneHeader is a wrapper around the header query functions to fetch a
// single header. It is used solely by the fetcher.
func (p *basePeer) RequestOneHeader(hash common.Hash) error {
	p.Log().Debug("Fetching single header", "hash", hash)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Hash: hash}, Amount: uint64(1), Skip: uint64(0), Reverse: false})
}

// RequestHeadersByHash fetches a batch of blocks' headers corresponding to the
// specified header query, based on the hash of an origin block.
func (p *basePeer) RequestHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	p.Log().Debug("Fetching batch of headers", "count", amount, "fromhash", origin, "skip", skip, "reverse", reverse)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Hash: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

// RequestHeadersByNumber fetches a batch of blocks' headers corresponding to the
// specified header query, based on the number of an origin block.
func (p *basePeer) RequestHeadersByNumber(origin uint64, amount int, skip int, reverse bool) error {
	p.Log().Debug("Fetching batch of headers", "count", amount, "fromnum", origin, "skip", skip, "reverse", reverse)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Number: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

// RequestBodies fetches a batch of blocks' bodies corresponding to the hashes
// specified.
func (p *basePeer) RequestBodies(hashes []common.Hash) error {
	p.Log().Debug("Fetching batch of block bodies", "count", len(hashes))
	return p2p.Send(p.rw, GetBlockBodiesMsg, hashes)
}

// RequestNodeData fetches a batch of arbitrary data from a node's known state
// data, corresponding to the specified hashes.
func (p *basePeer) RequestNodeData(hashes []common.Hash) error {
	p.Log().Debug("Fetching batch of state data", "count", len(hashes))
	return p2p.Send(p.rw, GetNodeDataMsg, hashes)
}

// RequestReceipts fetches a batch of transaction receipts from a remote node.
func (p *basePeer) RequestReceipts(hashes []common.Hash) error {
	p.Log().Debug("Fetching batch of receipts", "count", len(hashes))
	return p2p.Send(p.rw, GetReceiptsMsg, hashes)
}

// Handshake executes the eth protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis blocks.
func (p *basePeer) Handshake(network uint64, td *big.Int, head common.Hash, genesis common.Hash) error {
	// Send out own handshake in a new thread
	errc := make(chan error, 2)
	var status statusData // safe to read after two values have been received from errc

	go func() {
		errc <- p2p.Send(p.rw, StatusMsg, &statusData{
			ProtocolVersion: uint32(p.version),
			NetworkId:       network,
			TD:              td,
			CurrentBlock:    head,
			GenesisBlock:    genesis,
		})
	}()
	go func() {
		errc <- p.readStatus(network, &status, genesis)
	}()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}
	p.td, p.head = status.TD, status.CurrentBlock
	return nil
}

func (p *basePeer) readStatus(network uint64, status *statusData, genesis common.Hash) (err error) {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != StatusMsg {
		return errResp(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, StatusMsg)
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	// Decode the handshake and make sure everything matches
	if err := msg.Decode(&status); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	if status.GenesisBlock != genesis {
		return errResp(ErrGenesisBlockMismatch, "%x (!= %x)", status.GenesisBlock[:8], genesis[:8])
	}
	if status.NetworkId != network {
		return errResp(ErrNetworkIdMismatch, "%d (!= %d)", status.NetworkId, network)
	}
	if int(status.ProtocolVersion) != p.version {
		return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", status.ProtocolVersion, p.version)
	}
	return nil
}

// String implements fmt.Stringer.
func (p *basePeer) String() string {
	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("klay/%2d", p.version),
	)
}

// ConnType returns the conntype of the peer.
func (p *basePeer) ConnType() p2p.ConnType {
	return p.Peer.ConnType()
}

// GetID returns the id of the peer.
func (p *basePeer) GetID() string {
	return p.id
}

// GetP2PPeerID returns the id of the p2p.Peer.
func (p *basePeer) GetP2PPeerID() discover.NodeID {
	return p.Peer.ID()
}

// GetAddr returns the address of the peer.
func (p *basePeer) GetAddr() common.Address {
	return p.addr
}

// SetAddr sets the address of the peer.
func (p *basePeer) SetAddr(addr common.Address) {
	p.addr = addr
}

// GetVersion returns the version of the peer.
func (p *basePeer) GetVersion() int {
	return p.version
}

// GetKnownBlocks returns the knownBlocks of the peer.
func (p *basePeer) GetKnownBlocks() *set.Set {
	return p.knownBlocks
}

// GetKnownTxs returns the knownBlocks of the peer.
func (p *basePeer) GetKnownTxs() *set.Set {
	return p.knownTxs
}

// GetP2PPeer returns the p2p.Peer.
func (p *basePeer) GetP2PPeer() *p2p.Peer {
	return p.Peer
}

// GetRW returns the MsgReadWriter of the peer.
func (p *basePeer) GetRW() p2p.MsgReadWriter {
	return p.rw
}

// singleChannelPeer is a peer that uses a single channel.
type singleChannelPeer struct {
	basePeer
}

type ByPassValidator struct{}

func (v ByPassValidator) ValidatePeerType(addr common.Address) error {
	return nil
}

// peerSet represents the collection of active peers currently participating in
// the Klaytn sub-protocol.
type peerSet struct {
	peers   map[string]Peer
	cnpeers map[common.Address]Peer
	rnpeers map[common.Address]Peer
	lock    sync.RWMutex
	closed  bool

	validator map[p2p.ConnType]p2p.PeerTypeValidator
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	peerSet := &peerSet{
		peers:     make(map[string]Peer),
		cnpeers:   make(map[common.Address]Peer),
		rnpeers:   make(map[common.Address]Peer),
		validator: make(map[p2p.ConnType]p2p.PeerTypeValidator),
	}

	peerSet.validator[node.CONSENSUSNODE] = ByPassValidator{}
	peerSet.validator[node.RANGERNODE] = ByPassValidator{}
	peerSet.validator[node.GENERALNODE] = ByPassValidator{}
	peerSet.validator[node.BRIDGENODE] = ByPassValidator{}

	return peerSet
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
func (ps *peerSet) Register(p Peer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return errClosed
	}
	if _, ok := ps.peers[p.GetID()]; ok {
		return errAlreadyRegistered
	}
	if p.ConnType() == node.CONSENSUSNODE {
		if _, ok := ps.cnpeers[p.GetAddr()]; ok {
			return errAlreadyRegistered
		}
		if err := ps.validator[node.CONSENSUSNODE].ValidatePeerType(p.GetAddr()); err != nil {
			return fmt.Errorf("fail to validate cntype: %s", err)
		}
		ps.cnpeers[p.GetAddr()] = p
	} else if p.ConnType() == node.RANGERNODE {
		if _, ok := ps.rnpeers[p.GetAddr()]; ok {
			return errAlreadyRegistered
		}
		if err := ps.validator[node.RANGERNODE].ValidatePeerType(p.GetAddr()); err != nil {
			return fmt.Errorf("fail to validate rntype: %s", err)
		}
		ps.rnpeers[p.GetAddr()] = p
	}
	ps.peers[p.GetID()] = p
	go p.Broadcast()

	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	p, ok := ps.peers[id]
	if !ok {
		return errNotRegistered
	}
	if p.ConnType() == node.CONSENSUSNODE {
		delete(ps.cnpeers, p.GetAddr())
	} else if p.ConnType() == node.RANGERNODE {
		delete(ps.rnpeers, p.GetAddr())
	}
	delete(ps.peers, id)
	p.Close()

	return nil
}

// istanbul BFT
func (ps *peerSet) Peers() map[string]Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	set := make(map[string]Peer)
	for id, p := range ps.peers {
		set[id] = p
	}
	return set
}

func (ps *peerSet) CNPeers() map[common.Address]Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	set := make(map[common.Address]Peer)
	for addr, p := range ps.cnpeers {
		set[addr] = p
	}
	return set
}

func (ps *peerSet) RNPeers() map[common.Address]Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	set := make(map[common.Address]Peer)
	for addr, p := range ps.rnpeers {
		set[addr] = p
	}
	return set
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id string) Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

// Len returns if the current number of peers in the set.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// PeersWithoutBlock retrieves a list of peers that do not have a given block in
// their set of known hashes.
func (ps *peerSet) PeersWithoutBlock(hash common.Hash) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.GetKnownBlocks().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) TypePeersWithoutBlock(hash common.Hash, nodetype p2p.ConnType) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.ConnType() == nodetype && !p.GetKnownBlocks().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) AnotherTypePeersWithoutBlock(hash common.Hash, nodetype p2p.ConnType) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.ConnType() != nodetype && !p.GetKnownBlocks().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) CNWithoutBlock(hash common.Hash) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.cnpeers))
	for _, p := range ps.cnpeers {
		if !p.GetKnownBlocks().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) TypePeers(nodetype p2p.ConnType) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.ConnType() == nodetype {
			list = append(list, p)
		}
	}
	return list
}

// PeersWithoutTx retrieves a list of peers that do not have a given transaction
// in their set of known hashes.
func (ps *peerSet) PeersWithoutTx(hash common.Hash) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.GetKnownTxs().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) TypePeersWithoutTx(hash common.Hash, nodetype p2p.ConnType) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.ConnType() == nodetype && !p.GetKnownTxs().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) TypePeersWithTx(hash common.Hash, nodetype p2p.ConnType) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.ConnType() == nodetype && p.GetKnownTxs().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) AnotherTypePeersWithoutTx(hash common.Hash, nodetype p2p.ConnType) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.ConnType() != nodetype && !p.GetKnownTxs().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

// TODO-KLAYTN drop or missing tx
func (ps *peerSet) AnotherTypePeersWithTx(hash common.Hash, nodetype p2p.ConnType) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if p.ConnType() != nodetype && p.GetKnownTxs().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

func (ps *peerSet) CNWithoutTx(hash common.Hash) []Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]Peer, 0, len(ps.cnpeers))
	for _, p := range ps.cnpeers {
		if !p.GetKnownTxs().Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

// BestPeer retrieves the known peer with the currently highest total difficulty.
func (ps *peerSet) BestPeer() Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	var (
		bestPeer Peer
		bestTd   *big.Int
	)
	for _, p := range ps.peers {
		if _, td := p.Head(); bestPeer == nil || td.Cmp(bestTd) > 0 {
			bestPeer, bestTd = p, td
		}
	}
	return bestPeer
}

// Close disconnects all peers.
// No new peers can be registered after Close has returned.
func (ps *peerSet) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, p := range ps.peers {
		p.GetP2PPeer().Disconnect(p2p.DiscQuitting)
	}
	ps.closed = true
}
