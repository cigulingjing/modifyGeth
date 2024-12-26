package tests

import (
	"math/big"
	"testing"
)

func TestInterest(t *testing.T) {
	backend := newTestBackend()
	miner := backend.CreateMiner()
	miner.Start()
	defer miner.Stop()
	tx0 := NewTx(backend.bc, 0, &(userAddress), big.NewInt(100), nil)
	backend.AddTx(tx0)
}
