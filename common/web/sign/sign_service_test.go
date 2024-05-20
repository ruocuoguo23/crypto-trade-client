package sign

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"testing"
)

func getSignService(walletName, appid string) (*Keys, error) {
	privKeyHex := "6b61d1d17a299d61deabe15a64db62fead7389bb3f8d20719353982beab02521"
	privKey, _ := hex.DecodeString(privKeyHex)
	d := new(big.Int).SetBytes(privKey)

	curve := secp256k1.S256()
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = curve
	priv.D = d
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(privKey)
	pubKey := secp256k1.CompressPubkey(priv.PublicKey.X, priv.PublicKey.Y)
	pubKeyHexStr := hex.EncodeToString(pubKey)
	fmt.Printf("private key hex string: %v\n", hex.EncodeToString(privKey))
	fmt.Printf("public key (Compressed): %v\n", pubKeyHexStr)

	return NewService(privKeyHex, map[string]string{walletName + "-" + appid: pubKeyHexStr})
}

func TestVerify(t *testing.T) {
	Convey("test verify", t, func() {
		Convey("Verify using secp256k1 should be ok", func() {
			walletName := "test1"
			appid := "java"
			signService, err := getSignService(walletName, appid)
			So(err, ShouldBeNil)

			oriBody := "abcdefg"
			sigExpect := "96fc1eae890c8ee274b1464f2338736add94a43aa9909973465791682ec53d2814141747174f6e29ac779a837b9cba5b52547a3db5502d1742892289ac9ff7801c"
			sig, err := signService.Sign([]byte(oriBody))
			So(err, ShouldBeNil)
			So(sig, ShouldEqual, sigExpect)

			// verify
			verified := signService.Verify([]byte(oriBody), sig, walletName, appid, Secp256k1)
			So(verified, ShouldBeTrue)

			// if using unknown curve algorithm, still use secp256k1
			verified = signService.Verify([]byte(oriBody), sig, walletName, appid, "")
			So(verified, ShouldBeTrue)

			// if sig is invalid, just like "", should return false
			verified = signService.Verify([]byte(oriBody), "", walletName, appid, Secp256k1)
			So(verified, ShouldBeFalse)
		})

		Convey("Verify using ed25519 should be ok", func() {
			pk, k, err := ed25519.GenerateKey(rand.Reader)
			So(err, ShouldBeNil)

			walletName := ""
			appid := "m"

			signService, err := NewService("", map[string]string{walletName + "-" + appid: hex.EncodeToString(pk)})
			So(err, ShouldBeNil)

			message := []byte("123456789")
			signature := ed25519.Sign(k, message)

			verified := signService.Verify(message, hex.EncodeToString(signature), walletName, appid, Ed25519)
			So(verified, ShouldBeTrue)

			// set invalid public key length
			signService.pubKeys[walletName+"-"+appid] = []byte{1, 2, 3}
			verified = signService.Verify(message, hex.EncodeToString(signature), walletName, appid, Ed25519)
			So(verified, ShouldBeFalse)
		})

		Convey("Test verify with fat ed25519 public key", func() {
			walletName := ""
			appid := "phemex-stake-job"
			// provided by spot team
			base64String := "2JbOttkBCyhdALv9BTmPJV/lfJJm+02RyU8zQf1VytU="

			// 解码 Base64 字符串为二进制数据
			binaryData, err := base64.StdEncoding.DecodeString(base64String)
			if err != nil {
				fmt.Println("Error decoding base64 string:", err)
				return
			}

			// 将二进制数据转换为十六进制字符串
			pk := hex.EncodeToString(binaryData)
			fmt.Println("Fat public key", pk)

			signService, err := NewService("", map[string]string{walletName + "-" + appid: pk})
			So(err, ShouldBeNil)

			message := []byte("[{\"requestTime\":1703473619000,\"currency\":\"PT\",\"requestKey\":\"stake-6ba5b6cc6df34f35b86f1587474a0da5\",\"projectKey\":\"PT-STAKE\",\"amountRq\":10.33333333000000000000,\"opType\":\"MANUAL\",\"fundsType\":\"SPOT\",\"network\":\"optimism\",\"userId\":926311,\"address\":\"0x840f2dac237541f433fb7703fac82d0490393949\",\"expiryTime\":1764547200}]")
			signature := "34ddd4c409d19c2a4cfee91c22d778098c0ccf29eef9c17390e6999953bf1b119608795fe1b586b6ed2a2042dfcbbf6fdea9e242136287e1c2e35cd2211a4b0d"

			verified := signService.Verify(message, signature, walletName, appid, Ed25519)
			So(verified, ShouldBeTrue)
		})
	})
}
