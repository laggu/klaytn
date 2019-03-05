// Copyright 2018 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package sc

import (
	"context"
	"errors"
	"fmt"
	"github.com/ground-x/klaytn"
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/blockchain/bloombits"
	"github.com/ground-x/klaytn/blockchain/state"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/blockchain/vm"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/common/math"
	"github.com/ground-x/klaytn/event"
	"github.com/ground-x/klaytn/networks/rpc"
	"github.com/ground-x/klaytn/node/cn/filters"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/storage/database"
	"math/big"
	"sync"
	"time"
)

const defaultGasPrice = 50 * params.Ston

var errBlockNumberUnsupported = errors.New("LocalBackend cannot access blocks other than the latest block")
var errGasEstimationFailed = errors.New("gas required exceeds allowance or always failing transaction")

// TODO-Klaytn currently LocalBackend is only for ServiceChain, especially Gateway SmartContract
type LocalBackend struct {
	subbrige *SubBridge

	events *filters.EventSystem // Event system for filtering log events live
	config *params.ChainConfig

	mu         sync.Mutex
	logfilters []*LogFilter
}

func NewLocalBackend(main *SubBridge) (*LocalBackend, error) {
	lb := &LocalBackend{
		subbrige: main,
		config:   main.blockchain.Config(),
		events:   filters.NewEventSystem(main.EventMux(), &filterLocalBackend{main}, false),
	}
	// add localbackend to evenhandler to handle log as LogEventListener
	main.eventhandler.AddListener(lb)
	return lb, nil
}

func (lb *LocalBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	if blockNumber != nil && blockNumber.Cmp(lb.subbrige.blockchain.CurrentBlock().Number()) != 0 {
		return nil, errBlockNumberUnsupported
	}
	statedb, _ := lb.subbrige.blockchain.State()
	return statedb.GetCode(contract), nil
}

func (lb *LocalBackend) CallContract(ctx context.Context, call klaytn.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if blockNumber != nil && blockNumber.Cmp(lb.subbrige.blockchain.CurrentBlock().Number()) != 0 {
		return nil, errBlockNumberUnsupported
	}
	state, err := lb.subbrige.blockchain.State()
	if err != nil {
		return nil, err
	}
	rval, _, _, err := lb.callContract(ctx, call, lb.subbrige.blockchain.CurrentBlock(), state)
	return rval, err
}

func (b *LocalBackend) callContract(ctx context.Context, call klaytn.CallMsg, block *types.Block, statedb *state.StateDB) ([]byte, uint64, bool, error) {
	// TODO-Klaytn Set sender address or use a default if none specified
	if call.From == (common.Address{}) {
		return nil, 0, false, errors.New("from address is not set !!")
	}
	// Set default gas & gas price if none were set
	gas, gasPrice := uint64(call.Gas), call.GasPrice // TODO-Klaytn-Issue136 gasPrice
	if gas == 0 {
		gas = math.MaxUint64 / 2
	}
	if gasPrice == nil || gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice) // TODO-Klaytn-Issue136 default gasPrice
	}

	intrinsicGas, err := types.IntrinsicGas(call.Data, call.To == nil, true)
	if err != nil {
		return nil, 0, false, err
	}

	// Create new call message
	msg := types.NewMessage(call.From, call.To, 0, call.Value, gas, gasPrice, call.Data, false, intrinsicGas)

	// Setup context so it may be cancelled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	statedb.SetBalance(msg.ValidatedSender(), math.MaxBig256)
	vmError := func() error { return nil }

	context := blockchain.NewEVMContext(msg, block.Header(), b.subbrige.blockchain, nil)
	evm := vm.NewEVM(context, statedb, b.config, &vm.Config{})
	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		evm.Cancel(vm.CancelByCtxDone)
	}()

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	gp := new(blockchain.GasPool).AddGas(math.MaxUint64) // TODO-Klaytn-Issue136
	res, gas, kerr := blockchain.ApplyMessage(evm, msg, gp)
	err = kerr.Err
	if err := vmError(); err != nil {
		return nil, 0, false, err
	}

	// Propagate error of Receipt as JSON RPC error
	if err == nil {
		err = blockchain.GetVMerrFromReceiptStatus(kerr.Status)
	}

	return res, gas, kerr.Status != types.ReceiptStatusSuccessful, err
}

func (lb *LocalBackend) PendingCodeAt(ctx context.Context, contract common.Address) ([]byte, error) {
	// TODO-Klaytn this is not pending code but latest code
	return lb.CodeAt(ctx, contract, lb.subbrige.blockchain.CurrentBlock().Number())
}

func (lb *LocalBackend) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return lb.subbrige.txPool.State().GetNonce(account), nil
}

func (lb *LocalBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	// TODO-Klaytn consider whether SuggestGasPrice is necessary or not
	return big.NewInt(1), nil
}

func (lb *LocalBackend) EstimateGas(ctx context.Context, call klaytn.CallMsg) (gas uint64, err error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		cap uint64
	)
	if uint64(call.Gas) >= params.TxGas {
		hi = uint64(call.Gas)
	} else {
		// Retrieve the current pending block to act as the gas ceiling
		// TODO-Klaytn consider whether using pendingBlock or not
		block := lb.subbrige.blockchain.CurrentBlock()
		hi = block.GasLimit()
	}
	cap = hi

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) bool {
		call.Gas = gas

		state, err := lb.subbrige.blockchain.State()
		_, _, failed, err := lb.callContract(ctx, call, lb.subbrige.blockchain.CurrentBlock(), state)
		if err != nil || failed {
			return false
		}
		return true
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		if !executable(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		if !executable(hi) {
			return 0, fmt.Errorf("gas required exceeds allowance or always failing transaction")
		}
	}
	return hi, nil
}

func (lb *LocalBackend) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return lb.subbrige.txPool.AddLocal(tx)
}

func (lb *LocalBackend) FilterLogs(ctx context.Context, query klaytn.FilterQuery) ([]types.Log, error) {
	// Convert the RPC block numbers into internal representations
	if query.FromBlock == nil {
		query.FromBlock = big.NewInt(rpc.LatestBlockNumber.Int64())
	}
	if query.ToBlock == nil {
		query.ToBlock = big.NewInt(rpc.LatestBlockNumber.Int64())
	}
	from := query.FromBlock.Int64()
	to := query.ToBlock.Int64()

	// Construct and execute the filter
	filter := filters.New(&filterLocalBackend{lb.subbrige}, from, to, query.Addresses, query.Topics)

	logs, err := filter.Logs(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]types.Log, len(logs))
	for i, log := range logs {
		res[i] = *log
	}
	return res, nil
}

func (lb *LocalBackend) SubscribeFilterLogs(ctx context.Context, query klaytn.FilterQuery, ch chan<- types.Log) (klaytn.Subscription, error) {
	lb.mu.Lock()
	// TODO-Klaytn improve logfilter management
	lb.logfilters = append(lb.logfilters, &LogFilter{query, ch})
	idx := len(lb.logfilters) - 1
	lb.mu.Unlock()
	return &FilterLogSubscription{idx: idx, localBackend: lb, err: make(chan error, 1)}, nil
}

func (lb *LocalBackend) Handle(logs []*types.Log) error {

	for _, filter := range lb.logfilters {
		addresses := filter.query.Addresses
		topics := filter.query.Topics
	Logs:
		for _, log := range logs {
			if len(addresses) > 0 && !includes(addresses, log.Address) {
				continue
			}
			// If the to filtered topics is greater than the amount of topics in logs, skip.
			if len(topics) > len(log.Topics) {
				continue Logs
			}
			for i, topics := range topics {
				match := len(topics) == 0 // empty rule set == wildcard
				for _, topic := range topics {
					if log.Topics[i] == topic {
						match = true
						break
					}
				}
				if !match {
					continue Logs
				}
			}
			filter.logCh <- *log
		}
	}
	return nil
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}
	return false
}

func (lb *LocalBackend) RemoveFilter(idx int) {
	lb.mu.Lock()
	if len(lb.logfilters) > idx+1 {
		lb.logfilters = append(lb.logfilters[:idx], lb.logfilters[idx+1:]...)
	} else {
		lb.logfilters = lb.logfilters[:idx]
	}
	lb.mu.Unlock()
}

// LogFilter has query,logCh to filter log
type LogFilter struct {
	query klaytn.FilterQuery
	logCh chan<- types.Log
}

// FilterLogSubscription is Subscription in SubscribeFilterLogs
type FilterLogSubscription struct {
	idx          int
	localBackend *LocalBackend
	mu           sync.Mutex
	err          chan error
	unsubscribed bool
}

func (fls *FilterLogSubscription) Unsubscribe() {
	fls.mu.Lock()
	if !fls.unsubscribed {
		fls.localBackend.RemoveFilter(fls.idx)
		fls.err <- nil
		close(fls.err)
	}
	fls.unsubscribed = true
	fls.mu.Unlock()
}

func (fls *FilterLogSubscription) Err() <-chan error {
	return fls.err
}

type filterLocalBackend struct {
	subbridge *SubBridge
}

func (fb *filterLocalBackend) ChainDB() database.DBManager {
	// TODO-Klaytn consider chain's chainDB instead of bridge's chainDB currently.
	logger.Error("use ChainDB in filterLocalBackend ")
	return fb.subbridge.chainDB
}
func (fb *filterLocalBackend) EventMux() *event.TypeMux {
	// TODO-Klaytn consider chain's eventMux instead of bridge's eventMux currently.
	logger.Error("use EventMux in filterLocalBackend ")
	return fb.subbridge.EventMux()
}

func (fb *filterLocalBackend) HeaderByNumber(ctx context.Context, block rpc.BlockNumber) (*types.Header, error) {
	// TODO-Klaytn consider pendingblock instead of latest block
	if block == rpc.LatestBlockNumber {
		return fb.subbridge.blockchain.CurrentHeader(), nil
	}
	return fb.subbridge.blockchain.GetHeaderByNumber(uint64(block.Int64())), nil
}

func (fb *filterLocalBackend) GetBlockReceipts(ctx context.Context, hash common.Hash) types.Receipts {
	return fb.subbridge.blockchain.GetReceiptsByBlockHash(hash)
}

func (fb *filterLocalBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	return fb.subbridge.blockchain.GetLogsByHash(hash), nil
}

func (fb *filterLocalBackend) SubscribeNewTxsEvent(ch chan<- blockchain.NewTxsEvent) event.Subscription {
	return fb.subbridge.txPool.SubscribeNewTxsEvent(ch)
}

func (fb *filterLocalBackend) SubscribeChainEvent(ch chan<- blockchain.ChainEvent) event.Subscription {
	return fb.subbridge.blockchain.SubscribeChainEvent(ch)
}

func (fb *filterLocalBackend) SubscribeRemovedLogsEvent(ch chan<- blockchain.RemovedLogsEvent) event.Subscription {
	return fb.subbridge.blockchain.SubscribeRemovedLogsEvent(ch)
}

func (fb *filterLocalBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return fb.subbridge.blockchain.SubscribeLogsEvent(ch)
}

func (fb *filterLocalBackend) BloomStatus() (uint64, uint64) {
	// TODO-Klaytn consider this number of sections.
	// BloomBitsBlocks (const : 4096), the number of processed sections maintained by the chain indexer
	return 4096, 0
}

func (fb *filterLocalBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	// TODO-Klaytn this method should implmentation to support indexed tag in solidity
	//for i := 0; i < bloomFilterThreads; i++ {
	//	go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, backend.bloomRequests)
	//}
}
