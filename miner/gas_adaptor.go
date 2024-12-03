package miner

import (
	"math/big"
)

// GasAdaptor 用于调整 PoW Gas
type GasAdaptor struct {
	minGas     uint64
	maxGas     uint64
	currentGas uint64
	// alpha 改为分子和分母，比如 alpha = 3/10
	alphaNumerator   uint64
	alphaDenominator uint64
}

// NewGasAdaptor 创建一个新的 GasAdaptor 实例
func NewGasAdaptor(minGas, maxGas, initialGas uint64, alphaNumerator, alphaDenominator uint64) *GasAdaptor {
	// Check if denominator is zero when numerator is not zero
	if alphaNumerator != 0 && alphaDenominator == 0 {
		panic("alphaDenominator cannot be zero when alphaNumerator is not zero")
	}

	return &GasAdaptor{
		minGas:           minGas,
		maxGas:           maxGas,
		currentGas:       initialGas,
		alphaNumerator:   alphaNumerator,
		alphaDenominator: alphaDenominator,
	}
}

// AdjustGas 更新 EMA 并调整 Gas
func (ga *GasAdaptor) AdjustGas(
	blockGasNumerator, blockGasDenominator uint64,
	parentEMAGasNumerator, parentEMAGasDenominator uint64,
) (newGas uint64, newAvgNumerator uint64, newAvgDenominator uint64) {
	// Check if denominators are zero when numerators are not zero
	if blockGasNumerator != 0 && blockGasDenominator == 0 {
		panic("blockGasDenominator cannot be zero when blockGasNumerator is not zero")
	}
	if parentEMAGasNumerator != 0 && parentEMAGasDenominator == 0 {
		panic("parentEMAGasDenominator cannot be zero when parentEMAGasNumerator is not zero")
	}

	// First term: alpha * blockGas
	// = (alphaNumerator * blockGasNumerator) / (alphaDenominator * blockGasDenominator)
	term1Num := new(big.Int).Mul(
		big.NewInt(0).SetUint64(ga.alphaNumerator),
		big.NewInt(0).SetUint64(blockGasNumerator),
	)
	term1Denom := new(big.Int).Mul(
		big.NewInt(0).SetUint64(ga.alphaDenominator),
		big.NewInt(0).SetUint64(blockGasDenominator),
	)

	// Second term: (1-alpha) * parentEMAGas
	// = ((alphaDenominator-alphaNumerator) * parentEMAGasNumerator) / (alphaDenominator * parentEMAGasDenominator)
	term2Num := new(big.Int).Mul(
		big.NewInt(0).SetUint64(ga.alphaDenominator-ga.alphaNumerator),
		big.NewInt(0).SetUint64(parentEMAGasNumerator),
	)
	term2Denom := new(big.Int).Mul(
		big.NewInt(0).SetUint64(ga.alphaDenominator),
		big.NewInt(0).SetUint64(parentEMAGasDenominator),
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

	// Calculate new Gas value
	newGas = new(big.Int).Div(reducedNum, reducedDenom).Uint64()

	// Limit Gas within min and max bounds
	if newGas < ga.minGas {
		newGas = ga.minGas
	} else if newGas > ga.maxGas {
		newGas = ga.maxGas
	}

	ga.currentGas = newGas

	// Scale down the values if they're too large for uint64
	scaledNum, scaledDenom := scaleDownBigInts(reducedNum, reducedDenom)
	return newGas, scaledNum, scaledDenom
}

// GetCurrentGas 返回当前 PoW Gas
func (ga *GasAdaptor) GetCurrentGas() uint64 {
	return ga.currentGas
}
