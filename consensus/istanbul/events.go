package istanbul

import "github.com/ground-x/go-gxplatform/common"

// RequestEvent is posted to propose a proposal
type RequestEvent struct {
	Proposal Proposal
}

// MessageEvent is posted for Istanbul engine communication
type MessageEvent struct {
	Hash    common.Hash
	Payload []byte
}

type CommitEvent struct {
	Payload []byte
}

// FinalCommittedEvent is posted when a proposal is committed
type FinalCommittedEvent struct {
}
