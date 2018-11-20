package core

import (
	"github.com/ground-x/go-gxplatform/common"
)

func (c *core) handleFinalCommitted() error {
	logger := c.logger.NewWith("state", c.state)
	logger.Trace("Received a final committed proposal")
	c.startNewRound(common.Big0)
	return nil
}
