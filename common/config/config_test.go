package config

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestCipherDecoder(t *testing.T) {
	f := CipherDecoder("dev")
	cipher = Cipher{Decrypt: func(_ string, data []byte) (string, error) {
		return string(data), nil
	}}
	d := base64.StdEncoding.EncodeToString([]byte("ciphertext"))
	result, err := f(reflect.TypeOf(""), reflect.TypeOf(""), "cipher:"+d)
	require.Nil(t, err)
	assert.Equal(t, "ciphertext", result)
	result, err = f(reflect.TypeOf(""), reflect.TypeOf(""), "plaintext")
	require.Nil(t, err)
	assert.Equal(t, "plaintext", result)
}
