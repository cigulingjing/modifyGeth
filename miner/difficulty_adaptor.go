package miner

import (
	"math"
	"math/big"
)

// DifficultyAdaptor 用于调整挖矿难度
type DifficultyAdaptor struct {
	targetPowRatio    float64
	alpha             float64
	fMin              float64
	fMax              float64
	currentAvgRatio   float64
	currentDifficulty *big.Int
}

// NewDifficultyAdaptor 创建一个新的 DifficultyAdaptor 实例
func NewDifficultyAdaptor(targetPowRatio, alpha, fMin, fMax float64, initialDifficulty *big.Int, initialAvgRatio float64) *DifficultyAdaptor {
	return &DifficultyAdaptor{
		targetPowRatio:    targetPowRatio,
		alpha:             alpha,
		fMin:              fMin,
		fMax:              fMax,
		currentAvgRatio:   initialAvgRatio,
		currentDifficulty: new(big.Int).Set(initialDifficulty),
	}
}

// AdjustDifficulty 根据当前区块的 PoW 交易比例调整难度
func (da *DifficultyAdaptor) AdjustDifficulty(currentPowRatio float64) *big.Int {
	// Update average PoW transaction ratio
	da.currentAvgRatio = da.alpha*currentPowRatio + (1-da.alpha)*da.currentAvgRatio

	// Calculate adjustment factor
	f := da.currentAvgRatio / da.targetPowRatio

	// Limit adjustment factor range
	fLimited := math.Max(math.Min(f, da.fMax), da.fMin)

	// Calculate new difficulty
	fBig := new(big.Float).SetFloat64(fLimited)
	newDifficultyFloat := new(big.Float).Mul(new(big.Float).SetInt(da.currentDifficulty), fBig)

	newDifficulty := new(big.Int)
	newDifficultyFloat.Int(newDifficulty)

	// Update current difficulty
	da.currentDifficulty.Set(newDifficulty)

	return da.currentDifficulty
}

// GetCurrentDifficulty 返回当前难度
func (da *DifficultyAdaptor) GetCurrentDifficulty() *big.Int {
	return new(big.Int).Set(da.currentDifficulty)
}

// GetCurrentAvgRatio 返回当前平均 PoW 交易比例
func (da *DifficultyAdaptor) GetCurrentAvgRatio() float64 {
	return da.currentAvgRatio
}
