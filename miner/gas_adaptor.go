package miner

// GasAdaptor 用于调整 PoW Gas
type GasAdaptor struct {
	minGas     uint64
	maxGas     uint64
	currentGas uint64
	emaGas     float64
	alpha      float64
}

// NewGasAdaptor 创建一个新的 GasAdaptor 实例
func NewGasAdaptor(minGas, maxGas, initialGas uint64, alpha float64) *GasAdaptor {
	return &GasAdaptor{
		minGas:     minGas,
		maxGas:     maxGas,
		currentGas: initialGas,
		emaGas:     float64(initialGas),
		alpha:      alpha,
	}
}

// UpdateGas 更新 EMA 并调整 Gas
// 每次传入平均Gas
func (ga *GasAdaptor) AdjustGas(blockGas uint64) uint64 {
	// 更新 EMA
	ga.emaGas = ga.alpha*float64(blockGas) + (1-ga.alpha)*ga.emaGas

	// 计算新的 Gas 值
	newGas := uint64(ga.emaGas)

	// 限制 Gas 在上下限之间
	if newGas < ga.minGas {
		newGas = ga.minGas
	} else if newGas > ga.maxGas {
		newGas = ga.maxGas
	}

	// 更新当前 Gas
	ga.currentGas = newGas
	return newGas
}

// GetCurrentGas 返回当前 PoW Gas
func (ga *GasAdaptor) GetCurrentGas() uint64 {
	return ga.currentGas
}

// GetEMAGas 返回当前 EMA Gas 值
func (ga *GasAdaptor) GetEMAGas() float64 {
	return ga.emaGas
}
