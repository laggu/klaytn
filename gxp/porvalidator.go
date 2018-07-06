package gxp

import (
	"github.com/ground-x/go-gxplatform/contracts/reward/contract"
	"io/ioutil"
	"fmt"
	"crypto/ecdsa"
	"math/big"
	"github.com/ground-x/go-gxplatform/gxpclient"
	"github.com/ground-x/go-gxplatform/log"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/accounts/keystore"
	"github.com/ground-x/go-gxplatform/crypto"
	"context"
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/core/types"
)

func (pm *ProtocolManager) PoRValidate(from common.Address, tx *types.Transaction) error {

	cnClient, err := gxpclient.Dial("ws://localhost:8546")
	if err != nil {
		log.Error("Fail to connect consensus node","err",err)
	}

	address := common.HexToAddress("0xfc7e764a355a6d9fd2432c35e4c1e68cd8c4a12a")
	instance, err := contract.NewGXPReward(address, cnClient)
	if err != nil {
		return err
	}

	file := "/Users/jun/data/istanbul/node6/keystore/UTC--2018-05-30T08-36-11.646393438Z--605c0b13afc58eb93088fda85cc6b9b94ce9c0d0"
	password := "<password>"

	keyjson, err := ioutil.ReadFile(file)
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		fmt.Println("json key failed to decrypt: %v", err)
	}

	publicKey := key.PrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Error("public key","fromAddress",fromAddress)

	nonce, err := cnClient.PendingNonceAt(context.Background(), common.HexToAddress("0x605c0b13afc58eb93088fda85cc6b9b94ce9c0d0"))
	if err != nil {
		log.Error("fail to call pendingnonce", "err", err)
	}else{
		log.Error("nonce","nonce",nonce)
	}

	transcOpt := bind.NewKeyedTransactor(key.PrivateKey)
	transcOpt.Nonce = big.NewInt(int64(nonce))
	transcOpt.Value = big.NewInt(10)
	transcOpt.GasLimit = uint64(117600)
	transcOpt.GasPrice = big.NewInt(0)

	_, rerr := instance.Reward(transcOpt,common.HexToAddress("0x91d6f7d2537d8a0bd7d487dcc59151ebc00da706"))
	if rerr != nil {
		log.Error("fail to call reward", "err", rerr)
	}


	log.Error("received tx","addr",from,"nonce",tx.Nonce(),"to",tx.To(),"value",tx.Value())

	return nil
}
