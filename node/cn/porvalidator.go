// Copyright 2018 The klaytn Authors
// This file is part of the klaytn library.
//
// The klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the klaytn library. If not, see <http://www.gnu.org/licenses/>.

package cn

import (
	"context"
	"github.com/ground-x/klaytn/accounts/abi/bind"
	"github.com/ground-x/klaytn/blockchain/types"
	"github.com/ground-x/klaytn/client"
	"github.com/ground-x/klaytn/common"
	"github.com/ground-x/klaytn/contracts/reward/contract"
	"math/big"
)

func (pm *ProtocolManager) PoRValidate(from common.Address, tx *types.Transaction) error {

	if pm.rewardcontract == (common.Address{}) {
		logger.Error("reward contract address is not set")
		return nil
	}

	//TODO-GX manage target node to call smart-contract
	cnClient, err := client.Dial("ws://" + pm.getWSEndPoint())
	if err != nil {
		logger.Error("Fail to connect consensus node", "ws", pm.getWSEndPoint(), "err", err)
	}

	instance, err := contract.NewRNReward(common.HexToAddress(contract.RNRewardAddr), cnClient)
	if err != nil {
		return err
	}

	nonce, err := cnClient.PendingNonceAt(context.Background(), pm.rewardbase)
	if err != nil {
		logger.Error("fail to call pending nonce", "err", err)
	} else {
		logger.Error("nonce", "nonce", nonce)
	}

	var reward = big.NewInt(10)
	var chainID *big.Int
	transcOpt := bind.NewKeyedTransactorWithWallet(pm.rewardbase, pm.rewardwallet, chainID)
	transcOpt.Nonce = big.NewInt(int64(nonce))
	transcOpt.Value = reward
	transcOpt.GasLimit = uint64(117600)
	transcOpt.GasPrice = big.NewInt(0)

	_, rerr := instance.Reward(transcOpt, from)
	if rerr != nil {
		logger.Error("fail to call reward", "err", rerr)
	}

	logger.Error("received tx", "addr", from, "nonce", tx.Nonce(), "to", pm.rewardcontract, "value", tx.Value())

	return nil
}
