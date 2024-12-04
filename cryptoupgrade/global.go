package cryptoupgrade

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// * I hope the address of the contract is fixed, which involves the design of the underlying code
var (
	CodeStorageAddress   = common.BytesToAddress([]byte{67})
	PullCodeEventHash    = crypto.Keccak256Hash([]byte("pullcode(string)"))
)
