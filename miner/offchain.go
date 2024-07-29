package miner

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

// create receipt for offchain compute
func applyTransactionOffchain(msg *core.Message, config *params.ChainConfig, gp *core.GasPool, statedb *state.StateDB, blockNumber *big.Int, blockHash common.Hash, tx *types.Transaction, usedGas *uint64, evm *vm.EVM) (*types.Receipt, error) {
	// Create a new context to be used in the EVM environment.
	txContext := core.NewEVMTxContext(msg)
	evm.Reset(txContext, statedb)
	// Apply the transaction to the current state (included in the env).
	log.Info("Executor ApplyMessage, it's OK!")
	result, err := core.ApplyMessage(evm, msg, gp)
	if err != nil {
		return nil, err
	}
	// Update the state with pending changes.
	var root []byte
	if config.IsByzantium(blockNumber) {
		statedb.Finalise(true)
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(blockNumber)).Bytes()
	}
	*usedGas += result.UsedGas

	// Create a new receipt for the transaction, storing the intermediate root and gas used
	// by the tx.
	receipt := &types.Receipt{Type: tx.Type(), PostState: root, CumulativeGasUsed: *usedGas}
	if result.Failed() {
		log.Error("receipt status failed", "tx", tx.Hash().Hex(), "err", result.Err)
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = result.UsedGas

	if tx.Type() == types.BlobTxType {
		receipt.BlobGasUsed = uint64(len(tx.BlobHashes()) * params.BlobTxBlobGasPerBlob)
		receipt.BlobGasPrice = evm.Context.BlobBaseFee
	}

	// If the transaction created a contract, store the creation address in the receipt.
	if msg.To == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
	}

	// Set the receipt logs and create the bloom filter.
	receipt.Logs = statedb.GetLogs(tx.Hash(), blockNumber.Uint64(), blockHash)
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	receipt.BlockHash = blockHash
	receipt.BlockNumber = blockNumber
	receipt.TransactionIndex = uint(statedb.TxIndex())
	return receipt, err
}

// offChain calcuation will push the result into executor's offChain channel
func (e *executor) offchainCalc(data []byte) (result [][]byte) {
	if len(data) >= 32*3 {
		//
		param1 := data[0:32]
		param2 := data[32:64]
		param3 := data[64:96]
		result = append(result, param1, param2, param3)
		e.offChainCh <- true
		return
	} else {
		log.Error("the transaction belong WASH ,but the input len is illegal")
	}
	return
}

// 1.catch the result of calcuation 2.make other operators
// @attention: transaction(which call offChainCalc) and transaction(call this fun) must in the same block
func (e *executor) offchainResultCatch() {
	// catch result from channel
	result := <-e.offChainCh
	if result {
		fmt.Printf("get WASM result:%v\n", result)
	} else {
		fmt.Printf("The offchainCh in executor is empty\n")
	}
	//TODO other operators after catch the result
}
