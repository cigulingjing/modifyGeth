package voucher

// Used to solve some problems caused by abi coding
import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// Method declaration
	BalanceOf     Method
	Buy           Method
	Use           Method
	CreateVoucher Method
	RootAddress   = &common.Address{}
	gas           = uint64(100000000)
)

type VoucherInfos struct {
	OriginToken   common.Address
	MinAmount     *big.Int
	ExchangeRate  *big.Int
	BuyExpiration *big.Int
	UseExpiration *big.Int
}

func LoadContract(abiFileName string) (*abi.ABI, string) {
	data, err := os.ReadFile(abiFileName)
	if err != nil {
		fmt.Printf("Error reading file: %v", err)
		return nil, ""
	}
	// Using map to store content
	var contract map[string]interface{}
	err = json.Unmarshal(data, &contract)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, ""
	}
	// Parse abi
	contractABI, err := loadContractJSON(contract)
	if err != nil {
		fmt.Printf("Err in load file %s:%s", abiFileName, err)
		return nil, ""
	}
	// Parse bytecode, convert to string
	contractBytecode := contract["bytecode"].(string)
	// Initialize contract method
	BalanceOf = NewMethod(contractABI, "balanceOf", true, gas)
	Use = NewMethod(contractABI, "use", false, gas)
	Buy = NewMethod(contractABI, "buy", false, gas)
	CreateVoucher = NewMethod(contractABI, "createVoucher", false, gas)

	return contractABI, contractBytecode
}

func loadContractJSON(contract map[string]interface{}) (*abi.ABI, error) {
	// Parse abi from data.The abi field in json is an array, So convert it to []interface{}
	abidata, ok := contract["abi"].([]interface{})
	if !ok {
		fmt.Println("ABI is not a string")
	}
	abiJSON, err := json.Marshal(abidata)
	if err != nil {
		fmt.Println("Error Marshal json:", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		log.Fatalf("Error parsing ABI: %v", err)
	}
	return &parsedABI, nil
}

// Convert address to an array of length 32.
// func AddressToBytes(address common.Address) []byte {
// 	return common.LeftPadBytes(address.Bytes(), 32)
// }

// func BigIntToBytes(amount *big.Int) []byte {
// 	return common.LeftPadBytes(amount.Bytes(), 32)
// }
