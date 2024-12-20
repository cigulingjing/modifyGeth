package types

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// PowTx is a transaction with a hash_nonce field for Proof of Work.
type PowTx struct {
	ChainID     *big.Int
	Nonce       uint64
	GasTipCap   *big.Int // a.k.a. maxPriorityFeePerGas
	GasFeeCap   *big.Int // a.k.a. maxFeePerGas
	Gas         uint64
	To          *common.Address `rlp:"nil"` // nil means contract creation
	Value       *big.Int
	Data        []byte
	AccessList  AccessList
	HashNonce   uint64 // New field for controlling txHash
	StartHeight uint64 // New field for height verification

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *PowTx) copy() TxData {
	cpy := &PowTx{
		Nonce:       tx.Nonce,
		To:          copyAddressPtr(tx.To),
		Data:        common.CopyBytes(tx.Data),
		Gas:         tx.Gas,
		AccessList:  make(AccessList, len(tx.AccessList)),
		Value:       new(big.Int),
		ChainID:     new(big.Int),
		GasTipCap:   new(big.Int),
		GasFeeCap:   new(big.Int),
		HashNonce:   tx.HashNonce,
		StartHeight: tx.StartHeight,
		V:           new(big.Int),
		R:           new(big.Int),
		S:           new(big.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.
func (tx *PowTx) txType() byte           { return PowTxType }
func (tx *PowTx) chainID() *big.Int      { return tx.ChainID }
func (tx *PowTx) accessList() AccessList { return tx.AccessList }
func (tx *PowTx) data() []byte           { return tx.Data }
func (tx *PowTx) gas() uint64            { return tx.Gas }
func (tx *PowTx) gasFeeCap() *big.Int    { return tx.GasFeeCap }
func (tx *PowTx) gasTipCap() *big.Int    { return tx.GasTipCap }
func (tx *PowTx) gasPrice() *big.Int     { return tx.GasFeeCap }
func (tx *PowTx) value() *big.Int        { return tx.Value }
func (tx *PowTx) nonce() uint64          { return tx.Nonce }
func (tx *PowTx) to() *common.Address    { return tx.To }

func (tx *PowTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	tip := dst.Sub(tx.GasFeeCap, baseFee)
	if tip.Cmp(tx.GasTipCap) > 0 {
		tip.Set(tx.GasTipCap)
	}
	return tip.Add(tip, baseFee)
}

func (tx *PowTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *PowTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

func (tx *PowTx) encode(b *bytes.Buffer) error {
	return rlp.Encode(b, tx)
}

func (tx *PowTx) decode(input []byte) error {
	return rlp.DecodeBytes(input, tx)
}

// VerifyWithDifficulty checks if the transaction hash is smaller than the given difficulty.

// Hash returns the hash of the PowTx, which includes the hash_nonce.
func (tx *PowTx) Hash() common.Hash {
	return rlpHash([]interface{}{
		tx.ChainID,
		tx.Nonce,
		tx.GasTipCap,
		tx.GasFeeCap,
		tx.Gas,
		tx.To,
		tx.Value,
		tx.Data,
		tx.AccessList,
		tx.HashNonce,
		tx.StartHeight,
		tx.V,
		tx.R,
		tx.S,
	})
}

func VerifyTxWithDifficulty(tx *Transaction, difficulty *big.Int) (bool, *big.Int) {
	hash := tx.Hash()
	hashInt := new(big.Int).SetBytes(hash[:])

	// Calculate the target based on the difficulty
	// target = 2^256 / difficulty
	target := new(big.Int).Lsh(big.NewInt(1), 256)
	target.Div(target, difficulty)

	// Compare the hash with the target
	return hashInt.Cmp(target) < 0, hashInt
}

func VerifyTxHeight(tx *Transaction, currentHeight uint64, modHeight uint64) (bool, error) {
	if modHeight == 0 {
		return false, fmt.Errorf("modHeight cannot be zero")
	}

	// Get the inner PowTx data
	powTx, ok := tx.inner.(*PowTx)
	if !ok {
		return false, fmt.Errorf("not a PoW transaction")
	}

	hash := tx.Hash()
	hashInt := new(big.Int).SetBytes(hash[:])

	// Calculate modulo
	mod := new(big.Int).Mod(hashInt, big.NewInt(int64(modHeight))).Uint64()

	// Check if current height meets the requirement
	requiredHeight := powTx.StartHeight + mod + modHeight
	return currentHeight >= requiredHeight, nil
}

// EncodeRLP implements rlp.Encoder
func (tx *PowTx) EncodeRLP(w io.Writer) error {
	enc := &powTxData{
		AccountNonce: tx.Nonce,
		Price:        tx.GasTipCap,
		GasLimit:     tx.Gas,
		Recipient:    tx.To,
		Amount:       tx.Value,
		Payload:      tx.Data,
		V:            tx.V,
		R:            tx.R,
		S:            tx.S,
		ChainID:      tx.ChainID,
		HashNonce:    tx.HashNonce,
		StartHeight:  tx.StartHeight,
	}
	return rlp.Encode(w, enc)
}

// DecodeRLP implements rlp.Decoder
func (tx *PowTx) DecodeRLP(s *rlp.Stream) error {
	var dec powTxData
	if err := s.Decode(&dec); err != nil {
		return err
	}
	tx.Nonce = dec.AccountNonce
	tx.GasTipCap = dec.Price
	tx.Gas = dec.GasLimit
	tx.To = dec.Recipient
	tx.Value = dec.Amount
	tx.Data = dec.Payload
	tx.V = dec.V
	tx.R = dec.R
	tx.S = dec.S
	tx.ChainID = dec.ChainID
	tx.HashNonce = dec.HashNonce
	tx.StartHeight = dec.StartHeight
	return nil
}

// powTxData is the data structure that gets serialized/deserialized
type powTxData struct {
	AccountNonce uint64
	Price        *big.Int
	GasLimit     uint64
	Recipient    *common.Address
	Amount       *big.Int
	Payload      []byte
	V            *big.Int
	R            *big.Int
	S            *big.Int
	ChainID      *big.Int
	HashNonce    uint64
	StartHeight  uint64
}
