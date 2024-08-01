package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSecp256k1SignerAndVerifier(t *testing.T) {
	Convey("Test secp256k1 Verifier", t, func() {
		privateKey, err := crypto.GenerateKey()
		So(err, ShouldBeNil)
		signer := &secp256k1Signer{privateKey: privateKey}

		verifier := &secp256k1Verifier{publicKey: &privateKey.PublicKey}

		message := []byte("Hello, secp256k1!")
		signature, err := signer.Sign(message)
		So(err, ShouldBeNil)
		So(signature, ShouldNotBeNil)

		valid := verifier.Verify(message, signature)
		So(valid, ShouldBeTrue)
	})
}

func TestEd25519SignerAndVerifier(t *testing.T) {
	Convey("Test ed25519 Verifier", t, func() {
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		So(err, ShouldBeNil)
		signer := &ed25519Signer{privateKey: privateKey}
		verifier := &ed25519Verifier{publicKey: publicKey}

		message := []byte("Hello, ed25519!")
		signature, err := signer.Sign(message)
		So(err, ShouldBeNil)
		So(signature, ShouldNotBeNil)

		valid := verifier.Verify(message, signature)
		So(valid, ShouldBeTrue)
	})
}
