// Copyright 2018 The go-klaytn Authors
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
	"context"
	"fmt"
	"github.com/ground-x/go-gxplatform/accounts"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/blockchain/bloombits"
	"github.com/ground-x/go-gxplatform/blockchain/state"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/blockchain/vm"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/common/math"
	"github.com/ground-x/go-gxplatform/datasync/downloader"
	"github.com/ground-x/go-gxplatform/event"
	"github.com/ground-x/go-gxplatform/networks/rpc"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/ground-x/go-gxplatform/storage/database"
	"github.com/pkg/errors"
	"math/big"
	"time"
)

// RangerAPIBackend implements gxpapi.Backend for ranger nodes
type RangerAPIBackend struct {
	ranger *Ranger
}

func (b *RangerAPIBackend) GetTransactionInCache(hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64) {
	return nil, common.Hash{}, 0, 0
}

func (b *RangerAPIBackend) GetReceiptsInCache(blockHash common.Hash) (types.Receipts, error) {
	return nil, errors.New("doesn't support getreceiptsIncache")
}

func (b *RangerAPIBackend) ChainConfig() *params.ChainConfig {
	return b.ranger.chainConfig
}

func (b *RangerAPIBackend) CurrentBlock() *types.Block {
	return b.ranger.blockchain.CurrentBlock()
}

func (b *RangerAPIBackend) SetHead(number uint64) {
	//b.cn.protocolManager.downloader.Cancel()
	b.ranger.blockchain.SetHead(number)
}

func (b *RangerAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ranger.blockchain.CurrentBlock().Header(), nil
	}
	header := b.ranger.blockchain.GetHeaderByNumber(uint64(blockNr))
	if header == nil {
		return nil, fmt.Errorf("the block does not exist (block number: %d)", blockNr)
	}
	return header, nil
}

func (b *RangerAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.ranger.blockchain.CurrentBlock(), nil
	}
	block := b.ranger.blockchain.GetBlockByNumber(uint64(blockNr))
	if block == nil {
		return nil, fmt.Errorf("the block does not exist (block number: %d)", blockNr)
	}
	return block, nil
}

func (b *RangerAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.ranger.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *RangerAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	block := b.ranger.blockchain.GetBlockByHash(hash)
	if block == nil {
		return nil, fmt.Errorf("the block does not exist (block number: %s)", hash.String())
	}
	return block, nil
}

func (b *RangerAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.ranger.blockchain.GetReceiptsByBlockHash(hash), nil
}

func (b *RangerAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	return b.ranger.blockchain.GetLogsByHash(hash), nil
}

func (b *RangerAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.ranger.blockchain.GetTdByHash(blockHash)
}

func (b *RangerAPIBackend) GetEVM(ctx context.Context, msg blockchain.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := blockchain.NewEVMContext(msg, header, b.ranger.BlockChain(), nil)
	return vm.NewEVM(context, state, b.ranger.chainConfig, &vmCfg), vmError, nil
}

func (b *RangerAPIBackend) SubscribeRemovedLogsEvent(ch chan<- blockchain.RemovedLogsEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *RangerAPIBackend) SubscribeChainEvent(ch chan<- blockchain.ChainEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeChainEvent(ch)
}

func (b *RangerAPIBackend) SubscribeChainHeadEvent(ch chan<- blockchain.ChainHeadEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *RangerAPIBackend) SubscribeChainSideEvent(ch chan<- blockchain.ChainSideEvent) event.Subscription {
	return b.ranger.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *RangerAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.ranger.BlockChain().SubscribeLogsEvent(ch)
}

func (b *RangerAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return nil
}

func (b *RangerAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.ranger.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *RangerAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.ranger.txPool.Get(hash)
}

func (b *RangerAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) uint64 {
	return b.ranger.txPool.State().GetNonce(addr)
}

func (b *RangerAPIBackend) Stats() (pending int, queued int) {
	return b.ranger.txPool.Stats()
}

func (b *RangerAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.ranger.TxPool().Content()
}

func (b *RangerAPIBackend) SubscribeNewTxsEvent(ch chan<- blockchain.NewTxsEvent) event.Subscription {
	return b.ranger.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *RangerAPIBackend) Downloader() *downloader.Downloader {
	return b.ranger.Downloader()
}

func (b *RangerAPIBackend) ProtocolVersion() int {
	return b.ranger.GxpVersion()
}

func (b *RangerAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) { // TODO-GX-issue136 gasPrice
	return common.Big0, nil
}

func (b *RangerAPIBackend) ChainDB() database.DBManager {
	return b.ranger.ChainDB()
}

func (b *RangerAPIBackend) EventMux() *event.TypeMux {
	return b.ranger.EventMux()
}

func (b *RangerAPIBackend) AccountManager() *accounts.Manager {
	return b.ranger.AccountManager()
}

func (b *RangerAPIBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.ranger.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

const (
	// bloomFilterThreads is the number of goroutines used locally per filter to
	// multiplex requests onto the global servicing goroutines.
	bloomFilterThreads = 3

	// bloomRetrievalBatch is the maximum number of bloom bit retrievals to service
	// in a single batch.
	bloomRetrievalBatch = 16

	// bloomRetrievalWait is the maximum time to wait for enough bloom bit requests
	// to accumulate request an entire batch (avoiding hysteresis).
	bloomRetrievalWait = time.Duration(0)
)

func (b *RangerAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.ranger.bloomRequests)
	}
}
