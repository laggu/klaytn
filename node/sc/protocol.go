// Copyright 2018 The klaytn Authors
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
// This file is derived from eth/protocol.go (2018/06/04).
// Modified and improved for the klaytn development.

package sc

import (
	"github.com/ground-x/klaytn/common"
	"math/big"
)

const ProtocolMaxMsgSize = 10 * 1024 * 1024 // Maximum cap on the size of a protocol message

const (
	// Protocol messages belonging to servicechain/1
	StatusMsg              = 0x00
	NewBlockHashesMsg      = 0x01
	TxMsg                  = 0x02
	BlockHeadersRequestMsg = 0x03
	BlockHeadersMsg        = 0x04
	BlockBodiesRequestMsg  = 0x05
	BlockBodiesMsg         = 0x06
	NewBlockMsg            = 0x07

	ServiceChainTxsMsg                     = 0x08
	ServiceChainReceiptResponseMsg         = 0x09
	ServiceChainReceiptRequestMsg          = 0x0a
	ServiceChainParentChainInfoResponseMsg = 0x0b
	ServiceChainParentChainInfoRequestMsg  = 0x0c

	// Protocol messages belonging to klay/63
	NodeDataRequestMsg = 0x0d
	NodeDataMsg        = 0x0e
	ReceiptsRequestMsg = 0x0f
	ReceiptsMsg        = 0x10
)

// Protocol defines the protocol of the consensus
type SCProtocol struct {
	// Official short name of the protocol used during capability negotiation.
	Name string
	// Supported versions of the klaytn protocol (first is primary).
	Versions []uint
	// Number of implemented message corresponding to different protocol versions.
	Lengths []uint64
}

type errCode int

const (
	ErrMsgTooLarge = iota
	ErrDecode
	ErrInvalidMsgCode
	ErrProtocolVersionMismatch
	ErrNetworkIdMismatch
	ErrGenesisBlockMismatch
	ErrNoStatusMsg
	ErrExtraStatusMsg
	ErrSuspendedPeer
	ErrInvalidPeerHierarchy
	ErrUnexpectedTxType
	ErrFailedToGetStateDB
)

func (e errCode) String() string {
	return errorToString[int(e)]
}

// XXX change once legacy code is out
var errorToString = map[int]string{
	ErrMsgTooLarge:             "Message too long",
	ErrDecode:                  "Invalid message",
	ErrInvalidMsgCode:          "Invalid message code",
	ErrProtocolVersionMismatch: "Protocol version mismatch",
	ErrNetworkIdMismatch:       "NetworkId mismatch",
	ErrGenesisBlockMismatch:    "Genesis block mismatch",
	ErrNoStatusMsg:             "No status message",
	ErrExtraStatusMsg:          "Extra status message",
	ErrSuspendedPeer:           "Suspended peer",
	ErrInvalidPeerHierarchy:    "InvalidPeerHierarchy",
	ErrUnexpectedTxType:        "Unexpected tx type",
	ErrFailedToGetStateDB:      "Failed to get stateDB",
}

// statusData is the network packet for the status message.
type statusData struct {
	ProtocolVersion uint32
	NetworkId       uint64
	TD              *big.Int
	CurrentBlock    common.Hash
	GenesisBlock    common.Hash
	ChainID         *big.Int // A child chain must know parent chain's ChainID to sign a transaction.
	OnChildChain    bool     // OnChildChain presents if the peer is on child chain or not(same chain or parent chain).
}
