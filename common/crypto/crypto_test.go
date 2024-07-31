package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/sha3"
	"testing"
)

func TestSha256(t *testing.T) {
	Convey("Test Sha256", t, func() {
		data := []byte("hello")
		hash := sha256.Sum256(data)
		// to hex string
		hashStr := hex.EncodeToString(hash[:])
		So(hashStr, ShouldEqual, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824")
	})
}

func TestSha3(t *testing.T) {
	Convey("Test Sha3", t, func() {
		data := []byte("hello")
		hash := sha3.Sum256(data)
		// to hex string
		hashStr := hex.EncodeToString(hash[:])
		So(hashStr, ShouldEqual, "3338be694f50c5f338814986cdf0686453a888b84f424d792af4b9202398f392")
	})
}
