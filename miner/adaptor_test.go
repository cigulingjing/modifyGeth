package miner

import (
	"math/big"
	"testing"
)

// Test GasAdaptor

func TestGasAdaptorNormal(t *testing.T) {
	// Initialize with normal values
	ga := NewGasAdaptor(1000, 10000, 5000, 3, 10)

	// Test normal adjustment
	newGas, newAvgNum, newAvgDenom := ga.AdjustGas(6000, 1, 5000, 1)

	// Check if results are within expected range
	if newGas < 1000 || newGas > 10000 {
		t.Errorf("newGas out of range: got %d, want between %d and %d", newGas, 1000, 10000)
	}

	if newAvgNum == 0 || newAvgDenom == 0 {
		t.Error("average values should not be zero")
	}
}

func TestGasAdaptorEdgeCases(t *testing.T) {
	// Test min gas limit
	ga := NewGasAdaptor(1000, 10000, 5000, 3, 10)
	newGas, _, _ := ga.AdjustGas(100, 1, 100, 1)
	if newGas != 1000 {
		t.Errorf("expected min gas %d, got %d", 1000, newGas)
	}

	// Test max gas limit
	newGas, _, _ = ga.AdjustGas(20000, 1, 20000, 1)
	if newGas != 10000 {
		t.Errorf("expected max gas %d, got %d", 10000, newGas)
	}
}

func TestGasAdaptorPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with zero denominator")
		}
	}()
	NewGasAdaptor(1000, 10000, 5000, 3, 0)
}

// Test PoWAdaptor

func TestPoWAdaptorNormal(t *testing.T) {
	initialDiff := big.NewInt(1000000)
	minPrice := big.NewInt(100)
	maxPrice := big.NewInt(10000)

	pa := NewPoWAdaptor(
		3, 10, // targetPowRatio
		2, 10, // alpha
		8, 10, // fMin
		12, 10, // fMax
		initialDiff,
		1, 10, // kp
		1, 100, // ki
		minPrice,
		maxPrice,
	)

	// Test normal adjustment
	parentPrice := big.NewInt(1000)
	newDiff, newPrice, newAvgNum, newAvgDenom := pa.AdjustParameters(
		4, 10, // currentPowRatio
		3, 10, // parentAvgRatio
		parentPrice,
	)

	if newDiff.Sign() <= 0 {
		t.Error("new difficulty should be positive")
	}

	if newPrice.Cmp(minPrice) < 0 || newPrice.Cmp(maxPrice) > 0 {
		t.Error("new price out of valid range")
	}

	if newAvgNum == 0 || newAvgDenom == 0 {
		t.Error("average values should not be zero")
	}
}

func TestPoWAdaptorEdgeCases(t *testing.T) {
	initialDiff := big.NewInt(1000000)
	minPrice := big.NewInt(100)
	maxPrice := big.NewInt(10000)

	pa := NewPoWAdaptor(
		3, 10, // targetPowRatio
		2, 10, // alpha
		8, 10, // fMin
		12, 10, // fMax
		initialDiff,
		1, 10, // kp
		1, 100, // ki
		minPrice,
		maxPrice,
	)

	// Test min price boundary
	parentPrice := big.NewInt(100)
	_, newPrice, _, _ := pa.AdjustParameters(
		1, 10, // very low ratio to push price down
		1, 10,
		parentPrice,
	)

	if newPrice.Cmp(minPrice) < 0 {
		t.Error("price should not go below minimum")
	}

	// Test max price boundary
	parentPrice = big.NewInt(9000)
	_, newPrice, _, _ = pa.AdjustParameters(
		9, 10, // very high ratio to push price up
		9, 10,
		parentPrice,
	)

	if newPrice.Cmp(maxPrice) > 0 {
		t.Error("price should not go above maximum")
	}
}

func TestPoWAdaptorPanics(t *testing.T) {
	testCases := []struct {
		name string
		fn   func()
	}{
		{
			name: "zero alpha denominator",
			fn: func() {
				NewPoWAdaptor(3, 10, 2, 0, 8, 10, 12, 10,
					big.NewInt(1000000), 1, 10, 1, 100,
					big.NewInt(100), big.NewInt(10000))
			},
		},
		{
			name: "negative price",
			fn: func() {
				NewPoWAdaptor(3, 10, 2, 10, 8, 10, 12, 10,
					big.NewInt(1000000), 1, 10, 1, 100,
					big.NewInt(-100), big.NewInt(10000))
			},
		},
		{
			name: "min price greater than max price",
			fn: func() {
				NewPoWAdaptor(3, 10, 2, 10, 8, 10, 12, 10,
					big.NewInt(1000000), 1, 10, 1, 100,
					big.NewInt(10000), big.NewInt(100))
			},
		},
		{
			name: "zero initial difficulty",
			fn: func() {
				NewPoWAdaptor(3, 10, 2, 10, 8, 10, 12, 10,
					big.NewInt(0), 1, 10, 1, 100,
					big.NewInt(100), big.NewInt(10000))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for test case: %s", tc.name)
				}
			}()
			tc.fn()
		})
	}
}

func TestPoWAdaptorInvalidInputs(t *testing.T) {
	pa := NewPoWAdaptor(
		3, 10, // targetPowRatio
		2, 10, // alpha
		8, 10, // fMin
		12, 10, // fMax
		big.NewInt(1000000),
		1, 10, // kp
		1, 100, // ki
		big.NewInt(100),
		big.NewInt(10000),
	)

	testCases := []struct {
		name string
		fn   func()
	}{
		{
			name: "current ratio denominator zero",
			fn: func() {
				pa.AdjustParameters(1, 0, 3, 10, big.NewInt(1000))
			},
		},
		{
			name: "parent ratio denominator zero",
			fn: func() {
				pa.AdjustParameters(1, 10, 3, 0, big.NewInt(1000))
			},
		},
		{
			name: "nil parent price",
			fn: func() {
				pa.AdjustParameters(1, 10, 3, 10, nil)
			},
		},
		{
			name: "negative parent price",
			fn: func() {
				pa.AdjustParameters(1, 10, 3, 10, big.NewInt(-1000))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for test case: %s", tc.name)
				}
			}()
			tc.fn()
		})
	}
}
