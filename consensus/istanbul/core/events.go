package core

import (
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"github.com/ground-x/go-gxplatform/common"
)

type backlogEvent struct {
	src istanbul.Validator
	msg *message
	Hash common.Hash
}

type timeoutEvent struct{}
