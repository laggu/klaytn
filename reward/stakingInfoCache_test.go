package reward

import "testing"

// test cache limit of stakingInfoCache
func TestStakingInfoCache_Add_Limit(t *testing.T) {
	stakingInfoCache := newStakingInfoCache()

	for i := 1; i <= 10; i++ {
		testStakingInfo, _ := newEmptyStakingInfo(uint64(i))
		stakingInfoCache.add(testStakingInfo)

		if len(stakingInfoCache.cells) > maxStakingCache {
			t.Errorf("over the max limit of staking cache. Current Len : %v, MaxStakingCache : %v", len(stakingInfoCache.cells), maxStakingCache)
		}
	}
}

func TestStakingInfoCache_Add_SameNumber(t *testing.T) {
	stakingInfoCache := newStakingInfoCache()

	testStakingInfo1, _ := newEmptyStakingInfo(uint64(1))
	testStakingInfo2, _ := newEmptyStakingInfo(uint64(1))

	stakingInfoCache.add(testStakingInfo1)
	stakingInfoCache.add(testStakingInfo2)

	if len(stakingInfoCache.cells) > 1 {
		t.Errorf("StakingInfo with Same block number is saved to the cache stakingCache. result : %v, expected : %v ", len(stakingInfoCache.cells), maxStakingCache)
	}
}

// stakingInfo with minBlockNum should be deleted if add more than limit
func TestStakingInfoCache_Add(t *testing.T) {
	stakingInfoCache := newStakingInfoCache()

	for i := 1; i < 5; i++ {
		testStakingInfo, _ := newEmptyStakingInfo(uint64(i))
		stakingInfoCache.add(testStakingInfo)
	}

	testStakingInfo, _ := newEmptyStakingInfo(uint64(5))
	stakingInfoCache.add(testStakingInfo) // blockNum 1 should be deleted
	if stakingInfoCache.minBlockNum != 2 {
		t.Errorf("minBlockNum of staking cache is different from expected blocknum. result : %v, expected : %v", stakingInfoCache.minBlockNum, 2)
	}

	testStakingInfo, _ = newEmptyStakingInfo(uint64(6))
	stakingInfoCache.add(testStakingInfo) // blockNum 2 should be deleted
	if stakingInfoCache.minBlockNum != 3 {
		t.Errorf("minBlockNum of staking cache is different from expected blocknum. result : %v, expected : %v", stakingInfoCache.minBlockNum, 3)
	}
}

func TestStakingInfoCache_Get(t *testing.T) {
	stakingInfoCache := newStakingInfoCache()

	for i := 1; i <= 4; i++ {
		testStakingInfo, _ := newEmptyStakingInfo(uint64(i))
		stakingInfoCache.add(testStakingInfo)
	}

	// should find correct stakingInfo with given block number
	for i := uint64(1); i <= 4; i++ {
		testStakingInfo := stakingInfoCache.get(i)

		if testStakingInfo.BlockNum != i {
			t.Errorf("The block number of staking info is different. result : %v, expected : %v", testStakingInfo.BlockNum, i)
		}
	}

	// nothing should be found as no matched block number is in cache
	for i := uint64(5); i < 10; i++ {
		testStakingInfo := stakingInfoCache.get(i)

		if testStakingInfo != nil {
			t.Errorf("The result should be nil. result : %v", testStakingInfo)
		}
	}
}
