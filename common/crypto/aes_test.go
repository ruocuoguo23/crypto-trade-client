package crypto

import (
	"encoding/hex"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAES(t *testing.T) {
	Convey("Test AES Encryption and Decryption", t, func() {
		key := []byte("a very very very very secret key") // 32 bytes key
		plaintext := []byte("Hello, World!")

		Convey("Encrypt and Decrypt", func() {
			ciphertext, err := Encrypt(key, plaintext)
			So(err, ShouldBeNil)
			So(ciphertext, ShouldNotBeNil)

			decryptedText, err := Decrypt(key, ciphertext)
			So(err, ShouldBeNil)
			So(decryptedText, ShouldResemble, plaintext)
		})

		Convey("Encrypt and Decrypt with Hex Encoding", func() {
			ciphertext, err := Encrypt(key, plaintext)
			So(err, ShouldBeNil)
			So(ciphertext, ShouldNotBeNil)

			// Convert to hex string
			hexCiphertext := hex.EncodeToString(ciphertext)
			decodedCiphertext, err := hex.DecodeString(hexCiphertext)
			// print hexCiphertext
			t.Logf("hexCiphertext: %s", hexCiphertext)
			So(err, ShouldBeNil)

			decryptedText, err := Decrypt(key, decodedCiphertext)
			So(err, ShouldBeNil)
			So(decryptedText, ShouldResemble, plaintext)
		})
	})
}
