package config

import (
	"encoding/base64"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

const (
	keyAliasFmt = "alias/%s-phemex-wallet-encryption"
)

var (
	cipher = Cipher{
		Decrypt: KmsDecrypt,
	}
)

func CipherDecoder(env string) mapstructure.DecodeHookFuncType {
	return func(f, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.String {
			return data, nil
		}
		raw := data.(string)
		if strings.HasPrefix(raw, "cipher:") {
			raw = strings.TrimPrefix(raw, "cipher:")
			d, err := base64.StdEncoding.DecodeString(raw)
			if err != nil {
				return "", err
			}
			return cipher.Decrypt(fmt.Sprintf(keyAliasFmt, env), d)
		}
		return data, nil
	}
}

type TLSConfig struct {
	TLSDisable    bool
	TLSCaFile     string
	TLSCertFile   string
	TLSKeyFile    string
	TLSMinVersion string
	ServerName    string
}

type ListenerType string

var (
	ListenerTypeTCP  = ListenerType("tcp")
	ListenerTypeUnix = ListenerType("unix")
)

type ServerListenerConfig struct {
	Type    ListenerType
	Address string

	TLSConfig `yaml:",inline" mapstructure:",squash"`
}
