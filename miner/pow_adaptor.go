package miner

import (
	"fmt"
	"math/big"
)

// PoWAdaptor 用于调整 PoW 相关参数
type PoWAdaptor struct {
	// EMA 相关参数
	targetPowRatioNumerator   uint64
	targetPowRatioDenominator uint64
	alphaNumerator            uint64
	alphaDenominator          uint64

	// 难度调整参数
	fMinNumerator     uint64
	fMinDenominator   uint64
	fMaxNumerator     uint64
	fMaxDenominator   uint64
	currentDifficulty *big.Int

	// 价格调整参数
	kpNumerator   uint64
	kpDenominator uint64
	kiNumerator   uint64
	kiDenominator uint64
	minPrice      *big.Int
	maxPrice      *big.Int
	// Store accumulated error as fraction
	accumulatedErrorNumerator   *big.Int
	accumulatedErrorDenominator *big.Int
}

// NewPoWAdaptor 创建一个新的 PoWAdaptor 实例
func NewPoWAdaptor(
	targetPowRatioNumerator, targetPowRatioDenominator uint64,
	alphaNumerator, alphaDenominator uint64,
	fMinNumerator, fMinDenominator uint64,
	fMaxNumerator, fMaxDenominator uint64,
	initialDifficulty *big.Int,
	kpNumerator, kpDenominator uint64,
	kiNumerator, kiDenominator uint64,
	minPrice, maxPrice *big.Int,
) *PoWAdaptor {
	// Check if denominators are zero when numerators are not zero
	if alphaNumerator != 0 && alphaDenominator == 0 {
		panic("alphaDenominator cannot be zero when alphaNumerator is not zero")
	}
	if targetPowRatioNumerator != 0 && targetPowRatioDenominator == 0 {
		panic("targetPowRatioDenominator cannot be zero when targetPowRatioNumerator is not zero")
	}
	if fMinNumerator != 0 && fMinDenominator == 0 {
		panic("fMinDenominator cannot be zero when fMinNumerator is not zero")
	}
	if fMaxNumerator != 0 && fMaxDenominator == 0 {
		panic("fMaxDenominator cannot be zero when fMaxNumerator is not zero")
	}
	if kpNumerator != 0 && kpDenominator == 0 {
		panic("kpDenominator cannot be zero when kpNumerator is not zero")
	}
	if kiNumerator != 0 && kiDenominator == 0 {
		panic("kiDenominator cannot be zero when kiNumerator is not zero")
	}

	if targetPowRatioNumerator > targetPowRatioDenominator {
		panic("targetPowRatioNumerator must be less than or equal to targetPowRatioDenominator")
	}

	// Check if prices are valid
	if minPrice.Sign() < 0 || maxPrice.Sign() < 0 {
		panic("prices cannot be negative")
	}
	if minPrice.Cmp(maxPrice) > 0 {
		panic("minPrice must be less than or equal to maxPrice")
	}

	// Check if initial difficulty is valid
	if initialDifficulty.Sign() <= 0 {
		panic("initialDifficulty must be positive")
	}
	return &PoWAdaptor{
		targetPowRatioNumerator:     targetPowRatioNumerator,
		targetPowRatioDenominator:   targetPowRatioDenominator,
		alphaNumerator:              alphaNumerator,
		alphaDenominator:            alphaDenominator,
		fMinNumerator:               fMinNumerator,
		fMinDenominator:             fMinDenominator,
		fMaxNumerator:               fMaxNumerator,
		fMaxDenominator:             fMaxDenominator,
		currentDifficulty:           new(big.Int).Set(initialDifficulty),
		kpNumerator:                 kpNumerator,
		kpDenominator:               kpDenominator,
		kiNumerator:                 kiNumerator,
		kiDenominator:               kiDenominator,
		minPrice:                    new(big.Int).Set(minPrice),
		maxPrice:                    new(big.Int).Set(maxPrice),
		accumulatedErrorNumerator:   big.NewInt(0),
		accumulatedErrorDenominator: big.NewInt(1),
	}
}

// AdjustParameters 同时调整难度和价格
func (pa *PoWAdaptor) AdjustParameters(
	currentPowRatioNumerator, currentPowRatioDenominator uint64,
	parentAvgRatioNumerator, parentAvgRatioDenominator uint64,
	parentPrice *big.Int,
) (*big.Int, *big.Int, uint64, uint64) {
	// Check if input ratios are valid
	if currentPowRatioNumerator != 0 && currentPowRatioDenominator == 0 {
		panic("currentPowRatioDenominator cannot be zero when currentPowRatioNumerator is not zero")
	}
	if currentPowRatioNumerator > currentPowRatioDenominator {
		panic("currentPowRatioNumerator must be less than or equal to currentPowRatioDenominator")
	}
	if parentAvgRatioNumerator != 0 && parentAvgRatioDenominator == 0 {
		panic("parentAvgRatioDenominator cannot be zero when parentAvgRatioNumerator is not zero")
	}
	if parentAvgRatioNumerator > parentAvgRatioDenominator {
		panic("parentAvgRatioNumerator must be less than or equal to parentAvgRatioDenominator")
	}
	if parentPrice == nil {
		panic("parentPrice cannot be nil")
	}
	if parentPrice.Sign() < 0 {
		panic("parentPrice cannot be negative")
	}

	// First term: alpha * currentRatio
	// = (alphaNumerator * currentRatioNumerator) / (alphaDenominator * currentRatioDenominator)
	term1Num := new(big.Int).Mul(
		big.NewInt(0).SetUint64(pa.alphaNumerator),
		big.NewInt(0).SetUint64(currentPowRatioNumerator),
	)

	term1Denom := new(big.Int).Mul(
		big.NewInt(0).SetUint64(pa.alphaDenominator),
		big.NewInt(0).SetUint64(currentPowRatioDenominator),
	)
	// Second term: (1-alpha) * parentAvgRatio
	// = ((alphaDenominator-alphaNumerator) * parentAvgRatioNumerator) / (alphaDenominator * parentAvgRatioDenominator)
	term2Num := new(big.Int).Mul(
		big.NewInt(0).SetUint64(pa.alphaDenominator-pa.alphaNumerator),
		big.NewInt(0).SetUint64(parentAvgRatioNumerator),
	)
	term2Denom := new(big.Int).Mul(
		big.NewInt(0).SetUint64(pa.alphaDenominator),
		big.NewInt(0).SetUint64(parentAvgRatioDenominator),
	)

	// Reduce each fraction first
	term1Num, term1Denom = reduceFraction(term1Num, term1Denom)
	term2Num, term2Denom = reduceFraction(term2Num, term2Denom)

	// Find least common multiple of denominators
	commonDenom := lcm(term1Denom, term2Denom)

	// Calculate multipliers for each term
	term1Multiplier := new(big.Int).Div(commonDenom, term1Denom)
	term2Multiplier := new(big.Int).Div(commonDenom, term2Denom)

	// Convert terms to common denominator using minimal multiplication
	term1NumConverted := new(big.Int).Mul(term1Num, term1Multiplier)
	term2NumConverted := new(big.Int).Mul(term2Num, term2Multiplier)

	// Sum up numerator
	newAvgRatioNumerator := new(big.Int).Add(term1NumConverted, term2NumConverted)
	newAvgRatioDenominator := commonDenom

	// Reduce the final fraction
	reducedNum, reducedDenom := reduceFraction(newAvgRatioNumerator, newAvgRatioDenominator)
	if reducedNum.Cmp(reducedDenom) > 0 {
		fmt.Println("reducedNum > reducedDenom")
	}

	// Calculate new difficulty using the reduced ratio
	newDifficulty := pa.calculateNewDifficulty(reducedNum, reducedDenom)

	// Calculate new price
	newPrice := pa.calculateNewPrice(currentPowRatioNumerator, currentPowRatioDenominator, parentPrice)

	// Scale down the values if they're too large for uint64
	scaledNum, scaledDenom := scaleDownBigInts(reducedNum, reducedDenom)

	return newDifficulty, newPrice, scaledNum, scaledDenom
}

// calculateNewDifficulty 计算新的难度值
func (pa *PoWAdaptor) calculateNewDifficulty(newAvgRatioNumerator, newAvgRatioDenominator *big.Int) *big.Int {
	// Calculate avgRatioForDiff = newAvgRatio/alphaDenominator
	// = newAvgRatioNumerator/(newAvgRatioDenominator * alphaDenominator)
	avgRatioForDiffNumerator := new(big.Int).Set(newAvgRatioNumerator)
	avgRatioForDiffDenominator := new(big.Int).Mul(
		newAvgRatioDenominator,
		new(big.Int).SetUint64(pa.alphaDenominator),
	)

	// Calculate targetRatio = targetPowRatioNumerator/targetPowRatioDenominator
	targetRatioNumerator := new(big.Int).SetUint64(pa.targetPowRatioNumerator)
	targetRatioDenominator := new(big.Int).SetUint64(pa.targetPowRatioDenominator)

	// Calculate adjustmentFactor = avgRatioForDiff/targetRatio
	// = (avgRatioForDiffNumerator * targetRatioDenominator)/(avgRatioForDiffDenominator * targetRatioNumerator)
	adjustmentFactorNumerator := new(big.Int).Mul(avgRatioForDiffNumerator, targetRatioDenominator)
	adjustmentFactorDenominator := new(big.Int).Mul(avgRatioForDiffDenominator, targetRatioNumerator)

	// Calculate min/max factors as fractions
	minFactorNumerator := new(big.Int).SetUint64(pa.fMinNumerator)
	minFactorDenominator := new(big.Int).SetUint64(pa.fMinDenominator)

	maxFactorNumerator := new(big.Int).SetUint64(pa.fMaxNumerator)
	maxFactorDenominator := new(big.Int).SetUint64(pa.fMaxDenominator)

	// Compare fractions: a/b < c/d equivalent to a*d < c*b
	minComparison := new(big.Int).Mul(adjustmentFactorNumerator, minFactorDenominator).Cmp(
		new(big.Int).Mul(minFactorNumerator, adjustmentFactorDenominator))

	maxComparison := new(big.Int).Mul(adjustmentFactorNumerator, maxFactorDenominator).Cmp(
		new(big.Int).Mul(maxFactorNumerator, adjustmentFactorDenominator))

	if minComparison < 0 {
		adjustmentFactorNumerator = minFactorNumerator
		adjustmentFactorDenominator = minFactorDenominator
	} else if maxComparison > 0 {
		adjustmentFactorNumerator = maxFactorNumerator
		adjustmentFactorDenominator = maxFactorDenominator
	}

	// Calculate new difficulty = currentDifficulty * adjustmentFactor
	// Only perform the actual division at the end
	newDifficulty := new(big.Int).Mul(pa.currentDifficulty, adjustmentFactorNumerator)
	newDifficulty.Div(newDifficulty, adjustmentFactorDenominator)

	// Update current difficulty
	pa.currentDifficulty.Set(newDifficulty)

	return newDifficulty
}

// calculateNewPrice 计算新的价格
func (pa *PoWAdaptor) calculateNewPrice(
	currentRatioNumerator, currentRatioDenominator uint64,
	parentPrice *big.Int,
) *big.Int {
	// Calculate error = targetRatio - currentRatio
	// = (targetRatioNumerator*currentRatioDenominator - currentRatioNumerator*targetRatioDenominator) / (targetRatioDenominator*currentRatioDenominator)
	errorNumerator := new(big.Int).Mul(
		big.NewInt(int64(pa.targetPowRatioNumerator)),
		big.NewInt(int64(currentRatioDenominator)),
	)
	errorNumerator.Sub(
		errorNumerator,
		new(big.Int).Mul(
			big.NewInt(int64(currentRatioNumerator)),
			big.NewInt(int64(pa.targetPowRatioDenominator)),
		),
	)
	errorDenominator := new(big.Int).SetUint64(pa.targetPowRatioDenominator * currentRatioDenominator)

	// Reduce error fraction
	errorNumerator, errorDenominator = reduceFraction(errorNumerator, errorDenominator)

	// Update accumulated error as fraction
	newAccErrorNumerator := new(big.Int).Mul(pa.accumulatedErrorNumerator, errorDenominator)
	newAccErrorNumerator.Add(
		newAccErrorNumerator,
		new(big.Int).Mul(errorNumerator, pa.accumulatedErrorDenominator),
	)
	newAccErrorDenominator := new(big.Int).Mul(pa.accumulatedErrorDenominator, errorDenominator)

	// Reduce accumulated error fraction
	newAccErrorNumerator, newAccErrorDenominator = reduceFraction(newAccErrorNumerator, newAccErrorDenominator)

	pa.accumulatedErrorNumerator = newAccErrorNumerator
	pa.accumulatedErrorDenominator = newAccErrorDenominator

	// Calculate price adjustment
	term1 := new(big.Int).Mul(errorNumerator, big.NewInt(int64(pa.kpNumerator)))
	term1.Mul(term1, big.NewInt(int64(pa.kiDenominator)))

	term2 := new(big.Int).Mul(pa.accumulatedErrorNumerator, big.NewInt(int64(pa.kiNumerator)))
	term2.Mul(term2, big.NewInt(int64(pa.kpDenominator)))
	term2.Mul(term2, errorDenominator)

	totalAdjustmentNumerator := new(big.Int).Add(term1, term2)
	totalAdjustmentDenominator := new(big.Int).SetUint64(pa.kpDenominator * pa.kiDenominator)
	totalAdjustmentDenominator.Mul(totalAdjustmentDenominator, errorDenominator)
	totalAdjustmentDenominator.Mul(totalAdjustmentDenominator, pa.accumulatedErrorDenominator)

	// Reduce total adjustment fraction
	totalAdjustmentNumerator, totalAdjustmentDenominator = reduceFraction(totalAdjustmentNumerator, totalAdjustmentDenominator)

	// Calculate total adjustment
	totalAdjustment := new(big.Int).Div(totalAdjustmentNumerator, totalAdjustmentDenominator)

	// Calculate new price
	newPrice := new(big.Int).Add(parentPrice, totalAdjustment)

	// Limit price within min and max bounds
	if newPrice.Cmp(pa.minPrice) < 0 {
		newPrice.Set(pa.minPrice)
	} else if newPrice.Cmp(pa.maxPrice) > 0 {
		newPrice.Set(pa.maxPrice)
	}

	return newPrice
}

// 添加一些辅助方法
func (pa *PoWAdaptor) GetCurrentDifficulty() *big.Int {
	return new(big.Int).Set(pa.currentDifficulty)
}

func (pa *PoWAdaptor) GetAccumulatedError() (*big.Int, *big.Int) {
	return reduceFraction(
		new(big.Int).Set(pa.accumulatedErrorNumerator),
		new(big.Int).Set(pa.accumulatedErrorDenominator),
	)
}

func (pa *PoWAdaptor) ResetAccumulatedError() {
	pa.accumulatedErrorNumerator = big.NewInt(0)
	pa.accumulatedErrorDenominator = big.NewInt(1)
}
