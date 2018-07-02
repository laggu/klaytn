package core

import (
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/core/types"
	"github.com/ground-x/go-gxplatform/rlp"
)

func (c *core) sendProofTask(request *istanbul.Request) {
	log.Info("Proof Of Replication")
	//logger := c.logger.New("state", c.state)

	if c.backend.CurrentBlock().NumberU64() % 10 != 0 {
		return
	}

	targets := make(map[common.Address]bool)
	for _, addr := range c.backend.GetPeers() {
		var notval = true
		for _, val := range c.valSet.List() {
			if addr == val.Address() {
				notval = false
			}
		}
		if notval {
			targets[addr] = true
		}
	}

	proof := &types.Proof{
		Solver:       c.backend.Address(),
		BlockNumber:  c.backend.CurrentBlock().Number(),
		Nonce: 	      c.backend.CurrentBlock().Nonce(),
	}

	proofbytes, err := rlp.EncodeToBytes(proof)
	if err != nil {
		log.Error("fail to encode proof message","err",err)
		return
	}

	// broadcast another consensus nodes
	c.broadcast(&message{
		Code: msgStartProofRound,
		Msg:  proofbytes,
	})

	// TODO after +2/3 receive messages from another consensus nodes, send message to ranger node
	// send ranger node directly
	//c.backend.GossipPoRMsg(targets, proofbytes)
	c.backend.GossipProof(targets, *proof)
}

func (c *core) handleProofTask(msg *message, src istanbul.Validator) error {
	logger := c.logger.New("from", src, "state", c.state)

	// Check if the message comes from current proposer
	if !c.valSet.IsProposer(src.Address()) {
		logger.Warn("Ignore startProofTask messages from non-proposer")
		return errNotFromProposer
	}

	targets := make(map[common.Address]bool)
	// exclude validator nodes, only send ranger nodes
	for _, addr := range c.backend.GetPeers() {
		var notval = true
		for _, val := range c.valSet.List() {
			if addr == val.Address() {
				notval = false
			}
		}
		if notval {
			targets[addr] = true
		}
	}

	if msg.Code == msgStartProofRound {

		proof := new(types.Proof)
		if err := msg.Decode(&proof); err != nil {
			log.Error("Invalid proof RLP", "err", err)
		}else {
			c.backend.GossipProof(targets, *proof)
		}
	}

	return nil
}