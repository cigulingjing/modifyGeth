package cryptoupgrade

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO relative path manager
var CodeStorageABI, _, _ = abi.LoadHardhatContract("/home/ubuntu/project/modifyGeth/cryptoupgrade/contract/CodeStorage.json")

// * parse event from receipt
func ParseReceipt(receipt *types.Receipt) {
	eventSignature := []byte("pullcode(string)")
	eventHash := crypto.Keccak256Hash(eventSignature)

	for _, vLog := range receipt.Logs {
		fmt.Println(vLog)
		if vLog.Topics[0] == eventHash {
			fmt.Println("Event found in transaction logs!")
			// Event in codestorage.sol
			type pullcode struct {
				name string
			}
			var event pullcode
			err := CodeStorageABI.UnpackIntoInterface(&event, "pullcode", vLog.Data)
			if err != nil {
				fmt.Printf("Failed to unpack log data: %v", err)
			}
			fmt.Printf("Event data: Name=%s\n", event.name)
		}
	}
}

func BindPullcode(client *ethclient.Client) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{CodeStorageAddress},
		Topics:    [][]common.Hash{{PullCodeEventHash}}, // Event hash
	}
	logCh := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logCh)
	if err != nil {
		log.Error("Failed to subscribe to logs: %v", err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Info("Error while listening for logs: %v", err)
		case Log := <-logCh:
			// Parse name from event
			fmt.Println("Catch pull code event!")
			var name string
			err = CodeStorageABI.UnpackIntoInterface(&name, "pullcode", Log.Data)
			if err != nil {
				log.Error("Decode log data err!")
			}
			compressedCode := LookupCode(client, name)

			workPath,_:=os.Executable()
			log.Info("work path:",workPath)
			outputPath := "./cryptoupgrade/bin/src/" + name + ".go"
			err = decompressStringToFile(compressedCode, outputPath)
			if err != nil {
				log.Info(fmt.Sprintf("Decompressed algorithm %s to %s", name, outputPath))
			} else {
				log.Error(fmt.Sprintf("Error in decompressed algorithm %s to %s", name, outputPath))
			}
		}
	}

}

// * Through Client call contract, get the code of @name algorithm
func LookupCode(client *ethclient.Client, name string) string {
	if name == "" {
		log.Error("Err in parse name from pullcode event")
	}
	fmt.Printf("name: %v\n", name)
	
	lookupFuncName := "getCode"
	compressedCode := ""
	input, err := CodeStorageABI.Pack(lookupFuncName, name)
	if err != nil {
		log.Error(fmt.Sprintf("Call contract codeStorage err! Algorithm:%s", name))
	}

	msg := ethereum.CallMsg{
		To:   &CodeStorageAddress,
		Data: input,
	}
	output, err := client.CallContract(context.Background(), msg, nil)
	fmt.Printf("output: %v\n", output)
	if err != nil {
		log.Error(fmt.Sprintf("failed to call contract: %v", err))
	}

	err = CodeStorageABI.UnpackIntoInterface(&compressedCode, lookupFuncName, output)
	if err != nil {
		log.Error("Failed to unpack ouput of %s method in codeStorage contract", lookupFuncName)
	}
	fmt.Printf("compressedCode: %v\n", compressedCode)
	return compressedCode
}
