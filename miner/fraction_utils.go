package miner

import (
	"math/big"
)

// gcd calculates the Greatest Common Divisor of two big.Int numbers
func gcd(a, b *big.Int) *big.Int {
	if b.Sign() == 0 {
		return new(big.Int).Abs(a)
	}
	return gcd(b, new(big.Int).Rem(a, b))
}

// reduceFraction reduces a fraction to its lowest terms using GCD
func reduceFraction(numerator, denominator *big.Int) (*big.Int, *big.Int) {
	if denominator.Sign() == 0 {
		panic("denominator cannot be zero")
	}

	// Handle the case where numerator is 0
	if numerator.Sign() == 0 {
		return big.NewInt(0), big.NewInt(1)
	}

	// Calculate GCD
	gcdVal := gcd(new(big.Int).Set(numerator), new(big.Int).Set(denominator))

	// Reduce the fraction
	newNumerator := new(big.Int).Div(numerator, gcdVal)
	newDenominator := new(big.Int).Div(denominator, gcdVal)

	// Ensure denominator is positive
	if denominator.Sign() < 0 {
		newNumerator.Neg(newNumerator)
		newDenominator.Neg(newDenominator)
	}

	return newNumerator, newDenominator
}

// lcm calculates the Least Common Multiple of two big.Int numbers
func lcm(a, b *big.Int) *big.Int {
	if a.Sign() == 0 || b.Sign() == 0 {
		return big.NewInt(0)
	}
	gcdVal := gcd(new(big.Int).Set(a), new(big.Int).Set(b))
	// lcm = |a * b| / gcd(a, b)
	result := new(big.Int).Mul(a, b)
	result.Abs(result)
	return result.Div(result, gcdVal)
}

// scaleDownBigInts 通过移位操作缩小数值，保持比例不变
func scaleDownBigInts(num, denom *big.Int) (uint64, uint64) {
	// Create copies to avoid modifying original values
	n := new(big.Int).Set(num)
	d := new(big.Int).Set(denom)

	// While either number is too large for uint64
	for n.BitLen() > 63 || d.BitLen() > 63 {
		// Right shift both numbers by 1 to maintain ratio
		n.Rsh(n, 1)
		d.Rsh(d, 1)

		// Ensure denominator doesn't become zero
		if d.Sign() == 0 {
			d.SetInt64(1)
		}
	}

	return n.Uint64(), d.Uint64()
}
