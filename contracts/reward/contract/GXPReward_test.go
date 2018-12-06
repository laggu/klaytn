package contract

import (
	"context"
	"fmt"
	"github.com/ground-x/go-gxplatform/accounts/abi/bind"
	"github.com/ground-x/go-gxplatform/accounts/abi/bind/backends"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/crypto"
	"log"
	"math/big"
	"testing"
)

func TestSmartContract(t *testing.T) {
	// Generate a new random account and a funded simulator
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	key2, _ := crypto.GenerateKey()
	auth2 := bind.NewKeyedTransactor(key2)

	alloc := blockchain.GenesisAlloc{auth.From: {Balance: big.NewInt(1000000000000000000)}, auth2.From: {Balance: big.NewInt(1000000000000000000)}}
	sim := backends.NewSimulatedBackend(alloc)

	// Deploy a token contract on the simulated blockchain
	_, _, reward, err := DeployGXPReward(auth, sim)
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}
	// Print the current (non existent) and pending name of the contract
	tx, err := reward.Reward(&bind.TransactOpts{From: auth.From, Signer: auth.Signer, Value: big.NewInt(500000000)}, auth2.From)
	if err != nil {
		log.Fatalf("Failed to call reward : %v", err)
	}
	fmt.Println("reward tx.hash :", tx.Hash().Hex())

	// Commit all pending transactions in the simulator and print the names again
	sim.Commit()

	balance, _ := reward.BalanceOf((&bind.CallOpts{Pending: true}), auth2.From)
	fmt.Println("balance :", balance)

	amount, _ := reward.TotalAmount((&bind.CallOpts{Pending: true}))
	fmt.Println("total amount :", amount)

	balance1, _ := sim.BalanceAt(context.Background(), auth2.From, big.NewInt(1))
	fmt.Println("before reward, balance :", balance1)

	tx2, err2 := reward.SafeWithdrawal(&bind.TransactOpts{From: auth2.From, Signer: auth2.Signer, Value: big.NewInt(0)})
	if err2 != nil {
		log.Fatalf("Failed to call reward : %v", err2)
	}
	fmt.Println("reward tx.hash :", tx2.Hash().Hex())

	sim.Commit()

	balance2, _ := sim.BalanceAt(context.Background(), auth2.From, big.NewInt(2))
	fmt.Println("after reward, balance :", balance2)

	if balance1.Cmp(balance2) >= 0 {
		log.Fatalf("Failed to withdraw safely")
	}
}
