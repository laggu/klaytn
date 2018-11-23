package work

import (
	"fmt"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/consensus"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/blockchain/state"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/event"
	"github.com/ground-x/go-gxplatform/storage/database"
	"github.com/ground-x/go-gxplatform/datasync/downloader"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/params"
	"sync/atomic"
	"github.com/ground-x/go-gxplatform/networks/p2p"
)

var logger = log.NewModuleLogger(log.Work)

// Backend wraps all methods required for mining.
type Backend interface {
	AccountManager() *accounts.Manager
	BlockChain() 	 *blockchain.BlockChain
	TxPool() 	     *blockchain.TxPool
	ChainDB() 		 database.DBManager
	ReBroadcastTxs(transactions types.Transactions)
}

// Miner creates blocks and searches for proof-of-work values.
type Miner struct {
	mux *event.TypeMux

	worker *worker

	coinbase common.Address
	mining   int32
	gxp      Backend
	engine   consensus.Engine

	canStart    int32 // can start indicates whether we can start the mining operation
	shouldStart int32 // should start indicates whether we should start after sync
}

func New(gxp Backend, config *params.ChainConfig, mux *event.TypeMux, engine consensus.Engine, nodetype p2p.ConnType) *Miner {
	miner := &Miner{
		gxp:      gxp,
		mux:      mux,
		engine:   engine,
		worker:   newWorker(config, engine, common.Address{}, gxp, mux, nodetype),
		canStart: 1,
	}
	// TODO-KLAYTN drop or missing tx
	miner.Register(NewCpuAgent(gxp.BlockChain(), engine, nodetype))
	go miner.update()

	return miner
}

// update keeps track of the downloader events. Please be aware that this is a one shot type of update loop.
// It's entered once and as soon as `Done` or `Failed` has been broadcasted the events are unregistered and
// the loop is exited. This to prevent a major security vuln where external parties can DOS you with blocks
// and halt your mining operation for as long as the DOS continues.
func (self *Miner) update() {
	events := self.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
out:
	for ev := range events.Chan() {
		switch ev.Data.(type) {
		case downloader.StartEvent:
			atomic.StoreInt32(&self.canStart, 0)
			if self.Mining() {
				self.Stop()
				atomic.StoreInt32(&self.shouldStart, 1)
				logger.Info("Mining aborted due to sync")
			}
		case downloader.DoneEvent, downloader.FailedEvent:
			shouldStart := atomic.LoadInt32(&self.shouldStart) == 1

			atomic.StoreInt32(&self.canStart, 1)
			atomic.StoreInt32(&self.shouldStart, 0)
			if shouldStart {
				self.Start(self.coinbase)
			}
			// unsubscribe. we're only interested in this event once
			events.Unsubscribe()
			// stop immediately and ignore all further pending events
			break out
		}
	}
}

func (self *Miner) Start(coinbase common.Address) {
	atomic.StoreInt32(&self.shouldStart, 1)
	self.SetCoinbase(coinbase)

	if atomic.LoadInt32(&self.canStart) == 0 {
		logger.Info("Network syncing, will start work afterwards")
		return
	}
	atomic.StoreInt32(&self.mining, 1)

	logger.Info("Starting mining operation")
	self.worker.start()
	self.worker.commitNewWork()
}

func (self *Miner) Stop() {
	self.worker.stop()
	atomic.StoreInt32(&self.mining, 0)
	atomic.StoreInt32(&self.shouldStart, 0)
}

func (self *Miner) Register(agent Agent) {
	if self.Mining() {
		agent.Start()
	}
	self.worker.register(agent)
}

func (self *Miner) Unregister(agent Agent) {
	self.worker.unregister(agent)
}

func (self *Miner) Mining() bool {
	return atomic.LoadInt32(&self.mining) > 0
}

func (self *Miner) HashRate() (tot int64) {
	if pow, ok := self.engine.(consensus.PoW); ok {
		tot += int64(pow.Hashrate())
	}
	// do we care this might race? is it worth we're rewriting some
	// aspects of the worker/locking up agents so we can get an accurate
	// hashrate?
	for agent := range self.worker.agents {
		if _, ok := agent.(*CpuAgent); !ok {
			tot += agent.GetHashRate()
		}
	}
	return
}

func (self *Miner) SetExtra(extra []byte) error {
	// istanbul BFT
	if uint64(len(extra)) > params.GetMaximumExtraDataSize(self.worker.chain.Config().IsBFT) {
		return fmt.Errorf("Extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize)
	}
	self.worker.setExtra(extra)
	return nil
}

// Pending returns the currently pending block and associated state.
func (self *Miner) Pending() (*types.Block, *state.StateDB) {
	return self.worker.pending()
}

// PendingBlock returns the currently pending block.
//
// Note, to access both the pending block and the pending state
// simultaneously, please use Pending(), as the pending state can
// change between multiple method calls
func (self *Miner) PendingBlock() *types.Block {
	return self.worker.pendingBlock()
}

func (self *Miner) SetCoinbase(addr common.Address) {
	self.coinbase = addr
	self.worker.setCoinbase(addr)
}
