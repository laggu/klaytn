package state

import (
	"bytes"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"github.com/ground-x/go-gxplatform/storage/statedb"
	"github.com/ground-x/go-gxplatform/storage/database"
)

// NewStateSync create a new state trie download scheduler.
func NewStateSync(root common.Hash, database database.DBManager) *statedb.TrieSync {
	var syncer *statedb.TrieSync
	callback := func(leaf []byte, parent common.Hash) error {
		var obj Account
		if err := rlp.Decode(bytes.NewReader(leaf), &obj); err != nil {
			return err
		}
		syncer.AddSubTrie(obj.Root, 64, parent, nil)
		syncer.AddRawEntry(common.BytesToHash(obj.CodeHash), 64, parent)
		return nil
	}
	syncer = statedb.NewTrieSync(root, database, callback)
	return syncer
}
