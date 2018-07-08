package gxp

import (
	"github.com/ground-x/go-gxplatform/contracts/reward/contract"
	"math/big"
	"github.com/ground-x/go-gxplatform/gxpclient"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/core/types"
	"context"
)

func (pm *ProtocolManager) PoRValidate(from common.Address, tx *types.Transaction) error {

	if pm.rewardcontract == (common.Address{}) {
		log.Error("reward contract address is not set")
		return nil
	}

	//TODO-GX manage target node to call smart-contract
	cnClient, err := gxpclient.Dial("ws://localhost:8546")
	if err != nil {
		log.Error("Fail to connect consensus node","err",err)
	}

	instance, err := contract.NewGXPReward(pm.rewardcontract , cnClient)
	if err != nil {
		return err
	}

	//file := "/Users/jun/data/istanbul/node6/keystore/UTC--2018-05-30T08-36-11.646393438Z--605c0b13afc58eb93088fda85cc6b9b94ce9c0d0"
	//password := "<password>"
	//
	//keyjson, err := ioutil.ReadFile(file)
	//key, err := keystore.DecryptKey(keyjson, password)
	//if err != nil {
	//	fmt.Println("json key failed to decrypt: %v", err)
	//}
	//
	//publicKey := key.PrivateKey.Public()
	//publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	//if !ok {
	//	log.Error("error casting public key to ECDSA")
	//}
	//
	//fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	//log.Error("public key","fromAddress",fromAddress)

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

	log.Error("received tx","addr",from,"nonce",tx.Nonce(),"to",tx.To(),"value",tx.Value())

	return nil
}
