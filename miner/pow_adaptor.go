package miner

import (
	"math/big"
)

const RATIO_PRECISION = 1000000 // 精度为 6 位小数

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
	kpNumerator      uint64
	kpDenominator    uint64
	kiNumerator      uint64
	kiDenominator    uint64
	minPrice         *big.Int
	maxPrice         *big.Int
	accumulatedError int64
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
	return &PoWAdaptor{
		targetPowRatioNumerator:   targetPowRatioNumerator,
		targetPowRatioDenominator: targetPowRatioDenominator,
		alphaNumerator:            alphaNumerator,
		alphaDenominator:          alphaDenominator,
		fMinNumerator:             fMinNumerator,
		fMinDenominator:           fMinDenominator,
		fMaxNumerator:             fMaxNumerator,
		fMaxDenominator:           fMaxDenominator,
		currentDifficulty:         new(big.Int).Set(initialDifficulty),
		kpNumerator:               kpNumerator,
		kpDenominator:             kpDenominator,
		kiNumerator:               kiNumerator,
		kiDenominator:             kiDenominator,
		minPrice:                  new(big.Int).Set(minPrice),
		maxPrice:                  new(big.Int).Set(maxPrice),
		accumulatedError:          0,
	}
}

// AdjustParameters 同时调整难度和价格
func (pa *PoWAdaptor) AdjustParameters(
	currentPowRatioNumerator, currentPowRatioDenominator uint64,
	parentAvgRatioNumerator, parentAvgRatioDenominator uint64,
	parentPrice *big.Int,
) (*big.Int, *big.Int, uint64, uint64) {
	// 计算新的平均比率
	currentRatio := new(big.Int).Mul(
		big.NewInt(int64(currentPowRatioNumerator)),
		big.NewInt(RATIO_PRECISION),
	)
	currentRatio.Div(currentRatio, big.NewInt(int64(currentPowRatioDenominator)))

	parentRatio := new(big.Int).Mul(
		big.NewInt(int64(parentAvgRatioNumerator)),
		big.NewInt(RATIO_PRECISION),
	)
	parentRatio.Div(parentRatio, big.NewInt(int64(parentAvgRatioDenominator)))

	// 计算新的 EMA
	newAvgRatio := new(big.Int).Mul(
		big.NewInt(int64(pa.alphaNumerator)),
		currentRatio,
	)
	newAvgRatio.Mul(newAvgRatio, big.NewInt(RATIO_PRECISION))

	parentTerm := new(big.Int).Mul(
		big.NewInt(int64(pa.alphaDenominator-pa.alphaNumerator)),
		parentRatio,
	)

	newAvgRatio.Add(newAvgRatio, parentTerm)

	// 保存新的平均比率
	newAvgRatioNumerator := newAvgRatio.Uint64()
	newAvgRatioDenominator := pa.alphaDenominator * RATIO_PRECISION

	// 计算新难度
	newDifficulty := pa.calculateNewDifficulty(newAvgRatio)

	// 计算新价格
	newPrice := pa.calculateNewPrice(currentRatio, parentPrice)

	return newDifficulty, newPrice, newAvgRatioNumerator, newAvgRatioDenominator
}

// calculateNewDifficulty 计算新的难度值
func (pa *PoWAdaptor) calculateNewDifficulty(newAvgRatio *big.Int) *big.Int {
	// 继续计算新的难度值
	avgRatioForDiff := new(big.Int).Div(newAvgRatio, big.NewInt(int64(pa.alphaDenominator)))

	// 计算调整因子 f = newAvgRatio / targetPowRatio
	targetRatio := new(big.Int).Mul(
		big.NewInt(int64(pa.targetPowRatioNumerator)),
		big.NewInt(RATIO_PRECISION),
	)
	targetRatio.Div(targetRatio, big.NewInt(int64(pa.targetPowRatioDenominator)))

	adjustmentFactor := new(big.Int).Mul(
		avgRatioForDiff,
		big.NewInt(RATIO_PRECISION),
	)
	adjustmentFactor.Div(adjustmentFactor, targetRatio)

	// 限制调整因子范围
	minFactor := new(big.Int).Mul(
		big.NewInt(int64(pa.fMinNumerator)),
		big.NewInt(RATIO_PRECISION),
	)
	minFactor.Div(minFactor, big.NewInt(int64(pa.fMinDenominator)))

	maxFactor := new(big.Int).Mul(
		big.NewInt(int64(pa.fMaxNumerator)),
		big.NewInt(RATIO_PRECISION),
	)
	maxFactor.Div(maxFactor, big.NewInt(int64(pa.fMaxDenominator)))

	if adjustmentFactor.Cmp(minFactor) < 0 {
		adjustmentFactor.Set(minFactor)
	} else if adjustmentFactor.Cmp(maxFactor) > 0 {
		adjustmentFactor.Set(maxFactor)
	}

	// 计算新难度
	newDifficulty := new(big.Int).Mul(pa.currentDifficulty, adjustmentFactor)
	newDifficulty.Div(newDifficulty, big.NewInt(RATIO_PRECISION))

	// 更新当前难度
	pa.currentDifficulty.Set(newDifficulty)

	return newDifficulty
}

// calculateNewPrice 计算新的价格
func (pa *PoWAdaptor) calculateNewPrice(currentRatio *big.Int, parentPrice *big.Int) *big.Int {
	// 计算目标比率
	targetRatio := new(big.Int).Mul(
		big.NewInt(int64(pa.targetPowRatioNumerator)),
		big.NewInt(RATIO_PRECISION),
	)
	targetRatio.Div(targetRatio, big.NewInt(int64(pa.targetPowRatioDenominator)))

	// 计算误差 error = targetPowRatio - currentRatio
	error := targetRatio.Int64() - currentRatio.Int64()
	pa.accumulatedError += error

	// 计算价格调整
	// adjustment = kp*error + ki*accumulatedError
	// = (kpNumerator*error*RATIO_PRECISION/kpDenominator + kiNumerator*accumulatedError*RATIO_PRECISION/kiDenominator)

	kpAdjustment := new(big.Int).Mul(
		big.NewInt(error),
		big.NewInt(int64(pa.kpNumerator)),
	)
	kpAdjustment.Mul(kpAdjustment, big.NewInt(RATIO_PRECISION))
	kpAdjustment.Div(kpAdjustment, big.NewInt(int64(pa.kpDenominator)))

	kiAdjustment := new(big.Int).Mul(
		big.NewInt(pa.accumulatedError),
		big.NewInt(int64(pa.kiNumerator)),
	)
	kiAdjustment.Mul(kiAdjustment, big.NewInt(RATIO_PRECISION))
	kiAdjustment.Div(kiAdjustment, big.NewInt(int64(pa.kiDenominator)))

	// 总调整量
	totalAdjustment := new(big.Int).Add(kpAdjustment, kiAdjustment)

	// 计算新价格
	newPrice := new(big.Int).Add(parentPrice, totalAdjustment)
	newPrice.Div(newPrice, big.NewInt(RATIO_PRECISION))

	// 限制价格范围
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

func (pa *PoWAdaptor) GetAccumulatedError() int64 {
	return pa.accumulatedError
}

func (pa *PoWAdaptor) ResetAccumulatedError() {
	pa.accumulatedError = 0
}
