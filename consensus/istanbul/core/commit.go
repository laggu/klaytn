package core

import (
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"reflect"
	)

func (c *core) sendCommit() {
	logger := c.logger.NewWith("state", c.state)
	if c.current.Preprepare == nil {
		logger.Error("Failed to get parentHash from roundState in sendCommit")
		return
	}

	sub := c.current.Subject()
	c.broadcastCommit(sub)
}

func (c *core) sendCommitForOldBlock(view *istanbul.View, digest common.Hash, prevHash common.Hash) {
	sub := &istanbul.Subject{
		View:   view,
		Digest: digest,
		PrevHash: prevHash,
	}
	c.broadcastCommit(sub)
}

func (c *core) broadcastCommit(sub *istanbul.Subject) {
	logger := c.logger.NewWith("state", c.state)

	encodedSubject, err := Encode(sub)
	if err != nil {
		logger.Error("Failed to encode", "subject", sub)
		return
	}

	c.broadcast(&message{
		Hash: sub.PrevHash,
		Code: msgCommit,
		Msg:  encodedSubject,
	})
}

func (c *core) handleCommit(msg *message, src istanbul.Validator) error {
	// Decode COMMIT message
	var commit *istanbul.Subject
	err := msg.Decode(&commit)
	if err != nil {
		return errFailedDecodeCommit
	}

	//logger.Error("receive handle commit","num", commit.View.Sequence)

	if err := c.checkMessage(msgCommit, commit.View); err != nil {
		//logger.Error("### istanbul/commit.go checkMessage","num",commit.View.Sequence,"err",err)
		return err
	}

	if err := c.verifyCommit(commit, src); err != nil {
		return err
	}

	c.acceptCommit(msg, src)

	// Commit the proposal once we have enough COMMIT messages and we are not in the Committed state.
	//
	// If we already have a proposal, we may have chance to speed up the consensus process
	// by committing the proposal without PREPARE messages.
	//logger.Error("### consensus check","len(commits)",c.current.Commits.Size(),"f(2/3)",2*c.valSet.F(),"state",c.state.Cmp(StateCommitted))
	if c.current.Commits.Size() > 2*c.valSet.F() && c.state.Cmp(StateCommitted) < 0 {
		// Still need to call LockHash here since state can skip Prepared state and jump directly to the Committed state.
		c.current.LockHash()
		c.commit()
	}

	return nil
}

// verifyCommit verifies if the received COMMIT message is equivalent to our subject
func (c *core) verifyCommit(commit *istanbul.Subject, src istanbul.Validator) error {
	logger := c.logger.NewWith("from", src, "state", c.state)

	sub := c.current.Subject()
	if !reflect.DeepEqual(commit, sub) {
		logger.Warn("Inconsistent subjects between commit and proposal", "expected", sub, "got", commit)
		return errInconsistentSubject
	}

	return nil
}

func (c *core) acceptCommit(msg *message, src istanbul.Validator) error {
	logger := c.logger.NewWith("from", src, "state", c.state)

	// Add the COMMIT message to current round state
	if err := c.current.Commits.Add(msg); err != nil {
		logger.Error("Failed to record commit message", "msg", msg, "err", err)
		return err
	}

	return nil
}
