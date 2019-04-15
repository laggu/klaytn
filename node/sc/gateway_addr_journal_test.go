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
	"fmt"
	"github.com/ground-x/klaytn/common"
	"os"
	"path"
	"testing"
)

func TestGateWayJournal(t *testing.T) {

	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), "test.rlp")); err != nil {
			t.Fatalf("fail to delete file %v", err)
		}
	}()

	journal := newGateWayAddrJournal(path.Join(os.TempDir(), "test.rlp"), &SCConfig{VTRecovery: true})
	if err := journal.load(func(journal GateWayJournal) error {
		fmt.Println("Local address ", journal.LocalAddress.Hex())
		fmt.Println("Remote address ", journal.RemoteAddress.Hex())
		fmt.Println("Paired", journal.Paired)
		return nil
	}); err != nil {
		t.Fatalf("fail to load journal %v", err)
	}
	if err := journal.rotate([]*GateWayJournal{}); err != nil {
		t.Fatalf("fail to rotate journal %v", err)
	}

	err := journal.insert(common.BytesToAddress([]byte("test1")), common.BytesToAddress([]byte("test2")), false)
	if err != nil {
		t.Fatalf("fail to insert address %v", err)
	}
	err = journal.insert(common.BytesToAddress([]byte("test2")), common.BytesToAddress([]byte("test3")), true)
	if err != nil {
		t.Fatalf("fail to insert address %v", err)
	}
	err = journal.insert(common.BytesToAddress([]byte("test3")), common.BytesToAddress([]byte("test1")), true)
	if err != nil {
		t.Fatalf("fail to insert address %v", err)
	}

	if err := journal.close(); err != nil {
		t.Fatalf("fail to close file %v", err)
	}

	journal = newGateWayAddrJournal(path.Join(os.TempDir(), "test.rlp"), &SCConfig{VTRecovery: true})

	if err := journal.load(func(journal GateWayJournal) error {
		if journal.LocalAddress.Hex() == "0x0000000000000000000000000000007465737431" {
			if journal.Paired {
				t.Fatalf("insert and load info mismatch: have %v, want %v", journal.Paired, false)
			}
		}
		if journal.LocalAddress.Hex() == "0x0000000000000000000000000000007465737432" &&
			journal.RemoteAddress.Hex() == "0x0000000000000000000000000000007465737433" {
			if !journal.Paired {
				t.Fatalf("insert and load info mismatch: have %v, want %v", journal.Paired, true)
			}
		}
		if journal.LocalAddress.Hex() == "0x0000000000000000000000000000007465737433" &&
			journal.RemoteAddress.Hex() == "0x0000000000000000000000000000007465737431" {
			if !journal.Paired {
				t.Fatalf("insert and load info mismatch: have %v, want %v", journal.Paired, true)
			}
		}
		fmt.Println("Local address ", journal.LocalAddress.Hex())
		fmt.Println("Remote address ", journal.RemoteAddress.Hex())
		fmt.Println("Paired", journal.Paired)
		return nil
	}); err != nil {
		t.Fatalf("fail to load journal %v", err)
	}
}

// TestGateWayJournalDisable tests insert method when VTRecovery is disabled.
func TestGateWayJournalDisable(t *testing.T) {
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), "test.rlp")); err != nil {
			t.Fatalf("fail to delete file %v", err)
		}
	}()

	// Step 1: Make new journal
	addrJournal := newGateWayAddrJournal(path.Join(os.TempDir(), "test.rlp"), &SCConfig{VTRecovery: false})

	if err := addrJournal.load(func(journal GateWayJournal) error {
		fmt.Println("Local address ", journal.LocalAddress.Hex())
		fmt.Println("Remote address ", journal.RemoteAddress.Hex())
		fmt.Println("Paired", journal.Paired)
		return nil
	}); err != nil {
		t.Fatalf("fail to load journal %v", err)
	}
	if err := addrJournal.rotate([]*GateWayJournal{}); err != nil {
		t.Fatalf("fail to rotate journal %v", err)
	}

	err := addrJournal.insert(common.BytesToAddress([]byte("test1")), common.BytesToAddress([]byte("test2")), false)
	if err != nil {
		t.Fatalf("fail to insert address %v", err)
	}

	if err := addrJournal.close(); err != nil {
		t.Fatalf("fail to close file %v", err)
	}

	// Step 2: Check journal is empty.
	addrJournal = newGateWayAddrJournal(path.Join(os.TempDir(), "test.rlp"), &SCConfig{VTRecovery: true})

	if err := addrJournal.load(func(journal GateWayJournal) error {
		fmt.Println("Local address ", journal.LocalAddress.Hex())
		fmt.Println("Remote address ", journal.RemoteAddress.Hex())
		fmt.Println("Paired", journal.Paired)
		addrJournal.cache = append(addrJournal.cache, &journal)
		return nil
	}); err != nil {
		t.Fatalf("fail to load journal %v", err)
	}

	fmt.Println("journal cache length", len(addrJournal.cache))
	if len(addrJournal.cache) > 0 {
		t.Fatalf("fail to disabling journal option %v", err)
	}
}
