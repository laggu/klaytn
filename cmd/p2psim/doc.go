// Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from cmd/p2psim/main.go (2018/06/04).
// Modified and improved for the klaytn development.

// p2psim provides a command-line client for a simulation HTTP API.
//
// Here is an example of creating a 2 node network with the first node
// connected to the second:
//
//     $ p2psim node create
//     Created node01
//
//     $ p2psim node start node01
//     Started node01
//
//     $ p2psim node create
//     Created node02
//
//     $ p2psim node start node02
//     Started node02
//
//     $ p2psim node connect node01 node02
//     Connected node01 to node02
package main
