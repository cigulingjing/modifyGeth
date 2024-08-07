package vm

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

var errPanguAdd = errors.New("error pangu add : input length must be 64 bytes")

type panguAdd struct{}

type panguCallData struct{}

func (p *panguAdd) RequiredGas(input []byte) uint64 {
	// 自定义Gas计算方法
	// Input为 tx msg 中的 data，如果需要按操作计算Gas，需要自行解析
	return 10
}

func (p *panguAdd) Run(input []byte, blkCtx BlockContext) ([]byte, error) {
	if len(input) != 64 {
		return nil, errPanguAdd
	}

	// 读取两个 uint256 数字
	a := new(uint256.Int).SetBytes(input[:32])
	b := new(uint256.Int).SetBytes(input[32:])

	fmt.Println("a:", a)
	fmt.Println("b:", b)
	// 计算和
	sum := new(uint256.Int).Add(a, b)
	fmt.Println("sum:", sum)
	fmt.Println("sum.Bytes():", sum.Bytes())
	return sum.Bytes(), nil

}

func (p *panguCallData) RequiredGas(input []byte) uint64 {
	// 自定义Gas计算方法
	// Input为 tx msg 中的 data，如果需要按操作计算Gas，需要自行解析
	return 100
}

func (p *panguCallData) Run(input []byte, blkCtx BlockContext) ([]byte, error) {
	if len(input) != 32 {
		return nil, errPanguAdd
	}
	fmt.Println("input:", input)
	txhash := common.BytesToHash(input[:32])
	fmt.Println("txhash:", txhash)
	// read calldata from txhash
	_, tx, _, _, _, err := blkCtx.BlockChainStateRead.GetTransaction(blkCtx.Rpcctx, txhash)
	if err != nil {
		return nil, err
	}

	callData := tx.Data()

	return callData, nil
}
