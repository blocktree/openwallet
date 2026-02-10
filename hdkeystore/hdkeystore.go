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
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/awnumar/memguard"
	"io/ioutil"
	"path/filepath"

	"github.com/blocktree/openwallet/v2/crypto/sha3"
)

const (
	CipherAes256GCM = "aes-256-gcm"

	keyHeaderKDF = "scrypt"

	keyHeaderArgon2idKDF = "argon2id"

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
	b, _ := GetRandomSecure(32)
	defer ClearData(b)
	return memguard.NewBufferFromBytes(b)
}

func getRuntimeAAD() *memguard.LockedBuffer {
	b, _ := GetRandomSecure(32)
	defer ClearData(b)
	return memguard.NewBufferFromBytes(b)
}

func encryptSeed(seed []byte, aadCall AADCall) (*memguard.LockedBuffer, error) {
	// 1. 计算 KeyID（只算一次）
	keyID := computeKeyID(seed)

	var aad []byte
	var err error

	// 2. 获取 AAD
	if aadCall != nil {
		aad, err = aadCall(keyID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate AAD: %w", err)
		}
		if len(aad) == 0 {
			return nil, errors.New("AAD must not be empty")
		}
	} else {
		aad = runtimeAAD.Data() // 注意：runtimeAAD 必须是 LockedBuffer
	}

	// 3. 执行加密
	encrypted, err := AesGCMEncrypt(seed, runtimeKey.Data(), aad)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %w", err)
	}
	defer ClearData(encrypted)

	// 4. 移入 LockedBuffer
	locker := memguard.NewBufferFromBytes(encrypted)
	return locker, nil
}

// decryptSeedTo 将解密结果直接写入 dst。
// keyID 用于重建 AAD；aadCall 可为 nil，表示使用默认 runtimeAAD。
func decryptSeedTo(dst []byte, encrypted []byte, keyID string, aadCall AADCall) error {
	var aad []byte
	var err error

	if aadCall != nil {
		aad, err = aadCall(keyID)
		if err != nil {
			return fmt.Errorf("failed to generate AAD for decryption: %w", err)
		}
		if len(aad) == 0 {
			return errors.New("AAD must not be empty")
		}
	} else {
		aad = runtimeAAD.Data()
	}

	return AesGCMDecryptToLocker(dst, encrypted, runtimeKey.Data(), aad)
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
func StoreLockerHDKey(dir, alias string, auth *memguard.LockedBuffer) (string, error) {
	seedBuf, err := GenerateLockedSeed(SeedLen)
	if err != nil {
		return "", err
	}
	defer seedBuf.Destroy()

	// 在单一作用域内使用明文
	var keyID string
	err = func() error {
		seed := seedBuf.Data() // ← 只调用一次！

		// 1. 计算 KeyID
		keyID = computeKeyID(seed)

		// 2. 构造元数据
		hdkeyMeta := &HDKey{
			Alias:    alias,
			KeyID:    keyID,
			RootPath: OpenwCoinTypePath,
		}

		// 3. 存储（传入 seed）
		ks := NewHDKeystore(dir, 0, 0)
		filePath := ks.JoinPath(KeyFileName(alias, keyID) + ".key")
		return ks.StoreLockerKeyWithSeed(filePath, hdkeyMeta, seed, auth)
	}()
	if err != nil {
		return "", fmt.Errorf("failed to store key: %w", err)
	}

	return keyID, nil
}

// StoreLockerKeyWithSeed 接收明文 seed（仅在锁定内存中有效）
func (ks *HDKeystore) StoreLockerKeyWithSeed(
	filename string,
	meta *HDKey,
	plainSeed []byte, // 来自 LockedBuffer.Data()
	auth *memguard.LockedBuffer,
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
func (ks HDKeystore) GetKey(rootId, filename string, auth *memguard.LockedBuffer) (*HDKey, error) {
	// Load the key from the keystore and decrypt its contents
	keyPath := ks.JoinPath(filename)
	keyjson, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return ks.GetKeyFromBase64(base64.StdEncoding.EncodeToString(keyjson), auth, nil)
}

// GetLockerKey 通过accountId读取钥匙
func (ks HDKeystore) GetLockerKey(path string, auth *memguard.LockedBuffer, aadCall AADCall) (*HDKey, error) {
	// Load the key from the keystore and decrypt its contents
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ks.GetKeyFromBase64(base64.StdEncoding.EncodeToString(keyjson), auth, aadCall)
}

func (ks HDKeystore) GetKeyFromBase64(keyJsonB64 string, auth *memguard.LockedBuffer, aadCall AADCall) (*HDKey, error) {
	// Load the key from the keystore and decrypt its contents
	keyjson, err := base64.StdEncoding.DecodeString(keyJsonB64)
	if err != nil {
		return nil, err
	}

	key, err := DecryptHDKeyByAes256GCMAndArgon2(keyjson, auth, aadCall)
	if err != nil {
		return nil, err
	}

	if key == nil || len(key.KeyID) == 0 {
		return nil, errors.New("HDKey decrypt invalid")
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

func (ks *HDKeystore) JoinDirPath(dir, filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	} else {
		return filepath.Join(dir, filename)
	}
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
