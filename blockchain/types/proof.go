package types

import (
	"errors"
	"github.com/ground-x/go-gxplatform/common"
	"io"
	"github.com/ground-x/go-gxplatform/ser/rlp"
	"strings"
	"math/big"
)

var (
	ProofDigest = common.HexToHash("0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365")
	ErrInvalidProof = errors.New("invalid proof message data")
)

type Proof struct {
	Solver        common.Address
	BlockNumber   *big.Int
	Nonce         uint64
}

// EncodeRLP serializes ist into the Ethereum RLP format.
func (pr *Proof) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		pr.Solver,
		pr.BlockNumber,
		pr.Nonce,
	})
}

// DecodeRLP implements rlp.Decoder, and load the istanbul fields from a RLP stream.
func (pr *Proof) DecodeRLP(s *rlp.Stream) error {
	var proof struct {
		Solver        common.Address
		BlockNumber   *big.Int
		Nonce         uint64
	}
	if err := s.Decode(&proof); err != nil {
		return err
	}
	pr.Solver, pr.BlockNumber, pr.Nonce = proof.Solver, proof.BlockNumber, proof.Nonce
	return nil
}

func (pr *Proof) Compare(v Proof) bool {

	comp := strings.Compare(pr.Solver.String(),v.Solver.String())
	if comp != 0 {
		return false
	}

	if pr.BlockNumber.Cmp(v.BlockNumber) != 0 {
		return false
	}
	if (pr.Nonce - pr.Nonce) != 0 {
		return false
	}
	return true
}
