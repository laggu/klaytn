// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
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
// This file is derived from internal/ethapi/api.go (2018/06/04).
// Modified and improved for the klaytn development.

package api

import (
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
)

// PublicServiceChainAPI provides an API to access the Klaytn service chain related information.
type PublicServiceChainAPI struct {
	b Backend
}

// NewPublicServiceChainAPI creates a new Klaytn service chain API.
func NewPublicServiceChainAPI(b Backend) *PublicServiceChainAPI {
	return &PublicServiceChainAPI{b}
}

// GetChildChainIndexingEnabled returns the current child chain indexing configuration.
func (s *PublicServiceChainAPI) GetChildChainIndexingEnabled() bool {
	return s.b.GetChildChainIndexingEnabled()
}

// ConvertChildChainBlockHashToParentChainTxHash returns a transaction hash of a transaction which contains
// ChainHashes, with the key made with given child chain block hash.
func (s *PublicServiceChainAPI) ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash common.Hash) common.Hash {
	return s.b.ConvertChildChainBlockHashToParentChainTxHash(ccBlockHash)
}

// GetLatestAnchoredBlockNumber returns the latest block number whose data has been anchored to the parent chain.
func (s *PublicServiceChainAPI) GetLatestAnchoredBlockNumber() uint64 {
	return s.b.GetLatestAnchoredBlockNumber()
}

// GetReceiptFromParentChain returns saved receipt received from parent chain.
// This receipt is for the transaction sent from the child chain and executed on the parent chain.
// It assumes that a child chain has only one parent chain.
func (s *PublicServiceChainAPI) GetReceiptFromParentChain(blockHash common.Hash) *types.Receipt {
	return s.b.GetReceiptFromParentChain(blockHash)
}

// GetChainAccountAddr returns the current chain address setting.
func (s *PublicServiceChainAPI) GetChainAccountAddr() string {
	return s.b.GetChainAccountAddr()
}

// GetAnchoringPeriod returns the period (in child chain blocks) of sending chain transaction
// from child chain to parent chain.
func (s *PublicServiceChainAPI) GetAnchoringPeriod() uint64 {
	return s.b.GetAnchoringPeriod()
}

// GetSentChainTxsLimit returns the maximum number of stored chain transactions
// in child chain node, which is for resending.
func (s *PublicServiceChainAPI) GetSentChainTxsLimit() uint64 {
	return s.b.GetSentChainTxsLimit()
}
