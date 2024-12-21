package miner

import (
	"fmt"

	"github.com/ethereum/go-ethereum/log"
)

// offChain calcuation will push the result into executor's offChain channel
func (e *executor) offchainCom(data []byte) (result [][]byte) {
	if len(data) >= 32*3 {
		//
		param1 := data[0:32]
		param2 := data[32:64]
		param3 := data[64:96]
		result = append(result, param1, param2, param3)
		// indicate the result is writed into stateDB
		e.offChainCh <- true
		return
	} else {
		log.Error("the transaction belong WASH ,but the input len is illegal")
	}
	return
}

// 1.catch the result of calcuation 2.make other operators
// @attention: transaction(which call offChainCalc) and transaction(call this fun) must in the same block
func (e *executor) offchainResultCatch(env *executor_env) {
	// catch result from channel
	result := <-e.offChainCh
	if result {
		fmt.Printf("get WASM result\n")
	} else {
		fmt.Printf("The offchainCh in executor is empty\n")
	}
	//TODO other operators after catch the result
}
