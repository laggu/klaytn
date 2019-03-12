// Copyright 2018 The klaytn Authors
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from quorum/consensus/istanbul/backend/snapshot.go (2018/06/04).
// Modified and improved for the klaytn development.

package backend

import (
	"bytes"
	"encoding/json"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/consensus/istanbul"
	"github.com/ground-x/klaytn/consensus/istanbul/validator"
	"github.com/ground-x/klaytn/governance"
	"github.com/ground-x/klaytn/params"
	"github.com/ground-x/klaytn/ser/rlp"
	"github.com/ground-x/klaytn/storage/database"
)

const (
	dbKeySnapshotPrefix = "istanbul-snapshot"
)

// Vote represents a single vote that an authorized validator made to modify the
// list of authorizations.
type Vote struct {
	Validator common.Address `json:"validator"` // Authorized validator that cast this vote
	Block     uint64         `json:"block"`     // Block number the vote was cast in (expire old votes)
	Address   common.Address `json:"address"`   // Account being voted on to change its authorization
	Authorize bool           `json:"authorize"` // Whether to authorize or deauthorize the voted account
}

// Tally is a simple vote tally to keep the current score of votes. Votes that
// go against the proposal aren't counted since it's equivalent to not voting.
type Tally struct {
	Authorize bool   `json:"authorize"` // Whether the vote it about authorizing or kicking someone
	Votes     uint64 `json:"votes"`     // Number of votes until now wanting to pass the proposal
}

// Snapshot is the state of the authorization voting at a given point in time.
type Snapshot struct {
	Epoch                   uint64                        // The number of blocks after which to checkpoint and reset the pending votes
	Number                  uint64                        // Block number where the snapshot was created
	Hash                    common.Hash                   // Block hash where the snapshot was created
	Votes                   []*Vote                       // List of votes cast in chronological order
	GovernanceVotes         []*governance.GovernanceVote  // List of governance related votes cast in chronological order
	Tally                   map[common.Address]Tally      // Current vote tally to avoid recalculating
	ValSet                  istanbul.ValidatorSet         // Set of authorized validators at this moment
	GovernanceTally         []*governance.GovernanceTally // Current governance vote tally to avoid recalculating
	GovernanceConfig        *params.GovernanceConfig      // Pointer to the GovernanceConfig
	PendingGovernanceConfig *params.GovernanceConfig      // Stores current GovernanceConfig which haven't been applied yet
}

// newSnapshot create a new snapshot with the specified startup parameters. This
// method does not initialize the set of recent validators, so only ever use if for
// the genesis block.
func newSnapshot(epoch uint64, number uint64, hash common.Hash, valSet istanbul.ValidatorSet, gconfig *params.GovernanceConfig) *Snapshot {
	snap := &Snapshot{
		Epoch:                   gconfig.Istanbul.Epoch,
		Number:                  number,
		Hash:                    hash,
		ValSet:                  valSet,
		Tally:                   make(map[common.Address]Tally),
		GovernanceConfig:        gconfig,
		PendingGovernanceConfig: gconfig.Copy(),
	}
	return snap
}

// newPendingGovernanceConfig creates a new GovernanceConfig which holds changes that haven't been applied
func newPendingGovernanceConfig(src *params.GovernanceConfig) *params.GovernanceConfig {
	newConfig := &params.GovernanceConfig{
		Reward:   &params.RewardConfig{},
		Istanbul: &params.IstanbulConfig{},
	}
	copyGovernanceConfig(src, newConfig)
	return newConfig
}

// copyGovernanceConfig copies from source to destination governance config
func copyGovernanceConfig(src *params.GovernanceConfig, dst *params.GovernanceConfig) {
	if src == nil || dst == nil {
		logger.Error("Both source and destination shouldn't be nil")
		return
	}
	dst.GovernanceMode = src.GovernanceMode
	dst.Reward.MintingAmount.Set(src.Reward.MintingAmount)
	dst.Reward.Ratio = src.Reward.Ratio
	dst.Reward.UseGiniCoeff = src.Reward.UseGiniCoeff
	dst.UnitPrice = src.UnitPrice
	dst.GoverningNode = src.GoverningNode
	dst.Istanbul.SubGroupSize = src.Istanbul.SubGroupSize
	dst.Istanbul.Epoch = src.Istanbul.Epoch
	dst.Istanbul.ProposerPolicy = src.Istanbul.ProposerPolicy
}

// loadSnapshot loads an existing snapshot from the database.
func loadSnapshot(epoch uint64, subSize int, db database.DBManager, hash common.Hash, gconfig *params.GovernanceConfig) (*Snapshot, error) {
	blob, err := db.ReadIstanbulSnapshot(hash)
	if err != nil {
		return nil, err
	}
	snap := new(Snapshot)
	if err := json.Unmarshal(blob, snap); err != nil {
		return nil, err
	}
	snap.GovernanceConfig = gconfig
	snap.Epoch = snap.GovernanceConfig.Istanbul.Epoch
	snap.ValSet.SetSubGroupSize(snap.GovernanceConfig.Istanbul.SubGroupSize)
	if snap.PendingGovernanceConfig == nil {
		snap.PendingGovernanceConfig = snap.GovernanceConfig.Copy()
	}

	return snap, nil
}

// store inserts the snapshot into the database.
func (s *Snapshot) store(db database.DBManager) error {
	blob, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return db.WriteIstanbulSnapshot(s.Hash, blob)
}

// copy creates a deep copy of the snapshot, though not the individual votes.
func (s *Snapshot) copy() *Snapshot {
	cpy := &Snapshot{
		Epoch:                   s.Epoch,
		Number:                  s.Number,
		Hash:                    s.Hash,
		ValSet:                  s.ValSet.Copy(),
		Votes:                   make([]*Vote, len(s.Votes)),
		GovernanceVotes:         make([]*governance.GovernanceVote, len(s.GovernanceVotes)),
		Tally:                   make(map[common.Address]Tally),
		GovernanceTally:         make([]*governance.GovernanceTally, len(s.GovernanceTally)),
		GovernanceConfig:        s.GovernanceConfig.Copy(),
		PendingGovernanceConfig: s.PendingGovernanceConfig.Copy(),
	}

	for address, tally := range s.Tally {
		cpy.Tally[address] = tally
	}
	copy(cpy.Votes, s.Votes)
	copy(cpy.GovernanceTally, s.GovernanceTally)
	copy(cpy.GovernanceVotes, s.GovernanceVotes)
	return cpy
}

// checkVote return whether it's a valid vote
func (s *Snapshot) checkVote(address common.Address, authorize bool) bool {
	_, validator := s.ValSet.GetByAddress(address)
	return (validator != nil && !authorize) || (validator == nil && authorize)
}

// cast adds a new vote into the tally.
func (s *Snapshot) cast(address common.Address, authorize bool) bool {
	// Ensure the vote is meaningful
	if !s.checkVote(address, authorize) {
		return false
	}
	// Cast the vote into an existing or new tally
	_, validator := s.ValSet.GetByAddress(address)
	if old, ok := s.Tally[address]; ok {
		old.Votes += validator.VotingPower()
		s.Tally[address] = old
	} else {
		s.Tally[address] = Tally{Authorize: authorize, Votes: validator.VotingPower()}
	}
	return true
}

// uncast removes a previously cast vote from the tally.
func (s *Snapshot) uncast(address common.Address, authorize bool) bool {
	// If there's no tally, it's a dangling vote, just drop
	_, validator := s.ValSet.GetByAddress(address)
	tally, ok := s.Tally[address]
	if !ok {
		return false
	}
	// Ensure we only revert counted votes
	if tally.Authorize != authorize {
		return false
	}
	// Otherwise revert the vote
	if tally.Votes > validator.VotingPower() {
		tally.Votes -= validator.VotingPower()
		s.Tally[address] = tally
	} else {
		delete(s.Tally, address)
	}
	return true
}

// apply creates a new authorization snapshot by applying the given headers to
// the original one.
func (s *Snapshot) apply(headers []*types.Header, gov *governance.Governance) (*Snapshot, error) {
	// Allow passing in no headers for cleaner code
	if len(headers) == 0 {
		return s, nil
	}
	// Sanity check that the headers can be applied
	for i := 0; i < len(headers)-1; i++ {
		if headers[i+1].Number.Uint64() != headers[i].Number.Uint64()+1 {
			return nil, errInvalidVotingChain
		}
	}
	if headers[0].Number.Uint64() != s.Number+1 {
		return nil, errInvalidVotingChain
	}
	// Iterate through the headers and create a new snapshot
	snap := s.copy()

	for _, header := range headers {
		// Remove any votes on checkpoint blocks
		number := header.Number.Uint64()
		if number%s.Epoch == 0 {
			snap.Votes = nil
			snap.Tally = make(map[common.Address]Tally)
			snap.GovernanceVotes = nil
			snap.GovernanceTally = nil
			gov.ClearVotes()
		}
		// Resolve the authorization key and check against validators
		validator, err := ecrecover(header)
		if err != nil {
			return nil, err
		}
		if _, v := snap.ValSet.GetByAddress(validator); v == nil {
			return nil, errUnauthorized
		}

		// Header authorized, discard any previous votes from the validator
		for i, vote := range snap.Votes {
			if vote.Validator == validator && vote.Address == header.Coinbase {
				// Uncast the vote from the cached tally
				snap.uncast(vote.Address, vote.Authorize)

				// Uncast the vote from the chronological list
				snap.Votes = append(snap.Votes[:i], snap.Votes[i+1:]...)
				break // only one vote allowed
			}
		}

		// Check if this governance vote duplicates
		// Handle Governance related vote
		gVote := new(governance.GovernanceVote)
		snap.handleGovernanceVote(snap, header, validator, gVote, gov)

		// TODO-Klaytn-Governance: The code below is preservered not to touch former working codes
		//  But it would be much better to be integrated with governance vote
		// Tally up the new vote from the validator
		var authorize bool
		switch {
		case bytes.Compare(header.Nonce[:], nonceAuthVote) == 0:
			authorize = true
		case bytes.Compare(header.Nonce[:], nonceDropVote) == 0:
			authorize = false
		default:
			return nil, errInvalidVote
		}
		if snap.cast(header.Coinbase, authorize) {
			snap.Votes = append(snap.Votes, &Vote{
				Validator: validator,
				Block:     number,
				Address:   header.Coinbase,
				Authorize: authorize,
			})
		}
		// If the vote passed, update the list of validators
		governanceMode := governance.GovernanceModeMap[s.GovernanceConfig.GovernanceMode]
		governingNode := s.GovernanceConfig.GoverningNode
		if tally := snap.Tally[header.Coinbase]; governanceMode == governance.GovernanceMode_None ||
			(governanceMode == governance.GovernanceMode_Single && gVote.Validator == governingNode) ||
			(governanceMode == governance.GovernanceMode_Ballot && tally.Votes > snap.ValSet.TotalVotingPower()/2) {
			if tally.Authorize {
				snap.ValSet.AddValidator(header.Coinbase)
			} else {
				snap.ValSet.RemoveValidator(header.Coinbase)

				// Discard any previous votes the deauthorized validator cast
				for i := 0; i < len(snap.Votes); i++ {
					if snap.Votes[i].Validator == header.Coinbase {
						// Uncast the vote from the cached tally
						snap.uncast(snap.Votes[i].Address, snap.Votes[i].Authorize)

						// Uncast the vote from the chronological list
						snap.Votes = append(snap.Votes[:i], snap.Votes[i+1:]...)

						i--
					}
				}
			}
			// Discard any previous votes around the just changed account
			for i := 0; i < len(snap.Votes); i++ {
				if snap.Votes[i].Address == header.Coinbase {
					snap.Votes = append(snap.Votes[:i], snap.Votes[i+1:]...)
					i--
				}
			}
			delete(snap.Tally, header.Coinbase)
		}
	}
	snap.Number += uint64(len(headers))
	snap.Hash = headers[len(headers)-1].Hash()

	if snap.ValSet.Policy() == istanbul.WeightedRandom {
		// TODO-Klaytn-Issue1166 We have to update block number of ValSet too.
		snap.ValSet.SetBlockNum(snap.Number)
	}

	return snap, nil
}

func (s *Snapshot) handleGovernanceVote(snap *Snapshot, header *types.Header, validator common.Address, gVote *governance.GovernanceVote, gov *governance.Governance) {

	governanceMode := governance.GovernanceModeMap[s.GovernanceConfig.GovernanceMode]
	governingNode := s.GovernanceConfig.GoverningNode

	if len(header.Vote) > 0 {
		if err := rlp.DecodeBytes(header.Vote, gVote); err != nil {
			logger.Error("Failed to decode a vote. This vote ignored", "number", header.Number, "key", gVote.Key, "value", gVote.Value, "validator", gVote.Validator)
		} else {
			// Parse vote.Value and make it has appropriate type
			gVote = gov.ParseVoteValue(gVote)

			// Check vote's validity
			if gov.CheckVoteValidity(gVote.Key, gVote.Value) {
				// Remove old vote with same validator and key
				s.removePreviousVote(snap, validator, gVote)

				// Add new Vote to snapshot.GovernanceVotes
				snap.GovernanceVotes = append(snap.GovernanceVotes, gVote)

				// Tally up the new vote. This will be cleared when Epoch ended.
				// Add to GovermanceTally if it doesn't exist
				s.addNewVote(snap, gVote, governanceMode, governingNode)

			} else {
				logger.Warn("Received Vote was invalid", "number", header.Number, "Validator", gVote.Validator, "key", gVote.Key, "value", gVote.Value)
			}
		}
	}
}

func (s *Snapshot) addNewVote(snap *Snapshot, gVote *governance.GovernanceVote, governanceMode int, governingNode common.Address) {
	_, v := snap.ValSet.GetByAddress(gVote.Validator)
	if v != nil {
		vp := uint64(v.VotingPower())
		currentVotes := changeGovernanceTally(snap, gVote.Key, gVote.Value, vp, true)
		if (governanceMode == governance.GovernanceMode_None ||
			(governanceMode == governance.GovernanceMode_Single && gVote.Validator == governingNode)) ||
			(governanceMode == governance.GovernanceMode_Ballot && currentVotes > uint64(snap.ValSet.TotalVotingPower())/2) {
			governance.ReflectVotes(*gVote, snap.PendingGovernanceConfig)
		}
	}
}

func (s *Snapshot) removePreviousVote(snap *Snapshot, validator common.Address, gVote *governance.GovernanceVote) {
	// Removing duplicated previous GovernanceVotes
	for idx, vote := range snap.GovernanceVotes {
		// Check if previous vote from same validator exists
		if vote.Validator == validator && vote.Key == gVote.Key {
			// Reduce Tally
			_, v := snap.ValSet.GetByAddress(vote.Validator)
			vp := uint64(v.VotingPower())
			changeGovernanceTally(snap, vote.Key, vote.Value, vp, false)

			// Remove the old vote from GovernanceVotes
			snap.GovernanceVotes = append(snap.GovernanceVotes[:idx], snap.GovernanceVotes[idx+1:]...)
			break
		}
	}
}

// changeGovernanceTally updates snapshot's tally for governance votes.
func changeGovernanceTally(snap *Snapshot, key string, value interface{}, vote uint64, isAdd bool) uint64 {
	found := false
	var currentVote uint64

	for idx, v := range snap.GovernanceTally {
		if v.Key == key && v.Value == value {
			if isAdd {
				snap.GovernanceTally[idx].Votes += vote
			} else {
				snap.GovernanceTally[idx].Votes -= vote
			}
			currentVote = snap.GovernanceTally[idx].Votes
			found = true
			break
		}
	}

	if !found {
		snap.GovernanceTally = append(snap.GovernanceTally, &governance.GovernanceTally{Key: key, Value: value, Votes: vote})
		return vote
	} else {
		return currentVote
	}
}

// validators retrieves the list of authorized validators in ascending order.
func (s *Snapshot) validators() []common.Address {
	validators := make([]common.Address, 0, s.ValSet.Size())
	for _, validator := range s.ValSet.List() {
		validators = append(validators, validator.Address())
	}
	return sortValidatorArray(validators)
}

func sortValidatorArray(validators []common.Address) []common.Address {
	for i := 0; i < len(validators); i++ {
		for j := i + 1; j < len(validators); j++ {
			if bytes.Compare(validators[i][:], validators[j][:]) > 0 {
				validators[i], validators[j] = validators[j], validators[i]
			}
		}
	}
	return validators
}

type snapshotJSON struct {
	Epoch                   uint64                        `json:"epoch"`
	Number                  uint64                        `json:"number"`
	Hash                    common.Hash                   `json:"hash"`
	Votes                   []*Vote                       `json:"votes"`
	GovernanceVotes         []*governance.GovernanceVote  `json:"governancevotes"`
	Tally                   map[common.Address]Tally      `json:"tally"`
	GovernanceTally         []*governance.GovernanceTally `json:"governancetally"`
	GovernanceConfig        *params.GovernanceConfig      `json:"governanceconfig"`
	PendingGovernanceConfig *params.GovernanceConfig      `json:"pendinggovernanceconfig"`

	// for validator set
	Validators   []common.Address        `json:"validators"`
	Policy       istanbul.ProposerPolicy `json:"policy"`
	SubGroupSize int                     `json:"subgroupsize"`

	// for weighted validator
	RewardAddrs       []common.Address `json:"rewardAddrs"`
	VotingPowers      []uint64         `json:"votingPower"`
	Weights           []int            `json:"weight"`
	Proposers         []common.Address `json:"proposers"`
	ProposersBlockNum uint64           `json:"proposersBlockNum"`
}

func (s *Snapshot) toJSONStruct() *snapshotJSON {
	var rewardAddrs []common.Address
	var votingPowers []uint64
	var weights []int
	var proposers []common.Address
	var proposersBlockNum uint64

	// TODO-Klaytn-Issue1166 For weightedCouncil
	if s.ValSet.Policy() == istanbul.WeightedRandom {
		rewardAddrs, votingPowers, weights, proposers, proposersBlockNum = validator.GetWeightedCouncilData(s.ValSet)
	}

	return &snapshotJSON{
		Epoch:                   s.GovernanceConfig.Istanbul.Epoch, // s.Epoch
		Number:                  s.Number,
		Hash:                    s.Hash,
		Votes:                   s.Votes,
		Tally:                   s.Tally,
		Validators:              s.validators(),
		Policy:                  istanbul.ProposerPolicy(s.GovernanceConfig.Istanbul.ProposerPolicy), //s.ValSet.Policy(),
		SubGroupSize:            s.GovernanceConfig.Istanbul.SubGroupSize,                            // s.ValSet.SubGroupSize(),
		RewardAddrs:             rewardAddrs,
		VotingPowers:            votingPowers,
		Weights:                 weights,
		Proposers:               proposers,
		ProposersBlockNum:       proposersBlockNum,
		GovernanceVotes:         s.GovernanceVotes,
		GovernanceTally:         s.GovernanceTally,
		GovernanceConfig:        s.GovernanceConfig,
		PendingGovernanceConfig: s.PendingGovernanceConfig,
	}
}

// Unmarshal from a json byte array
func (s *Snapshot) UnmarshalJSON(b []byte) error {
	var j snapshotJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	s.Epoch = j.Epoch
	s.Number = j.Number
	s.Hash = j.Hash
	s.Votes = j.Votes
	s.Tally = j.Tally

	// TODO-Klaytn-Issue1166 For weightedCouncil
	if j.Policy == istanbul.WeightedRandom {
		s.ValSet = validator.NewWeightedCouncil(j.Validators, j.RewardAddrs, j.VotingPowers, j.Weights, istanbul.ProposerPolicy(j.GovernanceConfig.Istanbul.ProposerPolicy), j.SubGroupSize, j.Number, j.ProposersBlockNum, nil)
		validator.RecoverWeightedCouncilProposer(s.ValSet, j.Proposers)
	} else {
		s.ValSet = validator.NewSubSet(j.Validators, istanbul.ProposerPolicy(j.GovernanceConfig.Istanbul.ProposerPolicy), j.GovernanceConfig.Istanbul.SubGroupSize)
	}
	s.GovernanceVotes = j.GovernanceVotes
	s.GovernanceTally = j.GovernanceTally
	s.GovernanceConfig = j.GovernanceConfig

	if j.PendingGovernanceConfig == nil {
		s.PendingGovernanceConfig = j.GovernanceConfig.Copy()
	} else {
		s.PendingGovernanceConfig = j.PendingGovernanceConfig
	}
	return nil
}

// Marshal to a json byte array
func (s *Snapshot) MarshalJSON() ([]byte, error) {
	j := s.toJSONStruct()
	return json.Marshal(j)
}
