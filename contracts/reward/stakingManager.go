package reward

import (
	"github.com/ground-x/klaytn/blockchain"
	"github.com/ground-x/klaytn/event"
	"github.com/ground-x/klaytn/params"
)

type StakingManager struct {
	abm              *AddressBookManager
	stakingInfoCache *stakingInfoCache

	bc           *blockchain.BlockChain
	chainHeadCh  chan blockchain.ChainHeadEvent
	chainHeadSub event.Subscription
}

func NewStakingManager(bc *blockchain.BlockChain) *StakingManager {
	abm := NewAddressBookManager(bc)
	sc := NewStakingInfoCache()

	return &StakingManager{
		abm:              abm,
		stakingInfoCache: sc,
		bc:               bc,
		chainHeadCh:      make(chan blockchain.ChainHeadEvent, chainHeadChanSize),
	}
}

// GetStakingInfoFromStakingCache returns corresponding staking information for a block of blockNum.
func (sm *StakingManager) GetStakingInfoFromStakingCache(blockNum uint64) *StakingInfo {
	number := params.CalcStakingBlockNumber(blockNum)

	stakingInfo, err := sm.updateStakingCache(number)
	if err != nil {
		logger.Error("Failed to get staking information", "Block number", blockNum, "number of staking block", number, "err", err)
		return nil
	}

	if stakingInfo.BlockNum != number {
		logger.Error("Invalid staking info from staking cache", "Block number", blockNum, "expected staking block number", number, "actual staking block number", stakingInfo.BlockNum)
		return nil
	}

	logger.Debug("Staking cache hit.", "Block number", blockNum, "number of staking block", number)
	return stakingInfo
}

// updateStakingCache updates staking cache with staking information of given block number.
func (sm *StakingManager) updateStakingCache(blockNum uint64) (*StakingInfo, error) {
	if cachedStakingInfo := sm.stakingInfoCache.get(blockNum); cachedStakingInfo != nil {
		// already updated
		return cachedStakingInfo, nil
	}

	stakingInfo, err := sm.abm.getStakingInfoFromAddressBook(blockNum)
	if err != nil {
		return nil, err
	}

	sm.stakingInfoCache.add(stakingInfo)

	logger.Info("Add new staking information to staking cache", "stakingInfo", stakingInfo)

	return stakingInfo, nil
}

// Subscribe setups a channel to listen chain head event and starts a goroutine to update staking cache.
func (sm *StakingManager) Subscribe() {
	sm.chainHeadSub = sm.bc.SubscribeChainHeadEvent(sm.chainHeadCh)

	go sm.waitHeadChain()
}

func (sm *StakingManager) waitHeadChain() {
	defer sm.chainHeadSub.Unsubscribe()

	logger.Info("Start listening chain head event to update staking cache.")

	for {
		// A real event arrived, process interesting content
		select {
		// Handle ChainHeadEvent
		case ev := <-sm.chainHeadCh:
			if sm.bc.ProposerPolicy() == params.WeightedRandom && params.IsStakingUpdateInterval(ev.Block.NumberU64()) {
				blockNum := ev.Block.NumberU64()
				logger.Debug("ChainHeadEvent arrived and try to update staking cache.", "Block number", blockNum)
				if _, err := sm.updateStakingCache(blockNum); err != nil {
					logger.Error("Failed to update staking cache", "err", err)
				}
			}
		case <-sm.chainHeadSub.Err():
			return
		}
	}
}

func (sm *StakingManager) Unsubscribe() {
	sm.chainHeadSub.Unsubscribe()
}
