package hdkeystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/openwallet/v2/crypto"
)

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	if plantText == nil || len(plantText) == 0 {
		return nil
	}
	length := len(plantText)
	unPadding := int(plantText[length-1])
	if length-unPadding <= 0 {
		return nil
	}
	return plantText[:(length - unPadding)]
}

func Aes256CBCEncrypt(plantText, key, iv []byte) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	if len(key) != 32 {
		return nil
	}
	if len(iv) != 16 {
		return nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())
	blockModel := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plantText))
	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext
}

func Aes256CBCDecrypt(ciphertext, key, iv []byte) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	if len(key) != 32 {
		return nil
	}
	if len(iv) != 16 {
		return nil
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plantText := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plantText, ciphertext)
	plantText = PKCS7UnPadding(plantText, block.BlockSize())
	if plantText == nil || len(plantText) == 0 {
		return nil
	}
	return plantText
}

// aesCBCDecryptHDKey 解密HDKey的文件内容
func aesCBCDecryptHDKey(keyProtected *encryptedHDKeyJSON, auth string) (keyBytes []byte, err error) {

	if keyProtected.Crypto.Cipher != CipherAes256CBC {
		return nil, fmt.Errorf("cipher not supported: %v", keyProtected.Crypto.Cipher)
	}

	mac, err := hex.DecodeString(keyProtected.Crypto.MAC)
	if err != nil {
		return nil, err
	}

	iv, err := hex.DecodeString(keyProtected.Crypto.CipherParams.IV)
	if err != nil {
		return nil, err
	}

	cipherText, err := hex.DecodeString(keyProtected.Crypto.CipherText)
	if err != nil {
		return nil, err
	}

	derivedKey, err := getKDFKey(keyProtected.Crypto, auth)
	if err != nil {
		return nil, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, ErrDecrypt
	}

	plainText := Aes256CBCDecrypt(cipherText, derivedKey[:32], iv)
	if len(plainText) == 0 {
		return nil, errors.New("aes decrypt invalid")
	}

	return plainText, nil
}
