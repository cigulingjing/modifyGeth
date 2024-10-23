package miner

import (
	"math/big"
)

// PriceAdaptor 用于调整 PoW 价格
type PriceAdaptor struct {
	targetPowRatio   float64
	kp               float64
	ki               float64
	minPrice         *big.Int
	maxPrice         *big.Int
	currentPrice     *big.Int
	accumulatedError float64
}

// NewPriceAdaptor 创建一个新的 PriceAdaptor 实例
func NewPriceAdaptor(targetPowRatio, kp, ki float64, minPrice, maxPrice, initialPrice *big.Int) *PriceAdaptor {
	return &PriceAdaptor{
		targetPowRatio:   targetPowRatio,
		kp:               kp,
		ki:               ki,
		minPrice:         new(big.Int).Set(minPrice),
		maxPrice:         new(big.Int).Set(maxPrice),
		currentPrice:     new(big.Int).Set(initialPrice),
		accumulatedError: 0,
	}
}

// AdjustPrice 根据当前区块的 PoW 交易比例调整价格
func (pa *PriceAdaptor) AdjustPrice(currentPowRatio float64) *big.Int {
	// Calculate error
	error := pa.targetPowRatio - currentPowRatio
	pa.accumulatedError += error

	// Calculate price adjustment
	adjustment := pa.kp*error + pa.ki*pa.accumulatedError

	// Convert current price to float for calculation
	currentPriceFloat := new(big.Float).SetInt(pa.currentPrice)
	adjustmentFloat := new(big.Float).SetFloat64(adjustment)

	// Calculate new price
	newPriceFloat := new(big.Float).Add(currentPriceFloat, adjustmentFloat)

	// Convert new price back to big.Int
	newPrice := new(big.Int)
	newPriceFloat.Int(newPrice)

	// Limit price range
	if newPrice.Cmp(pa.minPrice) < 0 {
		newPrice.Set(pa.minPrice)
	} else if newPrice.Cmp(pa.maxPrice) > 0 {
		newPrice.Set(pa.maxPrice)
	}

	// Update current price
	pa.currentPrice.Set(newPrice)

	return pa.currentPrice
}

// GetCurrentPrice 返回当前 PoW 价格
func (pa *PriceAdaptor) GetCurrentPrice() *big.Int {
	return new(big.Int).Set(pa.currentPrice)
}

// ResetAccumulatedError 重置累积误差
func (pa *PriceAdaptor) ResetAccumulatedError() {
	pa.accumulatedError = 0
}
