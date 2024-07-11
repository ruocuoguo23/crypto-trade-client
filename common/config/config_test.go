package config

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	Convey("Test LoadConfig", t, func() {
		// Create a temporary YAML configuration file
		configContent := `
Chains:
  - name: "ethereum"
    url: "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID"
  - name: "solana"
    url: "https://tiniest-wandering-flower.solana-mainnet.quiknode.pro/3f2cf77b66958c08189f7d289df7d0740e554be2"
  - name: "optimism"
    url: "https://practical-green-butterfly.optimism.quiknode.pro/d02f8d49bde8ccbbcec3c9a8962646db998ade83"
`
		tmpFile, err := os.CreateTemp("", "config.yaml")
		So(err, ShouldBeNil)
		defer func(name string) {
			_ = os.Remove(name)
		}(tmpFile.Name())

		_, err = tmpFile.Write([]byte(configContent))
		So(err, ShouldBeNil)
		err = tmpFile.Close()
		So(err, ShouldBeNil)

		// Load the configuration from the temporary file
		chainMap, err := LoadConfig(tmpFile.Name())
		So(err, ShouldBeNil)

		// Verify the contents of the map
		So(chainMap, ShouldContainKey, "ethereum")
		So(chainMap["ethereum"].URL, ShouldEqual, "https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID")

		So(chainMap, ShouldContainKey, "solana")
		So(chainMap["solana"].URL, ShouldEqual, "https://tiniest-wandering-flower.solana-mainnet.quiknode.pro/3f2cf77b66958c08189f7d289df7d0740e554be2")

		So(chainMap, ShouldContainKey, "optimism")
		So(chainMap["optimism"].URL, ShouldEqual, "https://practical-green-butterfly.optimism.quiknode.pro/d02f8d49bde8ccbbcec3c9a8962646db998ade83")
	})
}
