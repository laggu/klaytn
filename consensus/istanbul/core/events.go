package core

import "github.com/ground-x/go-gxplatform/consensus/istanbul"

type backlogEvent struct {
	src istanbul.Validator
	msg *message
}

type timeoutEvent struct{}
