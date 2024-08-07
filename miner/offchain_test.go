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
	"github.com/ethereum/go-ethereum/rlp"
)

// global variable
var (
	e *executor
	b *testWorkerBackend
	p *poter
)

func init() {
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
	// when all of txs in txPool are executed, seal a block
	errs := b.txPool.Add([]*types.Transaction{txSend, txCommon, txGet}, true, false)
	fmt.Printf("errs: %v\n", errs)
	// wait for consense
	time.Sleep(20 * time.Second)
	// catch result from executor environment
	state := e.env.state
	result := state.OffChainResult
	fmt.Printf("result in stateDB:%v\n", result)

	//  txGet2 will be executed in the second block, leading congestion of exuctor
	txGet2 := types.MustSignNewTx(testBankKey, signer, &types.AccessListTx{
		ChainID:  b.chain.Config().ChainID,
		Nonce:    3,
		To:       &testUserAddress,
		Value:    big.NewInt(1000),
		Gas:      30000,
		GasPrice: big.NewInt(params.InitialBaseFee),
		Data:     dataHeader,
	})
	// test the channel in second block whether is empty
	errs = b.txPool.Add([]*types.Transaction{txGet2}, true, false)
	fmt.Printf("errs: %v\n", errs)
	time.Sleep(20 * time.Second)
}

func HeaderJsonPrint(header *types.Header) {
	bs, _ := json.Marshal(header)
	var out bytes.Buffer
	json.Indent(&out, bs, "", "\t")
	fmt.Printf("The json of block: %v\n", out.String())
}

// Add attribute @incentive  to block header, test whether affect the system
func TestHeaderRlp(t *testing.T) {
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
	// Get the latest Block
	latestHeader := b.chain.CurrentBlock()

	// test json decode
	// HeaderJsonPrint(latestHeader)

	// test RLP decode
	rlpEncodedHeader, err := rlp.EncodeToBytes(latestHeader)
	if err != nil {
		t.Fatalf("Failed to RLP encode block header: %v", err)
	}
	var decodedHeader types.Header
	err = rlp.DecodeBytes(rlpEncodedHeader, &decodedHeader)
	if err != nil {
		t.Fatalf("Failed to RLP decode block header: %v", err)
	}
	HeaderJsonPrint(&decodedHeader)
}
