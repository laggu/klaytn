package validator

import (
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/consensus/istanbul"
	"math"
	"reflect"
	"sort"
	"sync"
	"math/rand"
	"github.com/ground-x/go-gxplatform/log"
	"strings"
	"strconv"
	"fmt"
)

const (
	defaultSubSetLength = 21
)

type defaultValidator struct {
	address common.Address
}

func (val *defaultValidator) Address() common.Address {
	return val.address
}

func (val *defaultValidator) String() string {
	return val.Address().String()
}

func (val *defaultValidator) Equal(val2 *defaultValidator) bool {
	return val.address == val.address
}

func (val *defaultValidator) Hash() int64 {
	return val.address.Hash().Big().Int64()
}

type defaultSet struct {

	subSize    int

	validators istanbul.Validators
	policy     istanbul.ProposerPolicy

	proposer    istanbul.Validator
	validatorMu sync.RWMutex
	selector    istanbul.ProposalSelector
}

func newDefaultSet(addrs []common.Address, policy istanbul.ProposerPolicy) *defaultSet {
	valSet := &defaultSet{}

	valSet.subSize = defaultSubSetLength
	valSet.policy = policy
	// init validators
	valSet.validators = make([]istanbul.Validator, len(addrs))
	for i, addr := range addrs {
		valSet.validators[i] = New(addr)
	}
	// sort validator
	sort.Sort(valSet.validators)
	// init proposer
	if valSet.Size() > 0 {
		valSet.proposer = valSet.GetByIndex(0)
	}
	valSet.selector = roundRobinProposer
	if policy == istanbul.Sticky {
		valSet.selector = stickyProposer
	}

	return valSet
}

func newDefaultSubSet(addrs []common.Address, policy istanbul.ProposerPolicy, subSize int) *defaultSet {
	valSet := &defaultSet{}

	valSet.subSize = subSize
	valSet.policy = policy
	// init validators
	valSet.validators = make([]istanbul.Validator, len(addrs))
	for i, addr := range addrs {
		valSet.validators[i] = New(addr)
	}
	// sort validator
	sort.Sort(valSet.validators)
	// init proposer
	if valSet.Size() > 0 {
		valSet.proposer = valSet.GetByIndex(0)
	}
	valSet.selector = roundRobinProposer
	if policy == istanbul.Sticky {
		valSet.selector = stickyProposer
	}

	return valSet
}

func (valSet *defaultSet) Size() int {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	return len(valSet.validators)
}

func (valSet *defaultSet) SubGroupSize() int {
	return valSet.subSize
}

func (valSet *defaultSet) SetSubGroupSize(size int) {
	valSet.subSize = size
}

func (valSet *defaultSet) List() []istanbul.Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	return valSet.validators
}

func (valSet *defaultSet) SubList(prevHash common.Hash) []istanbul.Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()

	if len(valSet.validators) <= valSet.subSize {
		return valSet.validators
	}
	hashstring := strings.TrimPrefix(prevHash.Hex(),"0x")
	if len(hashstring) > 15 {
		hashstring = hashstring[:15]
	}
	seed, err := strconv.ParseInt(hashstring, 16, 64)
	if err != nil {
		log.Error("input" ,"hash", prevHash.Hex())
		log.Error("fail to make sub-list of validators","seed", seed, "err",err)
		return valSet.validators
	}

	// shuffle
	subset := make([]istanbul.Validator,valSet.subSize)
	subset[0] = valSet.GetProposer()
	// next proposer
	// TODO how to sync next proposer (how to get exact next proposer ?)
	subset[1] = valSet.selector(valSet, subset[0].Address() , uint64(0))

	proposerIdx, _ := valSet.GetByAddress(subset[0].Address())
	nextproposerIdx, _ := valSet.GetByAddress(subset[1].Address())

	if proposerIdx == nextproposerIdx {
		log.Error("fail to make propser","current proposer idx", proposerIdx, "next idx", nextproposerIdx)
	}

	limit := len(valSet.validators)
	picker := rand.New(rand.NewSource(seed))

	pickSize := limit - 2
	indexs := make([]int, pickSize)
	idx := 0
	for i := 0; i < limit; i++ {
		if i != proposerIdx && i != nextproposerIdx {
			indexs[idx] = i
			idx++
		}
	}
	for i := 0; i < pickSize; i++ {
		randIndex := picker.Intn(pickSize)
		indexs[i], indexs[randIndex] = indexs[randIndex], indexs[i]
	}

	for i :=0; i < valSet.subSize-2; i++ {
		subset[i+2] = valSet.validators[indexs[i]]
	}

	if prevHash.Hex() == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		log.Error("### subList","prevHash", prevHash.Hex())
	}

	return subset
}

func (valSet *defaultSet) SubListWithProposer(prevHash common.Hash, proposer common.Address) []istanbul.Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()

	if len(valSet.validators) <= valSet.subSize {
		return valSet.validators
	}
	hashstring := strings.TrimPrefix(prevHash.Hex(),"0x")
	if len(hashstring) > 15 {
		hashstring = hashstring[:15]
	}
	seed, err := strconv.ParseInt(hashstring, 16, 64)
	if err != nil {
		log.Error("input" ,"hash", prevHash.Hex())
		log.Error("fail to make sub-list of validators","seed", seed, "err",err)
		return valSet.validators
	}

	// shuffle
	subset := make([]istanbul.Validator,valSet.subSize)
	subset[0] = New(proposer)
	// next proposer
	// TODO how to sync next proposer (how to get exact next proposer ?)
	subset[1] = valSet.selector(valSet, subset[0].Address() , uint64(0))

	proposerIdx, _ := valSet.GetByAddress(subset[0].Address())
	nextproposerIdx, _ := valSet.GetByAddress(subset[1].Address())

	// TODO-GX: remove this check code if the implementation is stable.
	if proposerIdx < 0 || nextproposerIdx < 0 {
		vals := "["
		for _, v := range valSet.validators {
			vals += fmt.Sprintf("%s,", v.Address().Hex())
		}
		vals += "]"
		log.Error("idx should not be negative!", "proposerIdx", proposerIdx, "nextproposerIdx", nextproposerIdx, "proposer", subset[0].Address().Hex(),
			"nextproposer", subset[1].Address().Hex(), "validators", vals)
	}

	if proposerIdx == nextproposerIdx {
		log.Error("fail to make propser","current proposer idx", proposerIdx, "next idx", nextproposerIdx)
	}

	limit := len(valSet.validators)
	picker := rand.New(rand.NewSource(seed))

	pickSize := limit - 2
	indexs := make([]int, pickSize)
	idx := 0
	for i := 0; i < limit; i++ {
		if i != proposerIdx && i != nextproposerIdx {
			indexs[idx] = i
			idx++
		}
	}
	for i := 0; i < pickSize; i++ {
		randIndex := picker.Intn(pickSize)
		indexs[i], indexs[randIndex] = indexs[randIndex], indexs[i]
	}

	for i :=0; i < valSet.subSize-2; i++ {
		subset[i+2] = valSet.validators[indexs[i]]
	}

	if prevHash.Hex() == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		log.Error("### subList","prevHash", prevHash.Hex())
	}

	return subset
}

func (valSet *defaultSet) IsSubSet() bool {
    return valSet.Size() > valSet.subSize
}

func (valSet *defaultSet) GetByIndex(i uint64) istanbul.Validator {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	if i < uint64(valSet.Size()) {
		return valSet.validators[i]
	}
	return nil
}

func (valSet *defaultSet) GetByAddress(addr common.Address) (int, istanbul.Validator) {
	for i, val := range valSet.List() {
		if addr == val.Address() {
			return i, val
		}
	}
	return -1, nil
}

func (valSet *defaultSet) GetProposer() istanbul.Validator {
	return valSet.proposer
}

func (valSet *defaultSet) IsProposer(address common.Address) bool {
	_, val := valSet.GetByAddress(address)
	return reflect.DeepEqual(valSet.GetProposer(), val)
}

func (valSet *defaultSet) CalcProposer(lastProposer common.Address, round uint64) {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()
	valSet.proposer = valSet.selector(valSet, lastProposer, round)
}

func calcSeed(valSet istanbul.ValidatorSet, proposer common.Address, round uint64) uint64 {
	offset := 0
	if idx, val := valSet.GetByAddress(proposer); val != nil {
		offset = idx
	}
	return uint64(offset) + round
}

func emptyAddress(addr common.Address) bool {
	return addr == common.Address{}
}

func roundRobinProposer(valSet istanbul.ValidatorSet, proposer common.Address, round uint64) istanbul.Validator {
	if valSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if emptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(valSet, proposer, round) + 1
	}
	pick := seed % uint64(valSet.Size())
	return valSet.GetByIndex(pick)
}

func stickyProposer(valSet istanbul.ValidatorSet, proposer common.Address, round uint64) istanbul.Validator {
	if valSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if emptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(valSet, proposer, round)
	}
	pick := seed % uint64(valSet.Size())
	return valSet.GetByIndex(pick)
}

func (valSet *defaultSet) AddValidator(address common.Address) bool {
	valSet.validatorMu.Lock()
	defer valSet.validatorMu.Unlock()
	for _, v := range valSet.validators {
		if v.Address() == address {
			return false
		}
	}
	valSet.validators = append(valSet.validators, New(address))
	// TODO: we may not need to re-sort it again
	// sort validator
	sort.Sort(valSet.validators)
	return true
}

func (valSet *defaultSet) RemoveValidator(address common.Address) bool {
	valSet.validatorMu.Lock()
	defer valSet.validatorMu.Unlock()

	for i, v := range valSet.validators {
		if v.Address() == address {
			valSet.validators = append(valSet.validators[:i], valSet.validators[i+1:]...)
			return true
		}
	}
	return false
}

func (valSet *defaultSet) Copy() istanbul.ValidatorSet {
	valSet.validatorMu.RLock()
	defer valSet.validatorMu.RUnlock()

	addresses := make([]common.Address, 0, len(valSet.validators))
	for _, v := range valSet.validators {
		addresses = append(addresses, v.Address())
	}
	return NewSubSet(addresses, valSet.policy, valSet.subSize)
}

func (valSet *defaultSet) F() int {
	if valSet.Size() > valSet.subSize {
		return int(math.Ceil(float64(valSet.subSize)/3)) - 1
	} else {
		return int(math.Ceil(float64(valSet.Size())/3)) - 1
	}
}

func (valSet *defaultSet) Policy() istanbul.ProposerPolicy { return valSet.policy }
