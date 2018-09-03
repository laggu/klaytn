package backend

import (
	"errors"
	"github.com/hashicorp/golang-lru"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/consensus"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"github.com/ground-x/go-gxplatform/networks/p2p"
	"github.com/ground-x/go-gxplatform/node"
)

const (
	istanbulMsg = 0x11
)

var (
	// errDecodeFailed is returned when decode message fails
	errDecodeFailed = errors.New("fail to decode istanbul message")
)

// Protocol implements consensus.Engine.Protocol
func (sb *backend) Protocol() consensus.Protocol {
	return consensus.Protocol{
		Name:     "istanbul",
		Versions: []uint{64},
		//Lengths:  []uint64{18},
		//Lengths:  []uint64{19},  // add PoRMsg
		Lengths:  []uint64{20},  // add PoRSendMsg
	}
}

// HandleMsg implements consensus.Handler.HandleMsg
func (sb *backend) HandleMsg(addr common.Address, msg p2p.Msg) (bool, error) {
	sb.coreMu.Lock()
	defer sb.coreMu.Unlock()

	if msg.Code == istanbulMsg {
		if !sb.coreStarted {
			return true, istanbul.ErrStoppedEngine
		}

		var cmsg istanbul.ConsensusMsg

		//var data []byte
		if err := msg.Decode(&cmsg); err != nil {
			return true, errDecodeFailed
		}
        data := cmsg.Payload
		hash := istanbul.RLPHash(data)

		// Mark peer's message
		ms, ok := sb.recentMessages.Get(addr)
		var m *lru.ARCCache
		if ok {
			m, _ = ms.(*lru.ARCCache)
		} else {
			m, _ = lru.NewARC(inmemoryMessages)
			sb.recentMessages.Add(addr, m)
		}
		m.Add(hash, true)

		// Mark self known message
		if _, ok := sb.knownMessages.Get(hash); ok {
			return true, nil
		}
		sb.knownMessages.Add(hash, true)

		go sb.istanbulEventMux.Post(istanbul.MessageEvent{
			Payload: data,
			Hash: cmsg.PrevHash,
		})

		return true, nil
	}
	return false, nil
}

func (sb *backend) ValidatePeerType(addr common.Address) bool {
	// istanbul.Start vs try to connect by peer
	for sb.chain == nil {
		return false
	}
	for _, val := range sb.getValidators(sb.chain.CurrentHeader().Number.Uint64(), sb.chain.CurrentHeader().Hash()).List() {
		if addr == val.Address() {
			return true
		}
	}
	return false
}

// SetBroadcaster implements consensus.Handler.SetBroadcaster
func (sb *backend) SetBroadcaster(broadcaster consensus.Broadcaster) {
	sb.broadcaster = broadcaster
	sb.broadcaster.RegisterValiator(node.CONSENSUSNODE, sb)
}

func (sb *backend) NewChainHead() error {
	sb.coreMu.RLock()
	defer sb.coreMu.RUnlock()
	if !sb.coreStarted {
		return istanbul.ErrStoppedEngine
	}
	go sb.istanbulEventMux.Post(istanbul.FinalCommittedEvent{})
	return nil
}
