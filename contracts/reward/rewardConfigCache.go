package reward

import (
	"errors"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/params"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

var rewardConfigCache *rewardConfig
var rewardConfigCacheLock sync.Mutex

func init() {
	rewardConfigCache = NewRewardConfig()
}

// Cache for parsed reward parameters from governance
type rewardConfig struct {
	blockNum uint64

	mintingAmount *big.Int
	cnRewardRatio *big.Int
	pocRatio      *big.Int
	kirRatio      *big.Int
	totalRatio    *big.Int
}

func NewRewardConfig() *rewardConfig {
	return &rewardConfig{
		mintingAmount: nil,
		cnRewardRatio: new(big.Int),
		pocRatio:      new(big.Int),
		kirRatio:      new(big.Int),
		totalRatio:    new(big.Int),
	}
}

func (config *rewardConfig) parseRewardRatio(ratio string) (int, int, int, error) {
	s := strings.Split(ratio, "/")
	if len(s) != 3 {
		return 0, 0, 0, errors.New("Invalid format")
	}
	cn, err1 := strconv.Atoi(s[0])
	poc, err2 := strconv.Atoi(s[1])
	kir, err3 := strconv.Atoi(s[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, errors.New("Parsing error")
	}
	return cn, poc, kir, nil
}

// getRewardGovernanceParameters retrieves reward parameters from governance. It also maintains a cache to reuse already parsed parameters.
func (configCache *rewardConfig) getRewardConfigCache(configManager ConfigManager, header *types.Header) *rewardConfig {
	rewardConfigCacheLock.Lock()
	defer rewardConfigCacheLock.Unlock()

	blockNum := header.Number.Uint64()

	// Cache hit condition
	// (1) blockNum is a key of cache.
	// (2) mintingAmount indicates whether cache entry is initialized or not
	// refresh at block (number -1) % epoch == 0 .
	// voting is calculated at epoch number in snapshot (which is 1 less than block header number)
	// cache refresh should be done after snapshot calculating vote.
	// so cache refresh for block header number should be 1 + epoch number
	// blockNumber cannot be 0 because this function is called by finalize() and finalize for genesis block isn't called
	epoch := configManager.Epoch()
	if (blockNum-1)%epoch == 0 || rewardConfigCache.blockNum+epoch < blockNum || rewardConfigCache.mintingAmount == nil {
		// Cache missed or not initialized yet. Let's parse governance parameters and update cache
		cn, poc, kir, err := rewardConfigCache.parseRewardRatio(configManager.Ratio())
		if err != nil {
			logger.Error("Error while parsing reward ratio of governance. Using default ratio", "err", err)

			cn = params.DefaultCNRewardRatio
			poc = params.DefaultPoCRewardRatio
			kir = params.DefaultKIRRewardRatio
		}

		// allocate new cache entry
		newRewardCache := NewRewardConfig()

		// update new cache entry
		if configManager.MintingAmount() == "" {
			logger.Error("No minting amount defined in governance. Use default value.", "Default minting amount", params.DefaultMintedKLAY)
			newRewardCache.mintingAmount = params.DefaultMintedKLAY
		} else {
			newRewardCache.mintingAmount, _ = big.NewInt(0).SetString(configManager.MintingAmount(), 10)
		}

		newRewardCache.blockNum = blockNum
		newRewardCache.cnRewardRatio.SetInt64(int64(cn))
		newRewardCache.pocRatio.SetInt64(int64(poc))
		newRewardCache.kirRatio.SetInt64(int64(kir))
		newRewardCache.totalRatio.Add(newRewardCache.cnRewardRatio, newRewardCache.pocRatio)
		newRewardCache.totalRatio.Add(newRewardCache.totalRatio, newRewardCache.kirRatio)

		// update cache
		rewardConfigCache = newRewardCache

		// TODO-Klaytn-RemoveLater Remove below trace later
		logger.Trace("Reward parameters updated from governance", "blockNum", newRewardCache.blockNum, "minting amount", newRewardCache.mintingAmount, "cn ratio", newRewardCache.cnRewardRatio, "poc ratio", newRewardCache.pocRatio, "kir ratio", newRewardCache.kirRatio)
	}

	return rewardConfigCache
}
