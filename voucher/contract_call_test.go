package voucher

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
)

func TestABILoad(t *testing.T) {
	file_name := "./abi/mutivoucher_payable.json"
	exmABI,exmBytecode := LoadContract(file_name)
	fmt.Printf("ABI: %v\n", exmABI)
	fmt.Printf("bytecode: %v\n", exmBytecode)
	// Construct input
	value := big.NewInt(10000)
	result, _ := exmABI.Pack("", value)
	hexResult := hex.EncodeToString(result)
	fmt.Printf("hexResult: %v\n", hexResult)
}
