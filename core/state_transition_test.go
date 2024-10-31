package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestPowTransition(t *testing.T) {
	// Setup test environment
	db := rawdb.NewMemoryDatabase()
	state, _ := state.New(common.Hash{}, state.NewDatabase(db), nil)

	// Create test account with some balance
	from := common.HexToAddress("0x1234")
	state.SetBalance(from, uint256.NewInt(1000000000000000000)) // 1 ETH

	// Create message
	msg := &Message{
		From:     from,
		To:       &common.Address{2},
		Value:    big.NewInt(1000),
		GasLimit: 21000,
		IsPow:    true, // Set as POW transaction
	}

	// Setup EVM context
	context := vm.BlockContext{
		PowGas:   50000,         // Set POW gas limit
		PowPrice: big.NewInt(1), // Set POW gas price
		CanTransfer: func(state vm.StateDB, from common.Address, amount *uint256.Int) bool {
			return true
		},
		Transfer: func(state vm.StateDB, from common.Address, to common.Address, amount *uint256.Int) {
			state.SubBalance(from, amount)
			state.AddBalance(to, amount)
		},
	}

	// Create EVM instance
	evm := vm.NewEVM(context, vm.TxContext{}, state, params.TestChainConfig, vm.Config{})

	// Create gas pool
	gp := new(GasPool).AddGas(1000000)

	// Execute transition
	st := NewStateTransition(evm, msg, gp)
	result, err := st.TransitionDb()

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(21000), result.UsedGas)
	// Check incentive calculation uses pow price
	expectedIncentive := new(uint256.Int).SetUint64(result.UsedGas)
	expectedIncentive.Mul(expectedIncentive, uint256.MustFromBig(context.PowPrice))
	assert.Equal(t, expectedIncentive, result.Incentive, "Incentive should be calculated using pow price")

	// Check refund calculation uses pow price
	if result.RefundedGas > 0 {
		expectedRefund := new(uint256.Int).SetUint64(result.RefundedGas)
		expectedRefund.Mul(expectedRefund, uint256.MustFromBig(context.PowPrice))
		// Get actual refund by checking balance change
		actualBalance := state.GetBalance(from)
		expectedBalance := uint256.NewInt(1000000000000000000)               // Initial balance
		expectedBalance.Sub(expectedBalance, uint256.MustFromBig(msg.Value)) // Subtract transfer value
		expectedBalance.Sub(expectedBalance, expectedIncentive)              // Subtract gas cost
		expectedBalance.Add(expectedBalance, expectedRefund)                 // Add refund
		assert.Equal(t, expectedBalance, actualBalance, "Refund should be calculated using pow price")
	}
}
