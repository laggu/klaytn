// Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
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
// This file is derived from core/genesis_alloc.go (2018/06/04).
// Modified and improved for the klaytn development.

package blockchain

// Constants containing the genesis allocation of built-in genesis blocks.
// Their content is an RLP-encoded list of (address, balance) tuples.
// Use mkalloc.go to create/update them.

// nolint: misspell
const mainnetAllocData = "\xda\u0654\x19t\x18\x1a?\x1bE+\x00\xec\u06ab\x19\xa2j\xf3\x8f\xcb\xff\x1e\x83\x98\x96\x80"

const testnetAllocData = "\xda\u0654\x19t\x18\x1a?\x1bE+\x00\xec\u06ab\x19\xa2j\xf3\x8f\xcb\xff\x1e\x83\x98\x96\x80"
