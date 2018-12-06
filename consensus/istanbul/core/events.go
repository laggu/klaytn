package core

import (
	"github.com/ground-x/go-gxplatform/common"
)

type backlogEvent struct {
	src  common.Address
	msg  *message
	Hash common.Hash
}

type timeoutEvent struct{}
