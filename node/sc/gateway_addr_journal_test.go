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

	journal := newGateWayAddrJournal(path.Join(os.TempDir(), "test.rlp"))
	if err := journal.load(func(journal GateWayJournal) error {
		fmt.Println("Addr ", journal.Address.Hex())
		fmt.Println("IsLocal", journal.IsLocal)
		return nil
	}); err != nil {
		t.Fatalf("fail to load journal %v", err)
	}
	if err := journal.rotate([]*GateWayJournal{}); err != nil {
		t.Fatalf("fail to rotate journal %v", err)
	}

	err := journal.insert(common.BytesToAddress([]byte("test1")), true)
	if err != nil {
		t.Fatalf("fail to insert address %v", err)
	}
	err = journal.insert(common.BytesToAddress([]byte("test2")), false)
	if err != nil {
		t.Fatalf("fail to insert address %v", err)
	}
	err = journal.insert(common.BytesToAddress([]byte("test3")), true)
	if err != nil {
		t.Fatalf("fail to insert address %v", err)
	}

	if err := journal.close(); err != nil {
		t.Fatalf("fail to close file %v", err)
	}
	journal = newGateWayAddrJournal(path.Join(os.TempDir(), "test.rlp"))

	if err := journal.load(func(journal GateWayJournal) error {
		if journal.Address.Hex() == "0x0000000000000000000000000000007465737431" {
			if !journal.IsLocal {
				t.Fatalf("insert and load info mismatch: have %v, want %v", journal.IsLocal, true)
			}
		}
		if journal.Address.Hex() == "0x0000000000000000000000000000007465737432" {
			if journal.IsLocal {
				t.Fatalf("insert and load info mismatch: have %v, want %v", journal.IsLocal, false)
			}
		}
		if journal.Address.Hex() == "0x0000000000000000000000000000007465737433" {
			if !journal.IsLocal {
				t.Fatalf("insert and load info mismatch: have %v, want %v", journal.IsLocal, true)
			}
		}
		fmt.Println("Addr ", journal.Address.Hex())
		fmt.Println("IsLocal", journal.IsLocal)
		return nil
	}); err != nil {
		t.Fatalf("fail to load journal %v", err)
	}
}
