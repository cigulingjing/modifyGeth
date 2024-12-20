package miner

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// Test GasAdaptor

func TestGasAdaptorNormal(t *testing.T) {
	// Initialize with normal values
	ga := NewGasAdaptor(1000, 10000, 5000, common.NewRational(3, 10))

	// Test normal adjustment
	currentRatio := common.NewRational(6000, 5000)   // Simulating 1.2 ratio
	parentAvgRatio := common.NewRational(5000, 5000) // Simulating 1.0 ratio

	newGas, newAvgRatio, newAvgDenom := ga.AdjustGas(currentRatio, parentAvgRatio)

	// Check if results are within expected range
	if newGas < 1000 || newGas > 10000 {
		t.Errorf("newGas out of range: got %d, want between %d and %d", newGas, 1000, 10000)
	}

	if newAvgRatio == 0 || newAvgDenom == 0 {
		t.Error("average values should not be zero")
	}
}

func TestGasAdaptorEdgeCases(t *testing.T) {
	// Test min gas limit
	ga := NewGasAdaptor(1000, 10000, 5000, common.NewRational(3, 10))
	newGas, _, _ := ga.AdjustGas(
		common.NewRational(100, 1000),  // Very low ratio
		common.NewRational(1000, 1000), // Normal ratio
	)
	if newGas != 1000 {
		t.Errorf("expected min gas %d, got %d", 1000, newGas)
	}

	// Test max gas limit
	newGas, _, _ = ga.AdjustGas(
		common.NewRational(20000, 1000), // Very high ratio
		common.NewRational(1000, 1000),  // Normal ratio
	)
	if newGas != 10000 {
		t.Errorf("expected max gas %d, got %d", 10000, newGas)
	}
}

// Test PoWAdaptor
func TestPoWAdaptorNormal(t *testing.T) {
	initialDiff := big.NewInt(100)
	minPrice := big.NewInt(100)
	maxPrice := big.NewInt(10000)

	pa := NewPoWAdaptor(
		common.NewRational(3, 10),  // targetPowRatio
		common.NewRational(2, 10),  // alpha
		common.NewRational(8, 10),  // fMin
		common.NewRational(12, 10), // fMax
		initialDiff,
		common.NewRational(1, 10),  // kp
		common.NewRational(1, 100), // ki
		minPrice,
		maxPrice,
	)

	// Test normal adjustment
	currentPowRatio := common.NewRational(4, 10) // Current PoW ratio
	parentAvgRatio := common.NewRational(3, 10)  // Parent average ratio
	parentPrice := big.NewInt(1000)              // Parent price

	newDiff, newPrice, newAvgRatio, newAvgDenom := pa.AdjustParameters(
		currentPowRatio,
		parentAvgRatio,
		parentPrice,
	)

	if newDiff.Cmp(big.NewInt(0)) <= 0 {
		t.Error("new difficulty should be positive")
	}

	if newPrice.Cmp(minPrice) < 0 || newPrice.Cmp(maxPrice) > 0 {
		t.Error("new price out of valid range")
	}

	if newAvgRatio == 0 || newAvgDenom == 0 {
		t.Error("average values should not be zero")
	}
}

func TestPoWAdaptorEdgeCases(t *testing.T) {
	initialDiff := big.NewInt(100)
	minPrice := big.NewInt(100)
	maxPrice := big.NewInt(10000)

	pa := NewPoWAdaptor(
		common.NewRational(3, 10),  // targetPowRatio
		common.NewRational(2, 10),  // alpha
		common.NewRational(8, 10),  // fMin
		common.NewRational(12, 10), // fMax
		initialDiff,
		common.NewRational(1, 10),  // kp
		common.NewRational(1, 100), // ki
		minPrice,
		maxPrice,
	)

	// Test min price boundary
	parentPrice := big.NewInt(100)
	_, newPrice, _, _ := pa.AdjustParameters(
		common.NewRational(1, 10), // very low ratio to push price down
		common.NewRational(1, 10),
		parentPrice,
	)

	if newPrice.Cmp(minPrice) < 0 {
		t.Error("price should not go below minimum")
	}

	// Test max price boundary
	parentPrice = big.NewInt(9000)
	_, newPrice, _, _ = pa.AdjustParameters(
		common.NewRational(9, 10), // very high ratio to push price up
		common.NewRational(9, 10),
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
			name: "zero price",
			fn: func() {
				NewPoWAdaptor(
					common.NewRational(3, 10),  // targetPowRatio
					common.NewRational(2, 10),  // alpha
					common.NewRational(8, 10),  // fMin
					common.NewRational(12, 10), // fMax
					big.NewInt(100),            // initialDiff
					common.NewRational(1, 10),  // kp
					common.NewRational(1, 100), // ki
					big.NewInt(0),              // zero minPrice
					big.NewInt(10000),          // maxPrice
				)
			},
		},
		{
			name: "min price greater than max price",
			fn: func() {
				NewPoWAdaptor(
					common.NewRational(3, 10),  // targetPowRatio
					common.NewRational(2, 10),  // alpha
					common.NewRational(8, 10),  // fMin
					common.NewRational(12, 10), // fMax
					big.NewInt(100),            // initialDiff
					common.NewRational(1, 10),  // kp
					common.NewRational(1, 100), // ki
					big.NewInt(100000),         // minPrice > maxPrice
					big.NewInt(100),            // maxPrice
				)
			},
		},
		{
			name: "zero initial difficulty",
			fn: func() {
				NewPoWAdaptor(
					common.NewRational(3, 10),  // targetPowRatio
					common.NewRational(2, 10),  // alpha
					common.NewRational(8, 10),  // fMin
					common.NewRational(12, 10), // fMax
					big.NewInt(0),              // zero initialDiff
					common.NewRational(1, 10),  // kp
					common.NewRational(1, 100), // ki
					big.NewInt(100),            // minPrice
					big.NewInt(100000),         // maxPrice
				)
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
	initialDiff := big.NewInt(100)
	minPrice := big.NewInt(100)
	maxPrice := big.NewInt(10000)

	pa := NewPoWAdaptor(
		common.NewRational(3, 10),  // targetPowRatio = 0.3
		common.NewRational(2, 10),  // alpha = 0.2
		common.NewRational(8, 10),  // fMin = 0.8
		common.NewRational(12, 10), // fMax = 1.2
		initialDiff,
		common.NewRational(1, 10),  // kp = 0.1
		common.NewRational(1, 100), // ki = 0.01
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
	currentPrice := big.NewInt(0).Add(
		minPrice,
		maxPrice,
	)
	currentPrice.Div(currentPrice, big.NewInt(2))
	avgRatio := common.NewRational(3, 10)

	// Simulate 1000 blocks
	numBlocks := 1000

	// Helper to calculate simulated PoW ratio based on difficulty
	simulatePowRatio := func(diff *big.Int, price *big.Int) *common.Rational {
		// Simulate miners responding to difficulty/price
		// Higher difficulty -> lower ratio
		// Higher price -> higher ratio
		baseRatio := 0.3 // Target ratio
		diffEffect := -0.1 * float64(diff.Uint64()) / float64(initialDiff.Uint64())
		priceEffect := 0.1 * float64(price.Uint64()) / float64(maxPrice.Uint64())

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

		return common.NewRational(uint64(ratio*1000), 1000)
	}

	for i := 0; i < numBlocks; i++ {
		// Simulate current block's PoW ratio
		currentRatio := simulatePowRatio(currentDiff, currentPrice)

		// Get new parameters
		newDiff, newPrice, newAvgRatio, newAvgDenom := pa.AdjustParameters(
			currentRatio,
			avgRatio,
			currentPrice,
		)

		// Store metrics
		metrics = append(metrics, blockMetrics{
			difficulty: newDiff,
			price:      newPrice,
			powRatio:   float64(currentRatio.Numerator) / float64(currentRatio.Denominator),
			avgRatio:   float64(newAvgRatio) / float64(newAvgDenom),
		})

		// Update for next block
		currentDiff = newDiff
		currentPrice = newPrice
		avgRatio = common.NewRational(newAvgRatio, newAvgDenom)
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
		maxChange = new(big.Int).Add(maxChange, big.NewInt(1))
		if diff.Cmp(maxChange) > 0 {
			t.Errorf("Too large difficulty change at block %d", i)
		}
	}

	// 4. Verify system stability in last 100 blocks
	var ratioVariance float64
	avgRatioValue := 0.0
	for i := len(metrics) - 100; i < len(metrics); i++ {
		avgRatioValue += metrics[i].powRatio
	}
	avgRatioValue /= 100

	for i := len(metrics) - 100; i < len(metrics); i++ {
		diff := metrics[i].powRatio - avgRatioValue
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
		common.NewRational(2, 10), // alpha = 0.2 (EMA weight)
	)

	type blockMetrics struct {
		gas      uint64
		avgRatio float64
	}

	// Store metrics for analysis
	metrics := make([]blockMetrics, 0)

	// Initial values
	currentGas := initialGas
	avgRatio := common.NewRational(15000000, 15000000) // Initial target

	// Simulate 1000 blocks
	numBlocks := 1000

	// Helper to calculate simulated gas usage based on current gas limit
	simulateGasUsage := func(gasLimit uint64) *common.Rational {
		// Simulate blocks using around 50% of gas limit with some random variation
		baseUsage := float64(gasLimit) * 0.5
		variation := float64(gasLimit) * 0.2 * (rand.Float64() - 0.5)
		usage := uint64(baseUsage + variation)

		if usage > gasLimit {
			usage = gasLimit
		}

		return common.NewRational(usage, gasLimit)
	}

	for i := 0; i < numBlocks; i++ {
		// Simulate current block's gas usage
		currentRatio := simulateGasUsage(currentGas)

		// Get new parameters
		newGas, newAvgRatio, newAvgDenom := ga.AdjustGas(
			currentRatio,
			avgRatio,
		)

		// Store metrics
		metrics = append(metrics, blockMetrics{
			gas:      newGas,
			avgRatio: float64(newAvgRatio) / float64(newAvgDenom),
		})

		// Update for next block
		currentGas = newGas
		avgRatio = common.NewRational(newAvgRatio, newAvgDenom)
	}

	// Analyze results

	// 1. Verify gas stays within bounds
	for i, m := range metrics {
		if m.gas < minGas || m.gas > maxGas {
			t.Errorf("Gas out of bounds at block %d: %v", i, m.gas)
		}
	}

	// 2. Verify no extreme oscillations
	for i := 1; i < len(metrics); i++ {
		change := float64(metrics[i].gas) / float64(metrics[i-1].gas)
		if change > 1.2 || change < 0.8 {
			t.Errorf("Too large gas change at block %d: %v", i, change)
		}
	}

	// 3. Verify system stability in last 100 blocks
	var gasVariance float64
	avgGas := uint64(0)
	for i := len(metrics) - 100; i < len(metrics); i++ {
		avgGas += metrics[i].gas
	}
	avgGas /= 100

	for i := len(metrics) - 100; i < len(metrics); i++ {
		diff := float64(metrics[i].gas) - float64(avgGas)
		gasVariance += diff * diff
	}
	gasVariance /= float64(avgGas * avgGas * 100)

	if gasVariance > 0.01 {
		t.Errorf("System not stable in final blocks. Variance: %v", gasVariance)
	}
}
