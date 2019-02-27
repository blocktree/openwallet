package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// AESEncrypt AES加密
// plantText 明文
// key密钥 16字节，24字节，32字节
// return 密文
func AESEncrypt(plantText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key) //选择加密算法
	if err != nil {
		return nil, err
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())

	blockModel := cipher.NewCBCEncrypter(block, key[:block.BlockSize()])

	ciphertext := make([]byte, len(plantText))

	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext, nil
}

// PKCS7Padding PKCS7填充
// ciphertext 明文
// blockSize 分组大小
// return 填充后的明文
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// AESDecrypt AES加密
// plantText 密码
// key密钥 16字节，24字节，32字节
// return 明文
func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key) //选择加密算法
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	plantText := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plantText, ciphertext)
	plantText = PKCS7UnPadding(plantText, block.BlockSize())
	return plantText, nil
}

// PKCS7UnPadding PKCS7清理填充
// ciphertext 填充后的明文
// blockSize 分组大小
// return 恢复后的明文
func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	if length - unpadding < 0 {
		return nil
	}
	return plantText[:(length - unpadding)]
}

/******** String扩展，获得AES加密方法 ********/

//AESProtocal String扩展，获得AES加密方法
type AESProtocal interface {
	AES(key string) (string, error)
	UnAES(aesBase64string string, key string) error
}
