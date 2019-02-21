// Copyright 2018 The klaytn Authors
// Copyright 2014 The go-ethereum Authors
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
//
// This file is derived from core/tx_pool.go (2018/06/04).
// Modified and improved for the klaytn development.

package blockchain

import (
	"errors"
	"fmt"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"math/big"
	"sync"
	"time"
)

var (
	ErrKnownTx   = errors.New("Known Transaction")
	ErrUnknownTx = errors.New("Unknown Transaction")
)

// TODO-Klaytn-Servicechain Add Metrics
//var (
//	// Metrics for the pending pool
//	pendingDiscardCounter   = metrics.NewRegisteredCounter("txpool/pending/discard", nil)
//	pendingReplaceCounter   = metrics.NewRegisteredCounter("txpool/pending/replace", nil)
//	pendingRateLimitCounter = metrics.NewRegisteredCounter("txpool/pending/ratelimit", nil) // Dropped due to rate limiting
//	pendingNofundsCounter   = metrics.NewRegisteredCounter("txpool/pending/nofunds", nil)   // Dropped due to out-of-funds
//
//	// Metrics for the queued pool
//	queuedDiscardCounter   = metrics.NewRegisteredCounter("txpool/queued/discard", nil)
//	queuedReplaceCounter   = metrics.NewRegisteredCounter("txpool/queued/replace", nil)
//	queuedRateLimitCounter = metrics.NewRegisteredCounter("txpool/queued/ratelimit", nil) // Dropped due to rate limiting
//	queuedNofundsCounter   = metrics.NewRegisteredCounter("txpool/queued/nofunds", nil)   // Dropped due to out-of-funds
//
//	// General tx metrics
//	invalidTxCounter     = metrics.NewRegisteredCounter("txpool/invalid", nil)
//	underpricedTxCounter = metrics.NewRegisteredCounter("txpool/underpriced", nil)
//	refusedTxCounter     = metrics.NewRegisteredCounter("txpool/refuse", nil)
//)

// ChainTxPoolConfig are the configuration parameters of the transaction pool.
type ChainTxPoolConfig struct {
	Journal   string        // Journal of local transactions to survive node restarts
	Rejournal time.Duration // Time interval to regenerate the local transaction journal

	AccountQueue uint64 // Maximum number of non-executable transaction slots permitted per account
	GlobalQueue  uint64 // Maximum number of non-executable transaction slots for all accounts

	Lifetime time.Duration // Maximum amount of time non-executable transaction are queued
}

// DefaultChainTxPoolConfig contains the default configurations for the transaction
// pool.
var DefaultChainTxPoolConfig = ChainTxPoolConfig{
	Journal:   "chain_transactions.rlp",
	Rejournal: time.Hour,

	AccountQueue: 64,
	GlobalQueue:  1024,

	Lifetime: 3 * time.Hour,
}

// sanitize checks the provided user configurations and changes anything that's
// unreasonable or unworkable.
func (config *ChainTxPoolConfig) sanitize() ChainTxPoolConfig {
	conf := *config
	if conf.Rejournal < time.Second {
		logger.Error("Sanitizing invalid chaintxpool journal time", "provided", conf.Rejournal, "updated", time.Second)
		conf.Rejournal = time.Second
	}

	if conf.Journal == "" {
		logger.Error("Sanitizing invalid chaintxpool journal file name", "updated", DefaultChainTxPoolConfig.Journal)
		conf.Journal = DefaultChainTxPoolConfig.Journal
	}

	return conf
}

// ChainTxPool contains all currently known chain transactions.
type ChainTxPool struct {
	config ChainTxPoolConfig
	chain  blockChain
	// TODO-Klaytn-Servicechain consider to remove singer. For now, caused of value transfer tx which don't have `from` value, I leave it.
	signer types.Signer
	mu     sync.RWMutex

	locals  *chainAccountSet // Set of local transaction to exempt from eviction rules
	journal *txJournal       // Journal of local transaction to back up to disk

	txMu sync.RWMutex

	queue map[common.Address]*txSortedMap // Queued but non-processable transactions

	// TODO-Klaytn-Servicechain refine heartbeat for the tx not for account.
	//beats map[common.Address]time.Time       // Last heartbeat from each known account
	all map[common.Hash]*types.Transaction // All transactions to allow lookups

	wg sync.WaitGroup // for shutdown sync

	closed chan struct{}
}

// NewChainTxPool creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func NewChainTxPool(config ChainTxPoolConfig, chain blockChain) *ChainTxPool {
	// Sanitize the input to ensure no vulnerable gas prices are set
	config = (&config).sanitize()

	// Create the transaction pool with its initial settings
	pool := &ChainTxPool{
		config: config,
		chain:  chain,
		queue:  make(map[common.Address]*txSortedMap),
		//beats:  make(map[common.Address]time.Time),
		all:    make(map[common.Hash]*types.Transaction),
		closed: make(chan struct{}),
	}

	pool.locals = newChainAccountSet()

	// load from disk
	pool.journal = newTxJournal(config.Journal)

	if err := pool.journal.load(pool.AddLocals); err != nil {
		logger.Error("Failed to load chain transaction journal", "err", err)
	}
	if err := pool.journal.rotate(pool.Pending()); err != nil {
		logger.Error("Failed to rotate chain transaction journal", "err", err)
	}

	// Start the event loop and return
	pool.wg.Add(1)
	go pool.loop()

	return pool
}

// SetEIP155Signer set signer of txpool.
func (pool *ChainTxPool) SetEIP155Signer(chainID *big.Int) {
	pool.signer = types.NewEIP155Signer(chainID)
}

// loop is the transaction pool's main event loop, waiting for and reacting to
// outside blockchain events as well as for various reporting and transaction
// eviction events.
func (pool *ChainTxPool) loop() {
	defer pool.wg.Done()

	evict := time.NewTicker(evictionInterval)
	defer evict.Stop()

	journal := time.NewTicker(pool.config.Rejournal)
	defer journal.Stop()

	// Keep waiting for and reacting to the various events
	for {
		select {
		// Handle inactive account transaction eviction
		case <-evict.C:
			pool.mu.Lock()
			for addr := range pool.queue {
				// Skip local transactions from the eviction mechanism
				if pool.locals.contains(addr) {
					continue
				}
				// TODO-Klaytn-Servicechain refine heartbeat for the tx not for account.
				//// Any non-locals old enough should be removed
				//if time.Since(pool.beats[addr]) > pool.config.Lifetime {
				//	for _, tx := range pool.queue[addr].Flatten() {
				//		pool.removeTx(tx.Hash())
				//	}
				//}
			}
			pool.mu.Unlock()

			// Handle local transaction journal rotation
		case <-journal.C:
			if pool.journal != nil {
				pool.mu.Lock()
				if err := pool.journal.rotate(pool.Pending()); err != nil {
					logger.Error("Failed to rotate local tx journal", "err", err)
				}
				pool.mu.Unlock()
			}
		case <-pool.closed:
			logger.Error("ChainTxPool loop is closing")
			return
		}
	}
}

// Stop terminates the transaction pool.
func (pool *ChainTxPool) Stop() {
	close(pool.closed)
	pool.wg.Wait()

	if pool.journal != nil {
		pool.journal.close()
	}
	logger.Info("Transaction pool stopped")
}

// stats retrieves the current pool stats, namely the number of pending transactions.
func (pool *ChainTxPool) stats() int {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	queued := 0
	for _, list := range pool.queue {
		queued += list.Len()
	}
	return queued
}

// Content retrieves the data content of the transaction pool, returning all the
// queued transactions, grouped by account and sorted by nonce.
func (pool *ChainTxPool) Content() map[common.Address]types.Transactions {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	queued := make(map[common.Address]types.Transactions)
	for addr, list := range pool.queue {
		queued[addr] = list.Flatten()
	}
	return queued
}

// GetTx get the tx by tx hash.
func (pool *ChainTxPool) GetTx(txHash common.Hash) (*types.Transaction, error) {
	tx, ok := pool.all[txHash]

	if ok {
		return tx, nil
	} else {
		return nil, ErrUnknownTx
	}
}

// Pending retrieves all currently known local transactions, grouped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *ChainTxPool) Pending() map[common.Address]types.Transactions {
	pool.txMu.Lock()
	defer pool.txMu.Unlock()

	pending := make(map[common.Address]types.Transactions)
	for addr, list := range pool.queue {
		pending[addr] = list.Flatten()
	}
	return pending
}

// PendingTxsByAddress retrieves pending transactions, grouped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
func (pool *ChainTxPool) PendingTxsByAddress(from *common.Address, limit int) types.Transactions {
	pool.txMu.Lock()
	defer pool.txMu.Unlock()

	var pendingTxs types.Transactions

	if list, exist := pool.queue[*from]; exist {
		pendingTxs = list.Flatten()

		if len(pendingTxs) > limit {
			return pendingTxs[0:limit]
		}
		return pendingTxs
	}
	return nil
}

// GetMaxTxNonce finds max nonce of the address.
func (pool *ChainTxPool) GetMaxTxNonce(from *common.Address) (uint64, error) {
	pool.txMu.Lock()
	defer pool.txMu.Unlock()

	maxNonce := uint64(0)

	if list, exist := pool.queue[*from]; exist {
		for _, t := range list.items {
			if maxNonce < t.Nonce() {
				maxNonce = t.Nonce()
			}
		}
		return maxNonce, nil
	}

	return 0, errors.New("Non-exist from address")
}

// add validates a transaction and inserts it into the non-executable queue for
// later pending promotion and execution. If the transaction is a replacement for
// an already pending or queued one, it overwrites the previous and returns this
// so outer code doesn't uselessly call promote.
func (pool *ChainTxPool) add(tx *types.Transaction) error {
	// If the transaction is already known, discard it
	hash := tx.Hash()
	if pool.all[hash] != nil {
		logger.Trace("Discarding already known transaction", "hash", hash)
		return ErrKnownTx
	}

	// TODO-Klaytn-Servicechain consider to remove singer. For now, caused of value transfer tx which don't have `from` value, I leave it.
	from, err := tx.From() //from, _ := types.Sender(pool.signer, tx)
	if err != nil {
		return err
	}

	if uint64(len(pool.all)) >= pool.config.GlobalQueue {
		logger.Trace("Rejecting a new Tx, because ChainTxPool is full and there is no room for the account", "hash", tx.Hash(), "account", from)
		refusedTxCounter.Inc(1)
		return fmt.Errorf("txpool is full: %d", uint64(len(pool.all)))
	}

	if pool.queue[from] == nil {
		pool.queue[from] = newTxSortedMap()
	}

	pool.queue[from].Put(tx)

	if pool.all[hash] == nil {
		pool.all[hash] = tx
	}

	// Mark local addresses and journal local transactions
	pool.locals.add(from)
	pool.journalTx(from, tx)

	logger.Trace("Pooled new future transaction", "hash", hash, "from", from, "to", tx.To())
	return nil
}

// journalTx adds the specified transaction to the local disk journal if it is
// deemed to have been sent from a local account.
func (pool *ChainTxPool) journalTx(from common.Address, tx *types.Transaction) {
	// Only journal if it's enabled and the transaction is local
	if pool.journal == nil || !pool.locals.contains(from) {
		return
	}
	if err := pool.journal.insert(tx); err != nil {
		logger.Error("Failed to journal local transaction", "err", err)
	}
}

// AddLocal enqueues a single transaction into the pool if it is valid, marking
// the sender as a local one.
func (pool *ChainTxPool) AddLocal(tx *types.Transaction) error {
	return pool.addTx(tx)
}

// AddLocals enqueues a batch of transactions into the pool if they are valid,
// marking the senders as a local ones.
func (pool *ChainTxPool) AddLocals(txs []*types.Transaction) []error {
	return pool.addTxs(txs)
}

// addTx enqueues a single transaction into the pool if it is valid.
func (pool *ChainTxPool) addTx(tx *types.Transaction) error {
	//senderCacher.recover(pool.signer, []*types.Transaction{tx})

	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Try to inject the transaction and update any state
	err := pool.add(tx)

	return err
}

// addTxs attempts to queue a batch of transactions if they are valid.
func (pool *ChainTxPool) addTxs(txs []*types.Transaction) []error {
	//senderCacher.recover(pool.signer, txs)

	pool.mu.Lock()
	defer pool.mu.Unlock()

	return pool.addTxsLocked(txs)
}

// addTxsLocked attempts to queue a batch of transactions if they are valid,
// whilst assuming the transaction pool lock is already held.
func (pool *ChainTxPool) addTxsLocked(txs []*types.Transaction) []error {
	// Add the batch of transaction, tracking the accepted ones
	dirty := make(map[common.Address]struct{})
	errs := make([]error, len(txs))

	for i, tx := range txs {
		var replace bool
		if errs[i] = pool.add(tx); errs[i] == nil {
			if !replace {
				// TODO-Klaytn-Servicechain consider to remove singer. For now, caused of value transfer tx which don't have `from` value, I leave it.
				from, err := tx.From() //from, _ := types.Sender(pool.signer, tx) // already validated
				errs[i] = err

				dirty[from] = struct{}{}
			}
		}
	}

	return errs
}

// Get returns a transaction if it is contained in the pool
// and nil otherwise.
func (pool *ChainTxPool) Get(hash common.Hash) *types.Transaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	return pool.all[hash]
}

// removeTx removes a single transaction from the queue.
func (pool *ChainTxPool) removeTx(hash common.Hash) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Fetch the transaction we wish to delete
	tx, ok := pool.all[hash]
	if !ok {
		return ErrUnknownTx
	}

	// TODO-Klaytn-Servicechain consider to remove singer. For now, caused of value transfer tx which don't have `from` value, I leave it.
	addr, err := tx.From() //addr, _ := types.Sender(pool.signer, tx) // already validated during insertion
	if err != nil {
		return err
	}

	// Remove it from the list of known transactions
	delete(pool.all, hash)

	// Transaction is in the future queue
	if future := pool.queue[addr]; future != nil {
		future.Remove(tx.Nonce())
		if future.Len() == 0 {
			delete(pool.queue, addr)
		}
	}

	return nil
}

// Remove removes transactions from the queue.
func (pool *ChainTxPool) Remove(txs types.Transactions) []error {
	errs := make([]error, len(txs))
	for i, tx := range txs {
		errs[i] = pool.removeTx(tx.Hash())
	}
	return errs
}

// RemoveTx removes a single transaction from the queue.
func (pool *ChainTxPool) RemoveTx(tx *types.Transaction) error {
	err := pool.removeTx(tx.Hash())
	if err != nil {
		logger.Error("RemoveTx", "err", err)
		return err
	}
	return nil
}

// chainAccountSet is simply a set of addresses to check for existence.
type chainAccountSet struct {
	accounts map[common.Address]struct{}
}

// newChainAccountSet creates a new address set with an associated signer for sender
// derivations.
func newChainAccountSet() *chainAccountSet {
	return &chainAccountSet{
		accounts: make(map[common.Address]struct{}),
	}
}

// contains checks if a given address is contained within the set.
func (as *chainAccountSet) contains(addr common.Address) bool {
	_, exist := as.accounts[addr]
	return exist
}

// add inserts a new address into the set to track.
func (as *chainAccountSet) add(addr common.Address) {
	as.accounts[addr] = struct{}{}
}
