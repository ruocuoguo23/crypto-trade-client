package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"os"
)

func main() {
	// 替换为你的 keystore 文件路径和密码
	keystoreFile := "/Users/jeff.wu/Downloads/77f15ce8b500cd20da5ac4be09c373932500ebeb.keystore"
	password := "wuyang2012"

	data, err := os.ReadFile(keystoreFile)
	if err != nil {
		fmt.Println("Error reading keystore file:", err)
		return
	}

	// 解析 keystore 文件
	key, err := keystore.DecryptKey(data, password)
	if err != nil {
		fmt.Println("Error decrypting key:", err)
		return
	}

	// 输出私钥的十六进制表示
	fmt.Printf("Private key in hex: %x\n", key.PrivateKey.D.Bytes())
}
