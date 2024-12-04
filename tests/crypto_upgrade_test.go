// This test file is used to @cruptoupgrade package

package tests

import (
	"math/big"
	"testing"
)

func TestCryptoUpgrade(t *testing.T) {
	// ! contract deployed in genesis block, will make mistakes
	backend := newTestBackend()
	miner := backend.CreateMiner()
	miner.Start()
	defer miner.Stop()
	// Upload
	input, err := contractAbi.Pack("uploadCode", "vdf", "hello world", big.NewInt(100))
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	tx1 := NewTx(backend.bc, 0, &contractAddress, big.NewInt(0), input)
	// Read
	input, err = contractAbi.Pack("getCode", "vdf")
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	tx2 := NewTx(backend.bc, 1, &contractAddress, big.NewInt(0), input)
	// Pull Code
	input, err = contractAbi.Pack("pullCode", "vdf")
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	tx3 := NewTx(backend.bc, 2, &contractAddress, big.NewInt(0), input)

	backend.AddTx(tx1)
	backend.AddTx(tx2)
	backend.AddTx(tx3)
}
