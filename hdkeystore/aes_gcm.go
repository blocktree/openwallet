package hdkeystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/blocktree/openwallet/v2/crypto"
	"io"
)

// GetRandomSecure 使用加密安全的随机数生成器生成指定字节数组（推荐）
func GetRandomSecure(l int) ([]byte, error) {
	randomIV := make([]byte, l)
	if _, err := io.ReadFull(rand.Reader, randomIV); err != nil {
		return nil, err
	}
	return randomIV, nil
}

// AesGCMEncrypt AES-GCM 加密基础方法
// 返回格式：Byte(Nonce + Ciphertext + AuthTag)
// Nonce: 12 字节（GCM 标准）
// AuthTag: 16 字节（128-bit 认证标签）
func AesGCMEncrypt(plaintext, key, additionalData []byte) ([]byte, error) {
	// 1. 输入验证
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	// 2. 创建 AES 加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 3. 创建 GCM 模式（AEAD: Authenticated Encryption with Associated Data）
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 4. 生成 Nonce（GCM 标准：12 字节）
	// 注意：Nonce 必须唯一，否则会破坏 GCM 安全性
	nonce, err := GetRandomSecure(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	// 5. 加密并生成认证标签（单步操作）
	// Seal 会自动：
	//   - 加密 plaintext
	//   - 对 ciphertext + additionalData 生成 GMAC 认证标签
	//   - 返回：ciphertext + authTag
	ciphertext := gcm.Seal(nil, nonce, plaintext, additionalData)

	// 6. 拼接：Nonce + Ciphertext + AuthTag
	result := append(nonce, ciphertext...)

	// 7. Base64 编码
	return result, nil
}

// AesGCMDecrypt AES-GCM 解密基础方法
// 会自动验证：
// 1. 认证标签（AuthTag）- 确保密文未被篡改
// 2. 附加认证数据（AAD）- 确保关联数据未被篡改
// 任何一项验证失败都会返回错误，拒绝解密
func AesGCMDecrypt(encryptedData, key, additionalData []byte) ([]byte, error) {
	// 1. 输入验证
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}

	// 2. 解码 Base64
	data := encryptedData
	if data == nil {
		return nil, errors.New("base64 decode failed")
	}

	// 3. 创建 AES 加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// 4. 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// 5. 检查数据长度（Nonce + Ciphertext + AuthTag）
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	// 6. 分离 Nonce 和 Ciphertext+AuthTag
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	// 7. 解密并验证认证标签（单步操作）
	// Open 会自动：
	//   - 验证 GMAC 认证标签（防篡改）
	//   - 验证 additionalData（如果提供）
	//   - 解密 ciphertext
	// 任何验证失败都会返回 error
	plaintext, err := gcm.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		// 🚨 这个错误非常重要！表示数据被篡改或密钥错误
		return nil, fmt.Errorf("authentication failed - data may be tampered: %w", err)
	}

	return plaintext, nil
}

// AesGCMDecryptToLocker 安全解密，将明文直接写入 dst，避免临时分配
// dst 必须足够大以容纳明文（最大 len(encryptedData) - 12 - 16）
func AesGCMDecryptToLocker(dst, encryptedData, key, additionalData []byte) error {
	// 1. 输入验证
	if len(key) != 32 {
		return errors.New("key must be 32 bytes for AES-256")
	}
	if encryptedData == nil {
		return errors.New("encrypted data is nil")
	}

	// 2. 创建 AES 解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// 3. 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// 4. 检查长度
	nonceSize := gcm.NonceSize()
	overhead := gcm.Overhead() // 通常是 16 字节（AuthTag）
	if len(encryptedData) < nonceSize+overhead {
		return errors.New("encrypted data too short")
	}

	// 5. 分离 Nonce 和 Ciphertext+Tag
	nonce := encryptedData[:nonceSize]
	ciphertextWithTag := encryptedData[nonceSize:]

	// 6. 关键：使用 dst 作为输出缓冲区
	// gcm.Open(dst[:0], ...) 会将明文追加到 dst 起始位置
	_, err = gcm.Open(dst[:0], nonce, ciphertextWithTag, additionalData)
	if err != nil {
		return fmt.Errorf("authentication failed - data may be tampered: %w", err)
	}

	return nil
}

// aesGCMDecryptHDKey 解密HDKey的文件内容
func aesGCMDecryptHDKey(keyProtected *encryptedHDKeyJSON, auth string) (keyBytes []byte, err error) {

	if keyProtected.Crypto.Cipher != CipherAes256GCM {
		return nil, fmt.Errorf("cipher not supported: %v", keyProtected.Crypto.Cipher)
	}

	mac, err := hex.DecodeString(keyProtected.Crypto.MAC)
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

	defer ClearData(derivedKey)

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, ErrDecrypt
	}

	scryptN := keyProtected.Crypto.KDFParams["n"]
	scryptP := keyProtected.Crypto.KDFParams["p"]
	salt := keyProtected.Crypto.KDFParams["salt"]

	if _, b := scryptN.(float64); !b {
		return nil, errors.New("scrypt n invalid")
	}

	if _, b := scryptP.(float64); !b {
		return nil, errors.New("scrypt p invalid")
	}

	if _, b := salt.(string); !b {
		return nil, errors.New("scrypt salt invalid")
	}

	aadStr := fmt.Sprintf(
		"keyid:%s|rootpath:%s|version:%d|cipher:%s|scrypt_n:%d|scrypt_r:%d|scrypt_p:%d|scrypt_dklen:%d|scrypt_salt:%s",
		keyProtected.KeyID,     // 保留：钱包唯一标识（不可改）
		keyProtected.RootPath,  // 保留：钱包路径（不可改）
		keyProtected.Version,   // 保留：版本号
		CipherAes256GCM,        // 保留：加密算法
		int(scryptN.(float64)), // 保留：scrypt参数
		scryptR,                // 保留：scrypt参数
		int(scryptP.(float64)), // 保留：scrypt参数
		scryptDKLen,            // 保留：scrypt参数
		salt,                   // 保留：scrypt盐
		// 可选：若Alias不可改，添加 |alias:%s", hdkey.Alias
	)

	plainText, err := AesGCMDecrypt(cipherText, derivedKey, []byte(aadStr))
	if err != nil {
		return nil, err
	}
	if len(plainText) == 0 {
		return nil, errors.New("aes gcm decrypt invalid")
	}

	return plainText, nil
}
