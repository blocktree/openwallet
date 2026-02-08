/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package hdkeystore

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/awnumar/memguard"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/blocktree/openwallet/v2/crypto/sha3"
)

const (
	CipherAes256GCM = "aes-256-gcm"

	keyHeaderKDF = "scrypt"

	keyHeaderArgon2IDKDF = "argon2id"

	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = 1 << 18

	// HighSecurityScryptN is the N parameter for Scrypt with higher security,
	// using approximately 1GB memory and taking a few seconds on a modern CPU.
	HighSecurityScryptN = 1 << 20

	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptP = 1

	// LightScryptN is the N parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptN = 1 << 12

	// LightScryptP is the P parameter of Scrypt encryption algorithm, using 4MB
	// memory and taking approximately 100ms CPU time on a modern processor.
	LightScryptP = 6

	scryptR     = 8
	scryptDKLen = 32

	//种子长度
	SeedLen = 64
)

var (
	//ErrLocked  = accounts.NewAuthNeededError("password or unlock")
	ErrNoMatch = errors.New("no key for given address or file")
	//ErrDecrypt 机密出错
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
	runtimeKey = getRuntimeKey()
	runtimeAAD = getRuntimeAAD()
)

// HDKeystore HDKey的存粗工具类
type HDKeystore struct {
	keysDirPath string
	//MasterKey   string
	scryptN int
	scryptP int
}

func getRuntimeKey() *memguard.LockedBuffer {
	b := randomBytes(32)
	defer ClearData(b)
	return memguard.NewBufferFromBytes(b)
}

func getRuntimeAAD() *memguard.LockedBuffer {
	b := randomBytes(32)
	defer ClearData(b)
	return memguard.NewBufferFromBytes(b)
}

func randomBytes(l int) []byte {
	bs := make([]byte, l)
	if _, err := io.ReadFull(rand.Reader, bs); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return bs
}

func encryptSeed(seed []byte) []byte {
	encrypted, err := AesGCMEncrypt(seed, runtimeKey.Data(), runtimeAAD.Data())
	if err != nil {
		panic(err)
	}
	return encrypted
}

// decryptSeedTo 将解密结果直接写入目标 buffer（不返回明文 slice）
func decryptSeedTo(dst []byte, encrypted []byte) error {
	return AesGCMDecryptToLocker(dst, encrypted, runtimeKey.Data(), runtimeAAD.Data())
}

// NewHDKeystore 实例化HDKeystore
func NewHDKeystore(keydir string, scryptN, scryptP int) *HDKeystore {
	keydir, _ = filepath.Abs(keydir)
	ks := &HDKeystore{}
	ks.keysDirPath = keydir
	ks.scryptN = scryptN
	ks.scryptP = scryptP
	return ks
}

// StoreLockerHDKey 重要：当前版本只使用这个创建钱包文件入口，使用AES-256-GCM保存文件
func StoreLockerHDKey(dir, alias, auth string) (string, error) {
	seed, err := GenerateLockedSeed(SeedLen)
	if err != nil {
		return "", err
	}
	defer seed.Destroy()

	// 2. 计算 KeyID（需明文）
	keyID := computeKeyID(seed.Data())

	// 3. 创建 keystore
	ks := NewHDKeystore(dir, 0, 0)

	// 4. 构造 HDKey 元数据（不含明文种子！）
	hdkeyMeta := &HDKey{
		Alias:    alias,
		KeyID:    keyID,
		RootPath: OpenwCoinTypePath,
		// 注意：encryptedSeed 暂不设置（或设为 nil）
	}

	// 5. 直接用 seed.Data() 加密并存储
	filePath := ks.JoinPath(KeyFileName(alias, keyID) + ".key")
	if err := ks.StoreLockerKeyWithSeed(filePath, hdkeyMeta, seed.Data(), auth); err != nil {
		return "", fmt.Errorf("failed to store key: %w", err)
	}

	return keyID, nil
}

// StoreLockerKeyWithSeed 接收明文 seed（仅在锁定内存中有效）
func (ks *HDKeystore) StoreLockerKeyWithSeed(
	filename string,
	meta *HDKey,
	plainSeed []byte, // 来自 LockedBuffer.Data()
	auth string,
) error {
	// mode=0默认的argon2派生密钥
	keyJSON, err := EncryptKeyByAes256GCMAndArgon2(meta, plainSeed, auth)
	if err != nil {
		return err
	}
	return writeKeyFile(filename, keyJSON)
}

// StoreHDKeyWithSeed 创建HDKey
//func StoreHDKeyWithSeed(dir, alias, auth string, seed []byte, scryptN, scryptP int) (*HDKey, string, error) {
//	ks := NewHDKeystore(dir, scryptN, scryptP)
//	key, filePath, err := storeNewKey(ks, alias, auth, seed)
//	return key, filePath, err
//}
//
//// storeNewKey 用随机种子生成HDKey
//func storeNewKey(ks *HDKeystore, alias, auth string, seed []byte) (*HDKey, string, error) {
//
//	key, err := NewHDKey(seed, alias, OpenwCoinTypePath)
//	if err != nil {
//		return nil, "", err
//	}
//	filePath := ks.JoinPath(KeyFileName(key.Alias, key.KeyID) + ".key")
//	ks.StoreKey(filePath, key, auth)
//	return key, filePath, err
//}

// GetKey 通过accountId读取钥匙
func (ks HDKeystore) GetKey(rootId, filename, auth string) (*HDKey, error) {
	// Load the key from the keystore and decrypt its contents
	keyPath := ks.JoinPath(filename)
	keyjson, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return ks.GetKeyFromBase64(rootId, base64.StdEncoding.EncodeToString(keyjson), auth)
}

func (ks HDKeystore) GetKeyFromBase64(rootId, keyJsonB64, auth string) (*HDKey, error) {
	// Load the key from the keystore and decrypt its contents
	keyjson, err := base64.StdEncoding.DecodeString(keyJsonB64)
	if err != nil {
		return nil, err
	}

	key, err := DecryptHDKeyByAes256GCMAndArgon2(keyjson, auth)
	if err != nil {
		return nil, err
	}

	if key == nil || len(key.KeyID) == 0 {
		return nil, errors.New("HDKey decrypt invalid")
	}

	if len(rootId) > 0 {
		// Make sure we're really operating on the requested key (no swap attacks)
		if key.KeyID != rootId {
			return nil, fmt.Errorf("key content mismatch: have account %s, want %s", key.KeyID, rootId)
		}
	}

	return key, nil
}

// JoinPath 文件路径组合
func (ks *HDKeystore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	} else {
		return filepath.Join(ks.keysDirPath, filename)
	}
}

// getDecryptedKey 获取解密后的钥匙
func (ks *HDKeystore) getDecryptedKey(alias, rootId, auth string) (*HDKey, error) {
	path := ks.JoinPath(KeyFileName(alias, rootId) + ".key")
	key, err := ks.GetKey(rootId, path, auth)
	return key, err
}

// GetExtendSeed 获得某个币种的扩展种子
func GetExtendSeed(seed []byte, masterKey string) ([]byte, error) {

	if len(seed) < MinSeedBytes || len(seed) > MaxSeedBytes {
		return nil, ErrInvalidSeedLen
	}

	hmac256 := hmac.New(sha3.New256, []byte(masterKey))
	hmac256.Write(seed)
	ext := hmac256.Sum(nil)
	return ext, nil
}
