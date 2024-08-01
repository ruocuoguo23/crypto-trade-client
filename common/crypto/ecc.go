package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Signer 定义签名接口
type Signer interface {
	Sign(message []byte) ([]byte, error)
}

// Verifier 定义验证接口
type Verifier interface {
	Verify(message, signature []byte) bool
}

// secp256k1Signer 实现 secp256k1 签名
type secp256k1Signer struct {
	privateKey *ecdsa.PrivateKey
}

// secp256k1Verifier 实现 secp256k1 验证
type secp256k1Verifier struct {
	publicKey *ecdsa.PublicKey
}

// ed25519Signer 实现 ed25519 签名
type ed25519Signer struct {
	privateKey ed25519.PrivateKey
}

// ed25519Verifier 实现 ed25519 验证
type ed25519Verifier struct {
	publicKey ed25519.PublicKey
}

func PublicKeyToBytes(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)
}

// Sign 使用 secp256k1 签名
func (s *secp256k1Signer) Sign(message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	signature, err := secp256k1.Sign(hash[:], crypto.FromECDSA(s.privateKey))
	if err != nil {
		return nil, err
	}
	signature[64] = signature[64] + 27
	return signature, nil
}

// Verify 使用 secp256k1 验证签名
func (v *secp256k1Verifier) Verify(message, signature []byte) bool {
	hash := sha256.Sum256(message)
	return secp256k1.VerifySignature(crypto.CompressPubkey(v.publicKey), hash[:], signature[:64])
}

// Sign 使用 ed25519 签名
func (s *ed25519Signer) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(s.privateKey, message), nil
}

// Verify 使用 ed25519 验证签名
func (v *ed25519Verifier) Verify(message, signature []byte) bool {
	return ed25519.Verify(v.publicKey, message, signature)
}
