package miner

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
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

func TestPoWAdaptorContinuousBlocks(t *testing.T) {
	// Initialize adaptor with reasonable parameters
	initialDiff := big.NewInt(1000000)
	minPrice := big.NewInt(100)
	maxPrice := big.NewInt(10000)

	pa := NewPoWAdaptor(
		3, 10, // targetPowRatio = 0.3
		2, 10, // alpha = 0.2 (EMA weight)
		8, 10, // fMin = 0.8
		12, 10, // fMax = 1.2
		initialDiff,
		1, 10, // kp = 0.1
		1, 100, // ki = 0.01
		minPrice,
		maxPrice,
	)

	type blockMetrics struct {
		difficulty *big.Int
		price      *big.Int
		powRatio   float64
		avgRatio   float64
	}

	// Store metrics for analysis
	metrics := make([]blockMetrics, 0)

	// Initial values
	currentDiff := initialDiff
	currentPrice := new(big.Int).Div(
		new(big.Int).Add(minPrice, maxPrice),
		big.NewInt(2),
	)
	avgRatioNum := uint64(3)
	avgRatioDenom := uint64(10)

	// Simulate 1000 blocks
	numBlocks := 1000

	// Helper to calculate simulated PoW ratio based on difficulty
	simulatePowRatio := func(diff *big.Int, price *big.Int) (uint64, uint64) {
		// Simulate miners responding to difficulty/price
		// Higher difficulty -> lower ratio
		// Higher price -> higher ratio
		baseRatio := 0.3 // Target ratio
		diffEffect := -0.1 * float64(diff.Int64()) / float64(initialDiff.Int64())
		priceEffect := 0.1 * float64(price.Int64()) / float64(maxPrice.Int64())

		ratio := baseRatio + diffEffect + priceEffect

		// Add some random noise
		ratio += (rand.Float64() - 0.5) * 0.1

		// Clamp between 0 and 1
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}

		return uint64(ratio * 1000), 1000
	}

	for i := 0; i < numBlocks; i++ {
		fmt.Println("block", i)
		// Simulate current block's PoW ratio
		currentRatioNum, currentRatioDenom := simulatePowRatio(currentDiff, currentPrice)

		// Get new parameters
		newDiff, newPrice, newAvgNum, newAvgDenom := pa.AdjustParameters(
			currentRatioNum, currentRatioDenom,
			avgRatioNum, avgRatioDenom,
			currentPrice,
		)
		if newAvgNum > newAvgDenom {
			fmt.Println("!")
		}

		// Store metrics
		metrics = append(metrics, blockMetrics{
			difficulty: new(big.Int).Set(newDiff),
			price:      new(big.Int).Set(newPrice),
			powRatio:   float64(currentRatioNum) / float64(currentRatioDenom),
			avgRatio:   float64(newAvgNum) / float64(newAvgDenom),
		})

		// Update for next block
		currentDiff = newDiff
		currentPrice = newPrice
		avgRatioNum = newAvgNum
		avgRatioDenom = newAvgDenom
	}

	// Analyze results

	// 1. Verify convergence to target ratio
	targetRatio := 0.3
	finalAvgRatio := metrics[len(metrics)-1].avgRatio
	if math.Abs(finalAvgRatio-targetRatio) > 0.1 {
		t.Errorf("Failed to converge to target ratio. Got %v, want %v ± 0.1",
			finalAvgRatio, targetRatio)
	}

	// 2. Check price stays within bounds
	for i, m := range metrics {
		if m.price.Cmp(minPrice) < 0 || m.price.Cmp(maxPrice) > 0 {
			t.Errorf("Price out of bounds at block %d: %v", i, m.price)
		}
	}

	// 3. Verify no extreme oscillations in difficulty
	for i := 1; i < len(metrics); i++ {
		diff := new(big.Int).Sub(metrics[i].difficulty, metrics[i-1].difficulty)
		maxChange := new(big.Int).Div(metrics[i-1].difficulty, big.NewInt(5))
		maxChange.Add(maxChange, big.NewInt(1))
		if diff.Abs(diff).Cmp(maxChange) > 0 {
			t.Errorf("Too large difficulty change at block %d: %v > %v", i, diff, maxChange)
		}
	}

	// 4. Verify system stability in last 100 blocks
	var ratioVariance float64
	avgRatio := 0.0
	for i := len(metrics) - 100; i < len(metrics); i++ {
		avgRatio += metrics[i].powRatio
	}
	avgRatio /= 100

	for i := len(metrics) - 100; i < len(metrics); i++ {
		diff := metrics[i].powRatio - avgRatio
		ratioVariance += diff * diff
	}
	ratioVariance /= 100

	if ratioVariance > 0.01 {
		t.Errorf("System not stable in final blocks. Variance: %v", ratioVariance)
	}
}

func TestGasAdaptorContinuousBlocks(t *testing.T) {
	// Initialize adaptor with reasonable parameters
	minGas := uint64(5000000)
	maxGas := uint64(30000000)
	initialGas := uint64(15000000)

	ga := NewGasAdaptor(
		minGas,
		maxGas,
		initialGas,
		2, 10, // alpha = 0.2 (EMA weight)
	)

	type blockMetrics struct {
		gas      uint64
		avgRatio float64
	}

	// Store metrics for analysis
	metrics := make([]blockMetrics, 0)

	// Initial values
	currentGas := initialGas
	avgRatioNum := uint64(15000000)  // Initial target
	avgRatioDenom := uint64(1000000) // Scale factor

	// Simulate 1000 blocks
	numBlocks := 1000

	// Helper to calculate simulated gas usage based on current gas limit
	simulateGasUsage := func(gasLimit uint64) (uint64, uint64) {
		// Simulate blocks using around 50% of gas limit with some random variation
		baseUsage := 0.5
		randomVariation := (rand.Float64() - 0.5) * 0.2 // ±10% variation
		usage := baseUsage + randomVariation

		// Clamp between 0 and 1
		if usage < 0 {
			usage = 0
		}
		if usage > 1 {
			usage = 1
		}

		// Convert to fraction with common scale factor
		scaleFactor := uint64(1000000)
		return uint64(usage * float64(gasLimit)), scaleFactor
	}

	for i := 0; i < numBlocks; i++ {
		// Simulate current block's gas usage
		gasUsedNum, gasUsedDenom := simulateGasUsage(currentGas)

		// Get new parameters
		newGas, newAvgNum, newAvgDenom := ga.AdjustGas(
			gasUsedNum, gasUsedDenom,
			avgRatioNum, avgRatioDenom,
		)

		// Store metrics
		metrics = append(metrics, blockMetrics{
			gas:      newGas,
			avgRatio: float64(newAvgNum) / float64(newAvgDenom),
		})

		// Update for next block
		currentGas = newGas
		avgRatioNum = newAvgNum
		avgRatioDenom = newAvgDenom
	}

	// Analyze results

	// 1. Verify gas stays within bounds
	for i, m := range metrics {
		if m.gas < minGas || m.gas > maxGas {
			t.Errorf("Gas out of bounds at block %d: %v", i, m.gas)
		}
	}

	// 2. Verify no extreme oscillations in gas limit
	for i := 1; i < len(metrics); i++ {
		change := float64(metrics[i].gas) / float64(metrics[i-1].gas)
		if change < 0.8 || change > 1.2 { // Allow max 20% change
			t.Errorf("Too large gas change at block %d: %v -> %v", i, metrics[i-1].gas, metrics[i].gas)
		}
	}

	// 3. Verify system stability in last 100 blocks
	var gasVariance float64
	avgGas := 0.0
	for i := len(metrics) - 100; i < len(metrics); i++ {
		avgGas += float64(metrics[i].gas)
	}
	avgGas /= 100

	for i := len(metrics) - 100; i < len(metrics); i++ {
		diff := float64(metrics[i].gas) - avgGas
		gasVariance += (diff * diff) / avgGas / avgGas
	}
	gasVariance /= 100

	if gasVariance > 0.01 { // Allow 1% variance
		t.Errorf("System not stable in final blocks. Variance: %v", gasVariance)
	}

}
