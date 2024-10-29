// Code generated by rlpgen. DO NOT EDIT.

package types

import "github.com/ethereum/go-ethereum/rlp"
import "io"

func (obj *Header) EncodeRLP(_w io.Writer) error {
	w := rlp.NewEncoderBuffer(_w)
	_tmp0 := w.List()
	w.WriteBytes(obj.ParentHash[:])
	w.WriteBytes(obj.UncleHash[:])
	w.WriteBytes(obj.Coinbase[:])
	w.WriteBytes(obj.Root[:])
	w.WriteBytes(obj.TxHash[:])
	w.WriteBytes(obj.ReceiptHash[:])
	w.WriteBytes(obj.Bloom[:])
	if obj.Difficulty == nil {
		w.Write(rlp.EmptyString)
	} else {
		if obj.Difficulty.Sign() == -1 {
			return rlp.ErrNegativeBigInt
		}
		w.WriteBigInt(obj.Difficulty)
	}
	if obj.Number == nil {
		w.Write(rlp.EmptyString)
	} else {
		if obj.Number.Sign() == -1 {
			return rlp.ErrNegativeBigInt
		}
		w.WriteBigInt(obj.Number)
	}
	w.WriteUint64(obj.GasLimit)
	w.WriteUint64(obj.GasUsed)
	w.WriteUint64(obj.Time)
	w.WriteBytes(obj.Extra)
	if obj.RandomNumber == nil {
		w.Write(rlp.EmptyString)
	} else {
		if obj.RandomNumber.Sign() == -1 {
			return rlp.ErrNegativeBigInt
		}
		w.WriteBigInt(obj.RandomNumber)
	}
	w.WriteBytes(obj.RandomRoot[:])
	w.WriteBytes(obj.MixDigest[:])
	w.WriteBytes(obj.Nonce[:])
	_tmp1 := obj.PoWGas != 0
	_tmp2 := obj.PowPrice != nil
	_tmp3 := len(obj.Tainted) > 0
	_tmp4 := obj.Incentive != nil
	_tmp5 := obj.BaseFee != nil
	_tmp6 := obj.WithdrawalsHash != nil
	_tmp7 := obj.BlobGasUsed != nil
	_tmp8 := obj.ExcessBlobGas != nil
	_tmp9 := obj.ParentBeaconRoot != nil
	if _tmp1 || _tmp2 || _tmp3 || _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 {
		w.WriteUint64(obj.PoWGas)
	}
	if _tmp2 || _tmp3 || _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 {
		if obj.PowPrice == nil {
			w.Write(rlp.EmptyString)
		} else {
			if obj.PowPrice.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.PowPrice)
		}
	}
	if _tmp3 || _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 {
		w.WriteBytes(obj.Tainted)
	}
	if _tmp4 || _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 {
		if obj.Incentive == nil {
			w.Write(rlp.EmptyString)
		} else {
			w.WriteUint256(obj.Incentive)
		}
	}
	if _tmp5 || _tmp6 || _tmp7 || _tmp8 || _tmp9 {
		if obj.BaseFee == nil {
			w.Write(rlp.EmptyString)
		} else {
			if obj.BaseFee.Sign() == -1 {
				return rlp.ErrNegativeBigInt
			}
			w.WriteBigInt(obj.BaseFee)
		}
	}
	if _tmp6 || _tmp7 || _tmp8 || _tmp9 {
		if obj.WithdrawalsHash == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteBytes(obj.WithdrawalsHash[:])
		}
	}
	if _tmp7 || _tmp8 || _tmp9 {
		if obj.BlobGasUsed == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteUint64((*obj.BlobGasUsed))
		}
	}
	if _tmp8 || _tmp9 {
		if obj.ExcessBlobGas == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteUint64((*obj.ExcessBlobGas))
		}
	}
	if _tmp9 {
		if obj.ParentBeaconRoot == nil {
			w.Write([]byte{0x80})
		} else {
			w.WriteBytes(obj.ParentBeaconRoot[:])
		}
	}
	w.ListEnd(_tmp0)
	return w.Flush()
}
