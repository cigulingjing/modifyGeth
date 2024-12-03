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
// parentEMAGas 也改为整数，实际值需要除以 PRECISION
const GAS_PRECISION uint64 = 1000000 // 精度为 6 位小数

func (ga *GasAdaptor) AdjustGas(
	blockGasNumerator, blockGasDenominator uint64,
	parentEMAGasNumerator, parentEMAGasDenominator uint64,
) (newGas uint64, newAvgNumerator uint64, newAvgDenominator uint64) {
	// Check if denominators are zero when numerators are not zero
	if blockGasNumerator != 0 && blockGasDenominator == 0 {
		panic("blockGasDenominator cannot be zero when blockGasNumerator is not zero")
	}
	// if blockGasNumerator > blockGasDenominator {
	// 	panic("blockGasNumerator must be less than or equal to blockGasDenominator")
	// }
	if parentEMAGasNumerator != 0 && parentEMAGasDenominator == 0 {
		panic("parentEMAGasDenominator cannot be zero when parentEMAGasNumerator is not zero")
	}

	// 使用整数计算 EMA
	// newEMA = (alpha * blockGas + (1-alpha) * parentEMA)
	// = (alphaNumerator * blockGas * PRECISION + (alphaDenominator-alphaNumerator) * parentEMAGas) / alphaDenominator

	// 先计算 blockGas 项
	blockGasTerm := new(big.Int).SetUint64(ga.alphaNumerator)
	blockGasTerm.Mul(blockGasTerm, new(big.Int).SetUint64(blockGasNumerator))
	blockGasTerm.Mul(blockGasTerm, new(big.Int).SetUint64(GAS_PRECISION))
	if blockGasDenominator == 0 {
		blockGasDenominator = 1
	}
	blockGasTerm.Div(blockGasTerm, new(big.Int).SetUint64(blockGasDenominator))

	// 再计算 parentEMAGas 项
	parentTerm := new(big.Int).SetUint64(ga.alphaDenominator - ga.alphaNumerator)
	parentTerm.Mul(parentTerm, new(big.Int).SetUint64(parentEMAGasNumerator))
	parentTerm.Mul(parentTerm, new(big.Int).SetUint64(GAS_PRECISION))
	if parentEMAGasDenominator == 0 {
		parentEMAGasDenominator = 1
	}
	parentTerm.Div(parentTerm, new(big.Int).SetUint64(parentEMAGasDenominator))

	// 合并两项并除以 alphaDenominator
	newEMAGas := new(big.Int).Add(blockGasTerm, parentTerm)

	// 设置新的平均值分子
	newAvgNumerator = newEMAGas.Uint64()
	// 设置新的平均值分母
	newAvgDenominator = ga.alphaDenominator * GAS_PRECISION

	// 计算新的 Gas 值
	newGas = newEMAGas.Div(newEMAGas, new(big.Int).SetUint64(GAS_PRECISION)).Uint64()

	// 限制 Gas 在上下限之间
	if newGas < ga.minGas {
		newGas = ga.minGas
	} else if newGas > ga.maxGas {
		newGas = ga.maxGas
	}

	ga.currentGas = newGas
	return newGas, newAvgNumerator, newAvgDenominator
}

// GetCurrentGas 返回当前 PoW Gas
func (ga *GasAdaptor) GetCurrentGas() uint64 {
	return ga.currentGas
}
