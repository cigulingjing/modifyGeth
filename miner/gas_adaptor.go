package miner

import (
	"github.com/ethereum/go-ethereum/common"
)

// GasAdaptor 用于调整 PoW Gas
type GasAdaptor struct {
	minGas     uint64
	maxGas     uint64
	alpha      *common.Rational
	currentGas uint64
}

// NewGasAdaptor 创建一个新的 GasAdaptor 实例
func NewGasAdaptor(minGas, maxGas, initialGas uint64, alpha *common.Rational) *GasAdaptor {
	if minGas >= maxGas {
		panic("minGas must be less than maxGas")
	}
	if initialGas < minGas || initialGas > maxGas {
		panic("initialGas must be between minGas and maxGas")
	}
	return &GasAdaptor{
		minGas:     minGas,
		maxGas:     maxGas,
		alpha:      alpha,
		currentGas: initialGas,
	}
}

// AdjustGas 更新 EMA 并调整 Gas
func (ga *GasAdaptor) AdjustGas(
	blockGasRatio *common.Rational,
	parentEMAGasRatio *common.Rational,
) (uint64, uint64, uint64) {
	// Update EMA
	oneMinusAlpha := common.NewRational(1, 1).Sub(ga.alpha)
	newEMAGasRatio := blockGasRatio.Mul(ga.alpha).Add(parentEMAGasRatio.Mul(oneMinusAlpha))

	// Adjust gas limit based on EMA
	if newEMAGasRatio.Compare(common.NewRational(8, 10)) > 0 {
		ga.currentGas = min(ga.currentGas+ga.currentGas/10, ga.maxGas)
	} else if newEMAGasRatio.Compare(common.NewRational(5, 10)) < 0 {
		ga.currentGas = max(ga.currentGas-ga.currentGas/10, ga.minGas)
	}

	return ga.currentGas, newEMAGasRatio.Numerator, newEMAGasRatio.Denominator
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// GetCurrentGas 返回当前 PoW Gas
func (ga *GasAdaptor) GetCurrentGas() uint64 {
	return ga.currentGas
}
