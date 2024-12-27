package tests

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/voucher"
	"github.com/holiman/uint256"
	// "google.golang.org/grpc"
)

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
func TestVoucherWithEVM(t *testing.T) {
	backend := newTestBackend()
	// Create Miner
	miner := backend.CreateMiner()
	miner.Start()
	defer miner.Stop()
	// Create EVM instance to call contract
	evm := newEVM(backend.bc)
	var (
		err            error
		flag           bool
		voucherName    = "BitCoin"
		balance        = new(big.Int)
		conversionRate = big.NewInt(2)
	)

	// Create new voucher
	_, err = voucher.CreateVoucher.Execute(evm, nil, &bankAddress, uint256.NewInt(0), voucherName, conversionRate)
	if err != nil {
		t.Fail()
	}
	// Bank buy voucher, value=1000 convert to 2000 BitCoin voucher
	_, err = voucher.Buy.Execute(evm, nil, &bankAddress, uint256.NewInt(1000), voucherName)
	if err != nil {
		t.Fail()
	}
	// Bank use voucher
	_, err = voucher.Use.Execute(evm, &flag, &bankAddress, uint256.NewInt(0), voucherName, big.NewInt(1000))
	if err != nil {
		t.Fail()
	}
	// Look up balance of user
	_, err = voucher.BalanceOf.Execute(evm, &balance, &bankAddress, uint256.NewInt(0), voucherName, bankAddress)
	fmt.Printf("Account balance: %v\n", balance)
	if err != nil {
		t.Fail()
	}
}

func NewTx(bc *core.BlockChain, Nonce int, to *common.Address, value *big.Int, data []byte) *types.Transaction {
	// Construct tx.Data
	signer := types.LatestSigner(bc.Config())
	tx := types.MustSignNewTx(bankKey, signer, &types.AccessListTx{
		ChainID:  bc.Config().ChainID,
		Nonce:    uint64(Nonce),
		To:       to,
		Value:    value,
		Gas:      63696,
		GasPrice: big.NewInt(params.InitialBaseFee),
		Data:     data,
	})
	return tx
}

// Construct TX to test voucher
func TestVoucherWithTx(t *testing.T) {
	backend := newTestBackend()
	miner := backend.CreateMiner()
	miner.Start()
	defer miner.Stop()

	// tx0 := newDeployContractTx(backend.bc, 0)
	// backend.AddTx(tx0)
	// VoucherAddress := backend.parseContractAddress()
	VoucherAddress := &contractAddress

	// Test to Create voucher
	convertRate := big.NewInt(1)
	tokenName := "BitCoin"
	input, err := contractAbi.Pack("createVoucher", tokenName, convertRate)
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	tx1 := NewTx(backend.bc, 0, VoucherAddress, big.NewInt(0), input)
	// Test to Buy voucher
	valueAmount := big.NewInt(1000000)
	input, err = contractAbi.Pack("buy", tokenName)
	if err != nil {
		t.Errorf("err: %v\n", err)
	}
	tx2 := NewTx(backend.bc, 1, VoucherAddress, valueAmount, input)

	// Construct type of 03 data: identifier + contractAddress + TokenName
	input2 := []byte{0x0A, 0x0D, 0x03}

	var result [20]byte
	copy(result[:], []byte(tokenName))
	input2 = append(input2, result[:]...)
	tx3 := NewTx(backend.bc, 2, &userAddress, big.NewInt(0), input2)

	backend.AddTx(tx1)
	backend.AddTx(tx2)
	backend.AddTx(tx3)
}

// Test whether consensus layer can work
func TestPot(t *testing.T) {
	backend := newTestBackend()
	miner := backend.CreateMiner()
	miner.Start()
	defer miner.Stop()

	signer := types.LatestSigner(backend.bc.Config())
	tx := types.MustSignNewTx(bankKey, signer, &types.AccessListTx{
		ChainID:  backend.bc.Config().ChainID,
		Nonce:    uint64(0),
		To:       &userAddress,
		Value:    big.NewInt(1),
		Gas:      63696,
		GasPrice: big.NewInt(params.InitialBaseFee),
	})

	backend.AddTx(tx)
}

func TestNormal(t *testing.T) {
	CreateVoucherString := "718b23b9000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000007426974436f696e00000000000000000000000000000000000000000000000000"
	data, _ := hex.DecodeString(CreateVoucherString)
	fmt.Printf("data: %v\n", data)

	input, err := contractAbi.Pack("buy", "BitCoin", big.NewInt(1))
	fmt.Printf("err: %v\n", err)
	fmt.Printf("input: %s\n", hex.EncodeToString(input))

	voucherName := "BitCoin"
	aByte := []byte(voucherName)

	a := hex.EncodeToString(aByte)
	// a overflow check
	if len(a) > 20 {
		a = a[:20]
		t.Error("Token name is too long:", len(a))
	} else {
		t.Log("a:", a)
	}

	b, _ := hex.DecodeString(a)
	fmt.Printf("b: %s\n", b)
}
