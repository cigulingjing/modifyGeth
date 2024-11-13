package ring

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/blake2b"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Refer to https://github.com/estensen/lrs/blob/master/ring/ring.go and https://github.com/mottla/Solidity-RingSignature/blob/master/CryptoNote1/ring_sign.go implementation
var defaultCurve = secp256k1.S256()

type RingKey []*ecdsa.PublicKey

func (rk RingKey) Bytes() []byte {
	var buf []byte
	for _, key := range rk {
		buf = append(buf, key.X.Bytes()...)
		buf = append(buf, key.Y.Bytes()...)
	}
	return buf
}

func RingKeyFromBytes(buf []byte, size int) RingKey {
	rk := make(RingKey, size)
	for i := 0; i < size; i++ {
		x := new(big.Int).SetBytes(buf[i*64 : i*64+32])
		y := new(big.Int).SetBytes(buf[i*64+32 : i*64+64])
		rk[i] = &ecdsa.PublicKey{
			Curve: defaultCurve,
			X:     x,
			Y:     y,
		}
	}
	return rk
}

type RingSig struct {
	Size       int              //size of the ring
	Message    [32]byte         //message to be signed
	C          *big.Int         //signature value
	S          []*big.Int       //random values(fake secret keys)
	PubkeyList RingKey          //ring of public keys
	Image      *ecdsa.PublicKey //image of the signer
	Curve      elliptic.Curve   // curve used for the ring signature
}

func NewRingSig(message [32]byte, pubKeys RingKey) (*RingSig, error) {
	ringSize := len(pubKeys)

	if ringSize < 2 {
		return nil, errors.New("size of ring less than two")
	}

	return &RingSig{
		Size:       ringSize,
		Message:    message,
		C:          nil,
		S:          make([]*big.Int, ringSize),
		PubkeyList: pubKeys,
		Image:      nil,
		Curve:      defaultCurve,
	}, nil
}

// Sign signs the message with the private key of the signer
func (rs *RingSig) Sign(signerKey *ecdsa.PrivateKey) error {
	ringSize := rs.Size

	signerPosition := -1
	// cal signer position
	for i, key := range rs.PubkeyList {
		if key == signerKey.Public().(*ecdsa.PublicKey) {
			signerPosition = i
			break
		}
	}

	if signerPosition >= ringSize || signerPosition < 0 {
		return errors.New("secret index out of range of ring size")
	}

	// setup
	pubKey := &signerKey.PublicKey // public key of the signer
	curve := pubKey.Curve

	// check that key at index s is indeed the signer
	if rs.PubkeyList[signerPosition] != pubKey {
		return errors.New("signer key does not match key at index")
	}

	// generate key image
	image := GenKeyImage(signerKey)
	rs.Image = image

	// record for the hash value c_i
	C := make([]*big.Int, ringSize)

	// start with the signer's key
	// pick random scalar rValue
	rValue, err := rand.Int(rand.Reader, curve.Params().P)
	if err != nil {
		return err
	}

	// L_j = rValue * G
	L_jx, L_jy := curve.ScalarBaseMult(rValue.Bytes())

	// R_j = rValue * H_p(P_j)
	H_jx, H_jy := HashPoint(pubKey)
	R_jx, R_jy := curve.ScalarMult(H_jx, H_jy, rValue.Bytes())

	// cal c_j = H(m, L_j, R_j)
	L_j, R_j := append(L_jx.Bytes(), L_jy.Bytes()...), append(R_jx.Bytes(), R_jy.Bytes()...)
	c_j := blake2b.Sum256(append(rs.Message[:], append(L_j, R_j...)...))
	index_j := (signerPosition + 1) % ringSize
	C[index_j] = new(big.Int).SetBytes(c_j[:])

	// start loop to compute c_i (i != j)
	for i := 1; i < ringSize; i++ {
		index := (signerPosition + i) % ringSize

		// pick random scalar sValue
		sValue_i, err := rand.Int(rand.Reader, curve.Params().N)
		if err != nil {
			return err
		}
		rs.S[index] = sValue_i

		if rs.PubkeyList[index] == nil {
			return fmt.Errorf("public key at index %d is nil", index)
		}
		// L_j+1 = sValue_j+1 * G + c_j+1 * P_j+1
		sx, sy := curve.ScalarBaseMult(sValue_i.Bytes())                                             // s = sValue_j+1 * G
		px, py := curve.ScalarMult(rs.PubkeyList[index].X, rs.PubkeyList[index].Y, C[index].Bytes()) // p = c_j+1 * P_j+1
		L_x, L_y := curve.Add(sx, sy, px, py)

		// R_j+1 = sValue_j+1 * H_p(P_j+1) + c_j+1 * I
		H_x, H_y := HashPoint(rs.PubkeyList[index])
		sx, sy = curve.ScalarMult(H_x, H_y, sValue_i.Bytes())         // s = sValue_j+1 * H_p(P_j+1)
		px, py = curve.ScalarMult(image.X, image.Y, C[index].Bytes()) // p = c_j+1 * I
		R_x, R_y := curve.Add(sx, sy, px, py)

		// cal c_j+1 = H(m, L_j+1, R_j+1)
		L_i, R_i := append(L_x.Bytes(), L_y.Bytes()...), append(R_x.Bytes(), R_y.Bytes()...)
		c_i := blake2b.Sum256(append(rs.Message[:], append(L_i, R_i...)...))

		if i == ringSize-1 {
			C[signerPosition] = new(big.Int).SetBytes(c_i[:])
		} else {
			C[(index+1)%ringSize] = new(big.Int).SetBytes(c_i[:])
		}
	}

	// recalculate to close the ring
	rs.S[signerPosition] = new(big.Int).Mod(new(big.Int).Sub(rValue, new(big.Int).Mul(C[signerPosition], signerKey.D)), curve.Params().N)

	// check that the ring is closed
	// rValue * G = sValue * G + cValue * P
	rx, ry := curve.ScalarBaseMult(rValue.Bytes())               // rValue * G
	sx, sy := curve.ScalarBaseMult(rs.S[signerPosition].Bytes()) // sValue * G
	px, py := curve.ScalarMult(rs.PubkeyList[signerPosition].X, rs.PubkeyList[signerPosition].Y, C[signerPosition].Bytes())
	l_x, l_y := curve.Add(sx, sy, px, py)

	// rValue * H_p(P) = sValue * H_p(P) + cValue * I
	px, py = curve.ScalarMult(rs.Image.X, rs.Image.Y, C[signerPosition].Bytes()) // px, py = c[signerPosition] * I
	hx, hy := HashPoint(rs.PubkeyList[signerPosition])
	tx, ty := curve.ScalarMult(hx, hy, rValue.Bytes())
	sx, sy = curve.ScalarMult(hx, hy, rs.S[signerPosition].Bytes()) // sx, sy = sValue * H_p(P)
	r_x, r_y := curve.Add(sx, sy, px, py)

	l := append(l_x.Bytes(), l_y.Bytes()...)
	r := append(r_x.Bytes(), r_y.Bytes()...)
	c := blake2b.Sum256(append(rs.Message[:], append(l, r...)...)) //check that c[signerPosition + 1] = H(m, L[signerPosition], R[signerPosition])

	if !bytes.Equal(rx.Bytes(), l_x.Bytes()) || !bytes.Equal(ry.Bytes(), l_y.Bytes()) || !bytes.Equal(tx.Bytes(), r_x.Bytes()) || !bytes.Equal(ty.Bytes(), r_y.Bytes()) || !bytes.Equal(C[(signerPosition+1)%ringSize].Bytes(), c[:]) {
		return errors.New("ring is not closed")
	}

	// everything ok, add c[0] to signature
	rs.C = C[0]

	return nil
}

// Verify verifies the ring signature, returns true if the signature is valid
func (rs *RingSig) Verify() bool {
	// setup
	curve := rs.Curve
	ringSize := rs.Size
	message := rs.Message
	pubKeyList := rs.PubkeyList
	image := rs.Image

	C := make([]*big.Int, ringSize)
	C[0] = rs.C

	S := rs.S

	// c[i+1] = H(m, s[i]*G + c[i]*P[i]) and c[0] = H(m, s[n-1]*G + c[n-1]*P[n-1]) where n is the ring size
	for i := 0; i < ringSize; i++ {
		fmt.Println(pubKeyList[i].X, pubKeyList[i].Y)
		fmt.Println(C[i].Bytes())
		// calculate L_i = si*G + Ci*P_i
		px, py := curve.ScalarMult(pubKeyList[i].X, pubKeyList[i].Y, C[i].Bytes()) // px, py = Ci*P_i
		sx, sy := curve.ScalarBaseMult(S[i].Bytes())                               // sx, sy = s[i]*G
		lx, ly := curve.Add(sx, sy, px, py)

		// R_i = si*H_p(P_i) + Ci*I
		px, py = curve.ScalarMult(image.X, image.Y, C[i].Bytes()) // px, py = c[i]*I
		hx, hy := HashPoint(pubKeyList[i])
		sx, sy = curve.ScalarMult(hx, hy, S[i].Bytes()) // sx, sy = s[i]*H_p(P[i])
		rx, ry := curve.Add(sx, sy, px, py)

		// c[i+1] = H(m, L_i, R_i)
		l := append(lx.Bytes(), ly.Bytes()...)
		r := append(rx.Bytes(), ry.Bytes()...)
		Ci := blake2b.Sum256(append(message[:], append(l, r...)...))

		if i == ringSize-1 {
			C[0] = new(big.Int).SetBytes(Ci[:])
		} else {
			C[i+1] = new(big.Int).SetBytes(Ci[:])
		}
	}

	return bytes.Equal(rs.C.Bytes(), C[0].Bytes())
}

func LinkCheck(sig1 *RingSig, sig2 *RingSig) bool {
	return sig1.Image.X.Cmp(sig2.Image.X) == 0 && sig1.Image.Y.Cmp(sig2.Image.Y) == 0
}

// GenKeyImage calculates key image
// It is not possible to make a correlation between the public key and the corresponding key image
// I = x * H_p(P) where H_p is a hash function that returns a point
// H_p(P) = blake2b(P) * G
func GenKeyImage(privKey *ecdsa.PrivateKey) *ecdsa.PublicKey {
	pubKey := privKey.Public().(*ecdsa.PublicKey)
	image := new(ecdsa.PublicKey)

	// blake2b(P)
	hx, hy := HashPoint(pubKey)

	// H_p(P) = x * blake2b(P) * G
	ix, iy := privKey.Curve.ScalarMult(hx, hy, privKey.D.Bytes())

	image.X = ix
	image.Y = iy
	return image
}

// HashPoint returns point on the curve
func HashPoint(p *ecdsa.PublicKey) (hx, hy *big.Int) {
	hash := blake2b.Sum256(append(p.X.Bytes(), p.Y.Bytes()...))
	return p.Curve.ScalarBaseMult(hash[:]) // g^H'()
}

// Hash is a function that returns a value in Z_p (ed25519 base field)
func Hash(m [32]byte, l, r []byte) [32]byte {
	return blake2b.Sum256(append(m[:], append(l, r...)...))
}

func PadTo32Bytes(in []byte) (out []byte) {
	out = append(out, in...)
	for {
		if len(out) == 32 {
			return
		}
		out = append([]byte{0}, out...)
	}
}

// SerializeRingSig serializes ring signature
func SerializeRingSig(rs *RingSig) ([]byte, error) {
	var buf []byte
	// add size and message
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(rs.Size))
	buf = append(buf, b[:]...)                        // 8 bytes
	buf = append(buf, PadTo32Bytes(rs.Message[:])...) // 32 bytes
	buf = append(buf, PadTo32Bytes(rs.C.Bytes())...)  // 32 bytes
	// 96 bytes each iteration
	for i := 0; i < rs.Size; i++ {
		buf = append(buf, PadTo32Bytes(rs.S[i].Bytes())...)
		buf = append(buf, PadTo32Bytes(rs.PubkeyList[i].X.Bytes())...)
		buf = append(buf, PadTo32Bytes(rs.PubkeyList[i].Y.Bytes())...)
	}
	// 64 bytes for image
	buf = append(buf, PadTo32Bytes(rs.Image.X.Bytes())...)
	buf = append(buf, PadTo32Bytes(rs.Image.Y.Bytes())...)

	if len(buf) != 32*(3*rs.Size+4)+8 {
		return []byte{}, errors.New("could not serialize ring signature")
	}

	return buf, nil
}

// DeserializeRingSig deserializes ring signature
func DeserializeRingSig(buf []byte) (*RingSig, error) {
	if len(buf) < 8 {
		return nil, errors.New("buffer too short when recovering size")
	}

	rs := new(RingSig)
	rs.Size = int(binary.BigEndian.Uint64(buf[0:8]))
	fmt.Println(rs.Size)

	if len(buf) < 72 {
		return nil, errors.New("buffer too short when recovering message and C")
	}

	// recover message and c
	var mBytes [32]byte
	copy(mBytes[:], buf[8:40])
	rs.Message = mBytes
	rs.C = new(big.Int).SetBytes(buf[40:72])

	if len(buf) < 96*rs.Size+136 {
		return nil, errors.New("buffer too short when recovering S and PubkeyList")
	}

	rs.S = make([]*big.Int, rs.Size)
	rs.PubkeyList = make([]*ecdsa.PublicKey, rs.Size)

	j := 0

	for index := 72; index < 96*rs.Size+72; index += 96 {
		si := new(big.Int).SetBytes(buf[index : index+32])
		xi := new(big.Int).SetBytes(buf[index+32 : index+64])
		yi := new(big.Int).SetBytes(buf[index+64 : index+96])

		rs.S[j] = si
		rs.PubkeyList[j] = &ecdsa.PublicKey{
			Curve: defaultCurve,
			X:     xi,
			Y:     yi,
		}

		j++
	}

	rs.Image = &ecdsa.PublicKey{
		Curve: defaultCurve,
		X:     new(big.Int).SetBytes(buf[96*rs.Size+72 : 96*rs.Size+104]),
		Y:     new(big.Int).SetBytes(buf[96*rs.Size+104 : 96*rs.Size+136]),
	}

	rs.Curve = defaultCurve

	return rs, nil
}
