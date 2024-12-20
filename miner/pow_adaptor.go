package miner

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PoWAdaptor 用于调整 PoW 相关参数
type PoWAdaptor struct {
	targetPowRatio *common.Rational
	alpha          *common.Rational

	// 难度调整参数
	fMin              *common.Rational
	fMax              *common.Rational
	currentDifficulty *big.Int

	// 价格调整参数
	kp               *common.Rational
	ki               *common.Rational
	minPrice         *big.Int
	maxPrice         *big.Int
	accumulatedError *common.Rational
}

// NewPoWAdaptor 创建一个新的 PoWAdaptor 实例
func NewPoWAdaptor(
	targetPowRatio *common.Rational,
	alpha *common.Rational,
	fMin *common.Rational,
	fMax *common.Rational,
	initialDifficulty *big.Int,
	kp *common.Rational,
	ki *common.Rational,
	minPrice, maxPrice *big.Int,
) *PoWAdaptor {
	if targetPowRatio.Compare(common.NewRational(1, 1)) > 0 {
		panic("targetPowRatio must be less than or equal to 1")
	}
	if minPrice.Cmp(maxPrice) > 0 {
		panic("minPrice must be less than or equal to maxPrice")
	}
	if initialDifficulty.Sign() <= 0 {
		panic("initialDifficulty must be positive")
	}
	if minPrice.Sign() <= 0 {
		panic("minPrice must be positive")
	}

	return &PoWAdaptor{
		targetPowRatio:    targetPowRatio,
		alpha:             alpha,
		fMin:              fMin,
		fMax:              fMax,
		currentDifficulty: initialDifficulty,
		kp:                kp,
		ki:                ki,
		minPrice:          minPrice,
		maxPrice:          maxPrice,
		accumulatedError:  common.NewRational(0, 1),
	}
}

// AdjustParameters 调整 PoW 参数
func (pa *PoWAdaptor) AdjustParameters(
	currentPowRatio *common.Rational,
	parentAvgRatio *common.Rational,
	parentPrice *big.Int,
) (*big.Int, *big.Int, uint64, uint64) {
	// Update EMA
	oneMinusAlpha := common.NewRational(1, 1).Sub(pa.alpha)
	newAvgRatio := currentPowRatio.Mul(pa.alpha).Add(parentAvgRatio.Mul(oneMinusAlpha))

	// Calculate error
	error := pa.targetPowRatio.Sub(newAvgRatio)
	pa.accumulatedError = pa.accumulatedError.Add(error)

	// Adjust difficulty
	var newDiff *common.Rational
	currentDiff_Rational := common.NewRational(pa.currentDifficulty.Uint64(), 1)
	if error.Compare(common.NewRational(0, 1)) > 0 {
		newDiff = currentDiff_Rational.Mul(pa.fMin)
	} else if error.Compare(common.NewRational(0, 1)) < 0 {
		newDiff = currentDiff_Rational.Mul(pa.fMax)
	} else {
		newDiff = currentDiff_Rational
	}
	pa.currentDifficulty = newDiff.ToBigInt()

	// Adjust price using PID control
	parentPriceRational := common.FromBigInt(parentPrice)
	proportionalTerm := error.Mul(pa.kp)
	integralTerm := pa.accumulatedError.Mul(pa.ki)
	priceAdjustment := proportionalTerm.Add(integralTerm)
	newPrice := parentPriceRational.Add(priceAdjustment).ToBigInt()

	// Ensure price stays within bounds
	if newPrice.Cmp(pa.minPrice) < 0 {
		newPrice = pa.minPrice
	} else if newPrice.Cmp(pa.maxPrice) > 0 {
		newPrice = pa.maxPrice
	}

	return newDiff.ToBigInt(), newPrice, newAvgRatio.Numerator, newAvgRatio.Denominator
}

// GetCurrentDifficulty returns the current difficulty
func (pa *PoWAdaptor) GetCurrentDifficulty() *big.Int {
	return pa.currentDifficulty
}

// GetAccumulatedError returns the accumulated error
func (pa *PoWAdaptor) GetAccumulatedError() *common.Rational {
	return pa.accumulatedError
}

// ResetAccumulatedError resets the accumulated error to zero
func (pa *PoWAdaptor) ResetAccumulatedError() {
	pa.accumulatedError = common.NewRational(0, 1)
}

func RationalFromBigInt(b *big.Int) *common.Rational {
	return common.NewRational(b.Uint64(), 1)
}
