package miner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

//global variable
var(
	e *executor 
	b *testWorkerBackend
	p *poter
)
func init(){
	var (
		db     = rawdb.NewMemoryDatabase()
		config = *params.AllCliqueProtocolChanges
	)
	config.Clique = &params.CliqueConfig{Period: 1, Epoch: 30000}
	engine := clique.New(config.Clique, db)
	// initialization of the client and the consensus client
	e, b = newTestExecutor(&config, engine, db, 0)
	p = newTestPotExecutor(&config, engine, db)
}
func TestOffchain(t *testing.T) {
	e.start()
	defer e.close()
	p.start()
	defer p.close()
	// construct the first transaction,attention the nonce of the transaction
	signer := types.LatestSigner(b.chain.Config())
	paramString := strings.Repeat("OHkSr95hmMK7CrCl5jerQllimbglRYrG", 3)
	paramBytes := []byte(paramString)
	dataHeader := []byte{0x0A, 0x0D, 0x01}
	dataHeader = append(dataHeader, paramBytes...)
	txSend := types.MustSignNewTx(testBankKey, signer, &types.AccessListTx{
		ChainID:  b.chain.Config().ChainID,
		Nonce:    0,
		To:       &testUserAddress,
		Value:    big.NewInt(1000),
		Gas:      30000,
		GasPrice: big.NewInt(params.InitialBaseFee),
		Data:     dataHeader,
	})
	// common ethereum transaction
	txCommon := types.MustSignNewTx(testBankKey, signer, &types.AccessListTx{
		ChainID:  b.chain.Config().ChainID,
		To:       &testUserAddress,
		Value:    big.NewInt(1000),
		Nonce:    1,
		Gas:      30000,
		GasPrice: big.NewInt(params.InitialBaseFee),
	})
	// construct the second transaction:read result of first offchain's transaction
	dataHeader = []byte{0x0A, 0x0D, 0x02}
	txGet := types.MustSignNewTx(testBankKey, signer, &types.AccessListTx{
		ChainID:  b.chain.Config().ChainID,
		Nonce:    2,
		To:       &testUserAddress,
		Value:    big.NewInt(1000),
		Gas:      30000,
		GasPrice: big.NewInt(params.InitialBaseFee),
		Data:     dataHeader,
	})

	errs := b.txPool.Add([]*types.Transaction{txSend, txCommon, txGet}, true, false)
	fmt.Printf("errs: %v\n", errs)
	//wait for consense 
	time.Sleep(30 * time.Second)
	//catch result from executor environment 
	state := e.env.state
	result := state.OffChainResult
	fmt.Printf("result in executor:%v\n", result)
}

// Add attribute @incentive  to block header, test whether affect the system 
func TestHeaderRlp(t *testing.T){
	e.start()
	defer e.close()
	p.start()
	defer p.close()
	// common ethereum transaction
	signer := types.LatestSigner(b.chain.Config())
	tx := types.MustSignNewTx(testBankKey, signer, &types.AccessListTx{
		ChainID:  b.chain.Config().ChainID,
		To:       &testUserAddress,
		Value:    big.NewInt(1000),
		Nonce:    0,
		Gas:      30000,
		GasPrice: big.NewInt(params.InitialBaseFee),
	})

	errs := b.txPool.Add([]*types.Transaction{tx}, true, false)
	fmt.Printf("errs: %v\n", errs)
	// wait for consense 
	time.Sleep(30 * time.Second)
	//Print as Json
	latestBlock:=b.chain.CurrentBlock()
	bs,_:=json.Marshal(latestBlock)
	var out bytes.Buffer
	json.Indent(&out,bs,"","\t")
	fmt.Printf("latestBlock: %v\n", out.String())
}