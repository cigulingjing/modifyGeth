package vm

import "github.com/ethereum/go-ethereum/crypto/ring"

// bls12381Pairing implements EIP-2537 Pairing precompile.
type panguRingsigVer struct{}

// RequiredGas returns the gas required to execute the pre-compiled contract.
func (rsv *panguRingsigVer) RequiredGas(input []byte) uint64 {
	return 100
}

func (rsv *panguRingsigVer) Run(input []byte, blkCtx BlockContext) ([]byte, error) {
	ringsig, err := ring.DeserializeRingSig(input)
	if err != nil {
		return nil, err
	}

	if !ringsig.Verify() {
		return []byte{0}, nil
	} else {
		return []byte{1}, nil
	}
}
