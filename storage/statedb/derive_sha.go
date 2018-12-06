package statedb

import (
	"bytes"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/ser/rlp"
)

type DeriveShaOrig struct{}

func (d DeriveShaOrig) DeriveSha(list types.DerivableList) common.Hash {
	keybuf := new(bytes.Buffer)
	trie := new(Trie)
	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		trie.Update(keybuf.Bytes(), list.GetRlp(i))
	}
	return trie.Hash()
}
