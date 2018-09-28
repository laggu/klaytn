package vm

import (
	"errors"
	"fmt"
	"github.com/ground-x/go-gxplatform/params"
)

// List execution errors
var (
	ErrOutOfGas                 = errors.New("out of gas")
	ErrCodeStoreOutOfGas        = errors.New("contract creation code storage out of gas")
	ErrDepth                    = errors.New("max call depth exceeded")
	ErrTraceLimitReached        = errors.New("the number of logs reached the specified limit")
	ErrInsufficientBalance      = errors.New("insufficient balance for transfer")
	ErrContractAddressCollision = errors.New("contract address collision")
	ErrTotalTimeLimitReached    = errors.New("reached the total execution time limit for txs in a block")
	ErrOpcodeCntLimitReached    = errors.New(fmt.Sprintf("reached the opcode count limit (%d) for tx", params.OpcodeCntLimit))
)
