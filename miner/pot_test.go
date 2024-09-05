package miner

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

var keystoreFile = "../build/bin/chain/node1/keystore/UTC--2024-07-16T03-48-59.278859185Z--0d77d7a0769eb6a70bc08c7d40c10aab052b8c4c"

var password = "123456"

var addr = "0x0d77d7A0769eB6a70bc08C7D40C10AaB052b8c4c"

var httpUrl = "http://localhost:39764"

// func TestSignTx(t *testing.T) {
// 	// 1. 读取keystore文件
// 	ks, err := os.ReadFile(keystoreFile)
// 	if err != nil {
// 		fmt.Println("ReadFile failed, err:", err)
// 		t.Error("NewKeyStore failed")
// 		return
// 	}

// 	fromKey, err := keystore.DecryptKey(ks, password)
// 	require.NoError(t, err)

// 	fromPrivkey := fromKey.PrivateKey
// 	fmt.Println(fromPrivkey)

// 	fromPublickey := fromKey.PrivateKey.Public()
// 	fmt.Println(fromPublickey)

// 	fromAddr := crypto.PubkeyToAddress(*fromPublickey.(*ecdsa.PublicKey))
// 	fmt.Println(fromAddr)

// 	// 2. 连接以太坊节点
// 	client, err := ethclient.Dial(httpUrl)
// 	if err != nil {
// 		fmt.Println("Dial failed, err:", err)
// 		t.Error("Dial failed")
// 		return
// 	}

// 	// 3. 构造交易
// 	toAddr := common.HexToAddress(addr)
// 	amount := big.NewInt(100)
// 	var gasPrice *big.Int = big.NewInt(200)
// 	var gasLimit uint64 = 300000
// 	nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
// 	if err != nil {
// 		fmt.Println("PendingNonceAt failed, err:", err)
// 		t.Error("PendingNonceAt failed")
// 		return
// 	}
// 	t.Log("nonce:", nonce)

// 	tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, gasPrice, []byte{})

// 	// 4. 签名交易
// 	signTx, err := client.SendTransaction()
// }

func TestMarshalTx(t *testing.T) {
	rawHexTx := "0xf8690164825208940d77d7a0769eb6a70bc08c7d40c10aab052b8c4c64850d02010203850539ed4b71a0c99fd3906b4a37acac8bbefe4569bdf9de2a29118954f42f69c0c87178e08761a00ac3f63790bf65a6c3ab7479f36fc8d0bcd67ba52397ab160fcb954f417e8575"
	tx := new(types.Transaction)
	rawBytesTx, _ := hexutil.Decode(rawHexTx)
	fmt.Println(rawBytesTx)
	tx.UnmarshalBinary(rawBytesTx)
	fmt.Println(tx)
	log.Info("tx", "tx:", tx)
	fmt.Println("111")
	// err := tx.UnmarshalJSON(rawBytesTx)
	// if err != nil {
	// 	t.Error("UnmarshalJSON failed")
	// 	return
	// }
	// fmt.Println(tx)
}
