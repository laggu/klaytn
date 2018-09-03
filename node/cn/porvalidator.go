package cn

import (
	"github.com/ground-x/go-gxplatform/contracts/reward/contract"
	"math/big"
	"github.com/ground-x/go-gxplatform/client"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"context"
)

func (pm *ProtocolManager) PoRValidate(from common.Address, tx *types.Transaction) error {

	if pm.rewardcontract == (common.Address{}) {
		log.Error("reward contract address is not set")
		return nil
	}

	//TODO-GX manage target node to call smart-contract
	cnClient, err := client.Dial("ws://" + pm.getWSEndPoint())
	if err != nil {
		log.Error("Fail to connect consensus node","ws",pm.getWSEndPoint(),"err",err)
	}

	instance, err := contract.NewRNReward(common.HexToAddress(contract.RNRewardAddr) , cnClient)
	if err != nil {
		return err
	}


	nonce, err := cnClient.PendingNonceAt(context.Background(), pm.rewardbase)
	if err != nil {
		log.Error("fail to call pending nonce", "err", err)
	}else{
		log.Error("nonce","nonce",nonce)
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
		log.Error("fail to call reward", "err", rerr)
	}

	log.Error("received tx","addr",from,"nonce",tx.Nonce(),"to",pm.rewardcontract,"value",tx.Value())

	return nil
}
