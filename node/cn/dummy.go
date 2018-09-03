package cn

import (
	"github.com/ground-x/go-gxplatform/blockchain/types"
	"github.com/ground-x/go-gxplatform/common"
	"github.com/ground-x/go-gxplatform/blockchain"
	"github.com/ground-x/go-gxplatform/event"
)

type EmptyTxPool struct {
	txFeed       event.Feed
}

func(re *EmptyTxPool) AddRemotes([]*types.Transaction) []error {
	return nil
}

func(re *EmptyTxPool) Pending() (map[common.Address]types.Transactions, error) {
	return map[common.Address]types.Transactions{}, nil
}

func(re *EmptyTxPool) SubscribeNewTxsEvent(newtxch chan<- blockchain.NewTxsEvent) event.Subscription {
	return re.txFeed.Subscribe(newtxch)
}
