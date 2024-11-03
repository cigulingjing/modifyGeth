package tests

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/clique"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/txpool/legacypool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/miner"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/voucher"
	"github.com/holiman/uint256"
	// "google.golang.org/grpc"
)

const weiPerEth = 1e18

// Global variable
var (
	// Create bank account
	bankKey, _  = crypto.GenerateKey()
	bankAddress = crypto.PubkeyToAddress(bankKey.PublicKey)
	// Create user account
	userKey, _  = crypto.GenerateKey()
	userAddress = crypto.PubkeyToAddress(userKey.PublicKey)
	// Contract setting, parse from json which is created by hardhat
	contractAbi, contractBytecode = voucher.LoadContract("../voucher/abi/mutivoucher.json")
)

// Implements the interface of miner.Backend
type TestBackend struct {
	bc     *core.BlockChain
	txPool *txpool.TxPool
}

func (m *TestBackend) NetworkId() uint64            { return 0 }
func (m *TestBackend) BlockChain() *core.BlockChain { return m.bc }
func (m *TestBackend) TxPool() *txpool.TxPool       { return m.txPool }

// Gensis block
func genesisBlock(period uint64, gasLimit uint64, faucet common.Address) *core.Genesis {
	config := *params.AllCliqueProtocolChanges
	config.Clique = &params.CliqueConfig{
		Period: period,
		Epoch:  config.Clique.Epoch,
	}

	// Assemble and return the genesis with the precompiles and faucet pre-funded
	return &core.Genesis{
		Config:     &config,
		ExtraData:  append(append(make([]byte, 32), faucet[:]...), make([]byte, crypto.SignatureLength)...),
		GasLimit:   gasLimit,
		BaseFee:    big.NewInt(params.InitialBaseFee),
		Difficulty: big.NewInt(1),
		Alloc: map[common.Address]core.GenesisAccount{
			faucet:      {Balance: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))},
			userAddress: {Balance: big.NewInt(1000000)},
		},
	}
}

// Create new blockchain to test
func newBlockChain() *core.BlockChain {
	database := rawdb.NewMemoryDatabase()
	triedb := trie.NewDatabase(database, nil)
	genesis := genesisBlock(15, 111500000000, bankAddress)
	chainConfig, _, err := core.SetupGenesisBlock(database, triedb, genesis)
	if err != nil {
		fmt.Printf("can't create new chain config: %v\n", err)
	}
	engine := clique.New(chainConfig.Clique, database)
	// Create blockchain
	bc, err := core.NewBlockChain(database, nil, genesis, nil, engine, vm.Config{}, nil, nil)
	if err != nil {
		fmt.Printf("can't create new chain %v\n", err)
	}
	return bc
}

// Create new txpool basing blockchain
func newTxPool(bc *core.BlockChain) *txpool.TxPool {
	testTxPoolConfig := legacypool.DefaultConfig
	pool := legacypool.New(testTxPoolConfig, bc)
	txpool, _ := txpool.New(new(big.Int).SetUint64(testTxPoolConfig.PriceLimit), bc, []txpool.SubPool{pool})
	return txpool
}

func newEVM(bc *core.BlockChain) *vm.EVM {
	// Create stateDB
	header := bc.CurrentBlock()
	stateDB, err := bc.StateAt(header.Root)
	if err != nil {
		fmt.Printf("failed to get stateDB: %v\n", err)
	}
	// Create EVM
	blockContext := core.NewEVMBlockContext(header, bc, nil)
	evm := vm.NewEVM(blockContext, vm.TxContext{}, stateDB, bc.Config(), vm.Config{})
	return evm
}

// Constrcut function to depoly contract
func newDeployContractTx(bc *core.BlockChain, Nonce int) *types.Transaction {
	signer := types.LatestSigner(bc.Config())
	tx0 := types.MustSignNewTx(bankKey, signer, &types.AccessListTx{
		ChainID:  bc.Config().ChainID,
		Nonce:    uint64(Nonce),
		Gas:      1000000000,
		GasPrice: big.NewInt(params.InitialBaseFee),
		Data:     common.FromHex(contractBytecode),
	})
	return tx0
}

func newTransaction(bc *core.BlockChain, Nonce int, data []byte) *types.Transaction {
	// Construct tx.Data
	fmt.Printf("data: %s\n", hex.EncodeToString(data))
	signer := types.LatestSigner(bc.Config())
	tx := types.MustSignNewTx(bankKey, signer, &types.AccessListTx{
		ChainID:  bc.Config().ChainID,
		Nonce:    uint64(Nonce),
		To:       &userAddress,
		Value:    big.NewInt(0),
		Gas:      1000000,
		GasPrice: big.NewInt(params.InitialBaseFee),
		Data:     data,
	})
	return tx
}

func parseContractAddress(bc *core.BlockChain) *common.Address {
	latestHeader := bc.CurrentBlock()
	receipts := bc.GetReceiptsByHash(latestHeader.Hash())
	if receipts == nil {
		fmt.Println("no receipts in latest block")
	} else {
		for i := range receipts {
			// Search the receipt which contractAddress is not none
			if receipts[i].ContractAddress != (common.Address{}) {
				fmt.Printf("ERC20Address: %v \n", receipts[i].ContractAddress)
				return &receipts[i].ContractAddress
			}
		}
	}
	return nil
}

func TestDataStruct(t *testing.T) {
	prefix := []byte{0x0A, 0x0D, 0x03}
	data := append(prefix, bankAddress.Bytes()...)
	data = append(data, userAddress.Bytes()...)
	fmt.Printf("hex.EncodeToString(data): %v\n", hex.EncodeToString(data))
	// Data's legitimacy check
	if len(data) == 43 && data[0] == 0x0A && data[1] == 0x0D && data[2] == 0x03 {
	} else {
		fmt.Println("Tx don't use non-native tokens")
	}
}

// Test voucher contract in EVM
func TestEVMDeploy(t *testing.T) {
	bc := newBlockChain()
	txpool := newTxPool(bc)
	backend := &TestBackend{bc, txpool}
	// Create Miner
	minerConfig := miner.Config{Etherbase: common.HexToAddress("123456789")}
	miner := miner.New(backend, &minerConfig, bc.Config(), new(event.TypeMux), bc.Engine(), nil)
	miner.Start()
	defer miner.Stop()
	// Attemp to construct transaction
	tx0 := newDeployContractTx(bc, 0)
	errs := txpool.Add([]*types.Transaction{tx0}, true, false)
	t.Logf("tx0 Add to errs: %v\n", errs)
	// wait for consense
	time.Sleep(10 * time.Second)
	// Test contract depolyment
	ERC20Address := parseContractAddress(bc)
	if ERC20Address == nil {
		t.Errorf("Contract address parse fail")
		t.Fail()
	}
	// Test to call contract
	evm := newEVM(bc)
	var err error
	balance := new(big.Int)
	var flag bool
	voucherName := "BitCoin"

	balanceOfMethod := voucher.BalanceOf.Bind(ERC20Address)
	useMethod := voucher.Use.Bind(ERC20Address)
	buyMethod := voucher.Buy.Bind(ERC20Address)
	CreateVoucherMethod := voucher.CreateVoucher.Bind(ERC20Address)
	// Create new voucher
	_, err = CreateVoucherMethod.Execute(evm, nil, &bankAddress, uint256.NewInt(0), voucherName, big.NewInt(1))
	if err != nil {
		t.Fail()
	}
	// Bank buy voucher, value=1000 convert to 1000 BitCoin voucher
	_, err = buyMethod.Execute(evm, nil, &bankAddress, uint256.NewInt(1000), voucherName, big.NewInt(1000))
	if err != nil {
		t.Fail()
	}
	// Bank use voucher
	_, err = useMethod.Execute(evm, &flag, &bankAddress, uint256.NewInt(0), voucherName, big.NewInt(1000))
	if err != nil {
		t.Fail()
	}
	// Look up balance of user
	_, err = balanceOfMethod.Execute(evm, &balance, &bankAddress, uint256.NewInt(0), voucherName, bankAddress)
	fmt.Printf("Account balance: %v\n", balance)
	if err != nil {
		t.Fail()
	}
}

func newTx(bc *core.BlockChain, Nonce int, to *common.Address, value *big.Int, data []byte) *types.Transaction {
	// Construct tx.Data
	signer := types.LatestSigner(bc.Config())
	tx := types.MustSignNewTx(bankKey, signer, &types.AccessListTx{
		ChainID:  bc.Config().ChainID,
		Nonce:    uint64(Nonce),
		To:       to,
		Value:    value,
		GasPrice: big.NewInt(params.InitialBaseFee),
		Data:     data,
	})
	return tx
}

// Construct TX to test voucher
func TestTxDeploy(t *testing.T) {
	bc := newBlockChain()
	txpool := newTxPool(bc)
	backend := &TestBackend{bc, txpool}
	minerConfig := miner.Config{Etherbase: common.HexToAddress("123456789")}
	miner := miner.New(backend, &minerConfig, bc.Config(), new(event.TypeMux), bc.Engine(), nil)
	miner.Start()
	defer miner.Stop()
	tx0 := newDeployContractTx(bc, 0)
	errs := txpool.Add([]*types.Transaction{tx0}, true, false)
	t.Logf("tx0 Add to errs: %v\n", errs)
	time.Sleep(20 * time.Second)
	VoucherAddress := parseContractAddress(bc)
	if VoucherAddress == nil {
		t.Errorf("Contract address parse fail")
		t.Fail()
	}

	convertRate := big.NewInt(1)
	tokenName := "BitCoin"
	// Test to Create voucher
	input, err := contractAbi.Pack("createVoucher", tokenName, convertRate)
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	tx1 := newTx(bc, 1, VoucherAddress, big.NewInt(0), input)
	errs = txpool.Add([]*types.Transaction{tx1}, true, false)
	t.Logf("tx1 Add to errs: %v\n", errs)

	// Test to Buy voucher
	amount := big.NewInt(1000000000000000000)
	valueAmount := big.NewInt(1)
	valueAmount.Mul(amount, convertRate)

	input, err = contractAbi.Pack("buy", tokenName, amount)
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	tx2 := newTx(bc, 2, VoucherAddress, valueAmount, input)
	errs = txpool.Add([]*types.Transaction{tx2}, true, false)
	t.Logf("tx2 Add to errs: %v\n", errs)

	// Construct type of 03 data: identifier + contractAddress + TokenName
	input2 := []byte{0x0A, 0x0D, 0x03}
	input2 = append(input2, VoucherAddress[:]...)

	a := make([]byte, hex.EncodedLen(len(tokenName)))
	hex.Encode(a, []byte(tokenName))
	input2 = append(input2, a...)

	tx3 := newTx(bc, 3, &userAddress, big.NewInt(0), input2)
	errs = txpool.Add([]*types.Transaction{tx3}, true, false)
	t.Logf("tx3 Add to errs: %v\n", errs)
	time.Sleep(30 * time.Second)
}

func TestNormal(t *testing.T) {
	CreateVoucherString := "718b23b9000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000007426974436f696e00000000000000000000000000000000000000000000000000"
	data, _ := hex.DecodeString(CreateVoucherString)
	fmt.Printf("data: %v\n", data)

	input, err := contractAbi.Pack("buy", "BitCoin", big.NewInt(1))
	fmt.Printf("err: %v\n", err)
	fmt.Printf("input: %s\n", hex.EncodeToString(input))

	voucherName := "BitCoin"
	a := hex.EncodeToString([]byte(voucherName))
	fmt.Printf("a: %v\n", a)

	b, _ := hex.DecodeString(a)
	fmt.Printf("b: %s\n", b)
}

// // Test whether contracts can be deployed using Genesis Blocks
// func TestGenesisDeploy(t *testing.T) {
// 	// Create new blockchain
// 	bc := newBlockChain()
// 	evm := newEVM(bc)
// 	// Test EVM
// 	input := ERC20.GetEncodedAbi(ERC20.totalSupplySelector, [][]byte{})
// 	gas := uint64(100000)
// 	value := uint256.NewInt(0)
// 	result, _, err := evm.Call(vm.AccountRef(bankAddress), contractAddress, input, gas, value)
// 	if err != nil {
// 		t.Errorf("failed to execute EVM contract: %v\n", err)
// 	} else {
// 		t.Logf("Result: %x\n", result)
// 	}
// }
