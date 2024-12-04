// * This file is to parse abi/bytecode from hardhat's artifacts output

package abi

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func LoadHardhatContract(abiFileName string) (*ABI, string, string) {
	data, err := os.ReadFile(abiFileName)
	if err != nil {
		fmt.Printf("Error reading file: %v", err)
		return nil, "", ""
	}
	// Using map to store content
	var contract map[string]interface{}
	err = json.Unmarshal(data, &contract)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, "", ""
	}
	// Parse abi
	contractABI, err := loadContractJSON(contract)
	if err != nil {
		fmt.Printf("Err in load file %s:%s", abiFileName, err)
		return nil, "", ""
	}
	// Parse bytecode, convert to string
	contractBytecode := contract["bytecode"].(string)
	deployedBytecode := contract["deployedBytecode"].(string)
	return contractABI, contractBytecode, deployedBytecode
}

func loadContractJSON(contract map[string]interface{}) (*ABI, error) {
	// Parse abi from data.The abi field in json is an array, So convert it to []interface{}
	abidata, ok := contract["abi"].([]interface{})
	if !ok {
		fmt.Println("ABI is not a string")
	}
	abiJSON, err := json.Marshal(abidata)
	if err != nil {
		fmt.Println("Error Marshal json:", err)
	}

	parsedABI, err := JSON(strings.NewReader(string(abiJSON)))
	if err != nil {
		fmt.Printf("Error parsing ABI: %v", err)
	}
	return &parsedABI, nil
}
