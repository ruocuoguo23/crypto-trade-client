package sign

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// Keys privKey and pubKey is not a pair normally
type Keys struct {
	// private key used to sign data
	privKey []byte
	// public used to check signature
	pubKeys map[string][]byte
}

func NewService(privKeyHex string, pubKeyHexes map[string]string) (*Keys, error) {
	privKey, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode private key during new Keys failed. err = %v", err)
	}

	pubKeys := make(map[string][]byte)
	for k, v := range pubKeyHexes {
		pubKey, err := hex.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("decode public key during new Keys failed. err = %v", err)
		}
		pubKeys[k] = pubKey
	}

	return &Keys{
		privKey: privKey,
		pubKeys: pubKeys,
	}, nil
}

func (s *Keys) Sign(body []byte) (string, error) {
	b := sha256.Sum256(body)
	sign, err := secp256k1.Sign(b[:], s.privKey)
	if err != nil {
		return "", fmt.Errorf("sign data failed. err = %v", err)
	}

	// The header byte: 0x1B = first key with even y, 0x1C = first key with odd y,
	//                  0x1D = second key with even y, 0x1E = second key with odd y
	// FIXME - it is not used
	sign[64] = sign[64] + 27
	return hex.EncodeToString(sign), nil
}

type SignatureAlgorithm string

const (
	Secp256k1 = SignatureAlgorithm("SECP256K1")
	Ed25519   = SignatureAlgorithm("ED25519")
)

func (s *Keys) Verify(body []byte, sig string, walletName string, appId string, curve SignatureAlgorithm) bool {
	k := walletName + "-" + appId
	if _, ok := s.pubKeys[k]; !ok {
		return false
	}
	sign, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	// check the Secp256k1 sign length before verify, for length not valid, it will panic
	// check the Ed25519 public key length before verify, for length not valid, it will panic
	if (curve == Secp256k1 && len(sign) < 64) || (curve == Ed25519 && len(s.pubKeys[k]) != 32) {
		return false
	}

	switch curve {
	case Ed25519:
		return ed25519.Verify(s.pubKeys[k], body, sign)
	case Secp256k1:
		b := sha256.Sum256(body)
		return secp256k1.VerifySignature(s.pubKeys[k], b[:], sign[:64])
	default:
		// if not specified or unknown, use secp256k1
		b := sha256.Sum256(body)
		return secp256k1.VerifySignature(s.pubKeys[k], b[:], sign[:64])
	}
}

func Construct(uri, method, contentType, walletName, appId string, body []byte) []byte {
	info := ""
	if method == "GET" {
		info = "url" + uri + "method" + method + "wallet" + walletName + "appid" + appId
	} else {
		info = "url" + uri + "method" + method + "content-type" + contentType + "wallet" + walletName + "appid" + appId
	}
	info = strings.ToLower(info)

	bs := make([]byte, len(info), len(info))
	copy(bs, []byte(info)[:])
	bs = append(bs, body...)
	return bs
}
