package ring

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/blake2b"
)

func TestRingSigCreation(t *testing.T) {
	message := blake2b.Sum256([]byte("Test message"))
	curve := defaultCurve

	// Generate public key list
	keyCount := 5
	pubKeys := make([]*ecdsa.PublicKey, keyCount)
	for i := 0; i < keyCount; i++ {
		privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		pubKeys[i] = &privKey.PublicKey
	}

	_, err := NewRingSig(message, pubKeys)
	if err != nil {
		t.Errorf("Failed to create RingSig: %v", err)
	}
}

func TestRingSigSignAndVerify(t *testing.T) {
	message := blake2b.Sum256([]byte("Test message"))
	curve := defaultCurve

	// Generate key pairs
	keyCount := 5
	privKeys := make([]*ecdsa.PrivateKey, keyCount)
	pubKeys := make([]*ecdsa.PublicKey, keyCount)
	for i := 0; i < keyCount; i++ {
		privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		privKeys[i] = privKey
		pubKeys[i] = &privKey.PublicKey
	}

	// Create RingSig
	rs, err := NewRingSig(message, pubKeys)
	if err != nil {
		t.Errorf("Failed to create RingSig: %v", err)
	}

	// Sign
	signerIndex := 2
	err = rs.Sign(privKeys[signerIndex])
	if err != nil {
		t.Errorf("Failed to sign: %v", err)
	}

	// Verify signature
	if !rs.Verify() {
		t.Errorf("Signature verification failed")
	} else {
		t.Logf("Signature verification passed")
	}
}

// TestRingSigLinkability tests the linkability of ring signatures.
// It generates a set of key pairs, creates two ring signatures with different messages
// but signed by the same private key, and verifies that the signatures are linkable.
func TestRingSigLinkability(t *testing.T) {
	message1 := blake2b.Sum256([]byte("First message"))
	message2 := blake2b.Sum256([]byte("Second message"))
	curve := defaultCurve

	// Generate key pairs
	keyCount := 3
	privKeys := make([]*ecdsa.PrivateKey, keyCount)
	pubKeys := make([]*ecdsa.PublicKey, keyCount)
	for i := 0; i < keyCount; i++ {
		privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		privKeys[i] = privKey
		pubKeys[i] = &privKey.PublicKey
	}

	// Create first RingSig and sign
	rs1, err := NewRingSig(message1, pubKeys)
	if err != nil {
		t.Errorf("Failed to create RingSig: %v", err)
	}
	signerIndex := 1
	err = rs1.Sign(privKeys[signerIndex])
	if err != nil {
		t.Errorf("Failed to sign: %v", err)
	}

	// Create second RingSig and sign
	rs2, err := NewRingSig(message2, pubKeys)
	if err != nil {
		t.Errorf("Failed to create RingSig: %v", err)
	}
	err = rs2.Sign(privKeys[signerIndex])
	if err != nil {
		t.Errorf("Failed to sign: %v", err)
	}

	// Check linkability
	if !LinkCheck(rs1, rs2) {
		t.Errorf("Signatures should be linkable but were not detected as such")
	}
}

func TestRingSigTampering(t *testing.T) {
	message := blake2b.Sum256([]byte("Test message"))
	curve := defaultCurve

	// Generate key pairs
	keyCount := 5
	privKeys := make([]*ecdsa.PrivateKey, keyCount)
	pubKeys := make([]*ecdsa.PublicKey, keyCount)
	for i := 0; i < keyCount; i++ {
		privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		privKeys[i] = privKey
		pubKeys[i] = &privKey.PublicKey
	}

	// Create RingSig and sign
	rs, err := NewRingSig(message, pubKeys)
	if err != nil {
		t.Errorf("Failed to create RingSig: %v", err)
	}
	signerIndex := 2
	err = rs.Sign(privKeys[signerIndex])
	if err != nil {
		t.Errorf("Failed to sign: %v", err)
	}

	// Tamper with the signature
	rs.C = big.NewInt(1001)

	// Verify signature should fail
	if rs.Verify() {
		t.Errorf("Signature was tampered with but verification still passed")
	} else {
		t.Logf("Signature was tampered with, verification failed")
	}
}

func TestRingSigSerialization(t *testing.T) {
	message := blake2b.Sum256([]byte("Test message"))
	curve := defaultCurve

	// Generate key pairs
	keyCount := 5
	privKeys := make([]*ecdsa.PrivateKey, keyCount)
	pubKeys := make([]*ecdsa.PublicKey, keyCount)
	for i := 0; i < keyCount; i++ {
		privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}
		privKeys[i] = privKey
		pubKeys[i] = &privKey.PublicKey
	}

	// Create RingSig
	rs, err := NewRingSig(message, pubKeys)
	if err != nil {
		t.Errorf("Failed to create RingSig: %v", err)
	}

	// Sign
	signerIndex := 2
	err = rs.Sign(privKeys[signerIndex])
	if err != nil {
		t.Errorf("Failed to sign: %v", err)
	}

	// Serialize and deserialize
	serialized, err := SerializeRingSig(rs)
	if err != nil {
		t.Errorf("Failed to serialize RingSig: %v", err)
	}
	deserialized, err := DeserializeRingSig(serialized)
	if err != nil {
		t.Errorf("Failed to deserialize RingSig: %v", err)
	}

	// Verify deserialized signature
	if !deserialized.Verify() {
		t.Errorf("Deserialized signature verification failed")
	} else {
		t.Logf("Deserialized signature verification passed")
	}

}
