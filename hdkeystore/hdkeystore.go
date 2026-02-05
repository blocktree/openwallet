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
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/blocktree/openwallet/v2/crypto/sha3"
)

const (
	CipherAes256GCM = "aes-256-gcm"

	CipherAes128CTR = "aes-128-ctr"

	keyHeaderKDF = "scrypt"

	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = 1 << 18

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
	randomKey  = randomBytes(32)
	randomAAD  = randomBytes(32)
)

// HDKeystore HDKey的存粗工具类
type HDKeystore struct {
	keysDirPath string
	//MasterKey   string
	scryptN int
	scryptP int
	cipher  string
}

func randomBytes(l int) []byte {
	bs := make([]byte, l)
	if _, err := io.ReadFull(rand.Reader, bs); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return bs
}

func encryptSeed(seed []byte) []byte {
	encrypted, err := AesGCMEncrypt(seed, randomKey, randomAAD)
	if err != nil {
		panic(err)
	}
	return encrypted
}

func decryptSeed(seed []byte) []byte {
	decrypted, err := AesGCMDecrypt(seed, randomKey, randomAAD)
	if err != nil {
		panic(err)
	}
	return decrypted
}

// NewHDKeystore 实例化HDKeystore
func NewHDKeystore(keydir string, scryptN, scryptP int, cipher ...string) *HDKeystore {
	keydir, _ = filepath.Abs(keydir)
	ks := &HDKeystore{}
	ks.keysDirPath = keydir
	ks.scryptN = scryptN
	ks.scryptP = scryptP
	if len(cipher) > 0 {
		ks.cipher = cipher[0]
	}
	return ks
}

// StoreHDKey 创建HDKey
func StoreHDKey(dir, alias, auth string, scryptN, scryptP int, cipher ...string) (*HDKey, string, error) {

	seed, err := GenerateSeed(SeedLen)
	if err != nil {
		return nil, "", err
	}

	//extSeed, err := GetExtendSeed(seed, masterKey)
	//if err != nil {
	//	return "", err
	//}

	return StoreHDKeyWithSeed(dir, alias, auth, seed, scryptN, scryptP, cipher...)
}

// StoreHDKey 创建HDKey
func StoreHDKeyWithSeed(dir, alias, auth string, seed []byte, scryptN, scryptP int, cipher ...string) (*HDKey, string, error) {
	ks := NewHDKeystore(dir, scryptN, scryptP, cipher...)
	key, filePath, err := storeNewKey(ks, alias, auth, seed)
	return key, filePath, err
}

// storeNewKey 用随机种子生成HDKey
func storeNewKey(ks *HDKeystore, alias, auth string, seed []byte) (*HDKey, string, error) {

	key, err := NewHDKey(seed, alias, OpenwCoinTypePath)
	if err != nil {
		return nil, "", err
	}
	filePath := ks.JoinPath(KeyFileName(key.Alias, key.KeyID) + ".key")
	ks.StoreKey(filePath, key, auth)
	return key, filePath, err
}

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

	var key *HDKey
	if ks.cipher == CipherAes256GCM {
		key, err = DecryptHDKeyByAes256GCM(keyjson, auth)
		if err != nil {
			return nil, err
		}
	} else if ks.cipher == CipherAes128CTR {
		key, err = DecryptHDKey(keyjson, auth)
		if err != nil {
			return nil, err
		}
	} else {
		key, err = DecryptHDKey(keyjson, auth)
		if err != nil {
			return nil, err
		}
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

// StoreKey 把HDKey重写加密写入到文件中
func (ks *HDKeystore) StoreKey(filename string, key *HDKey, auth string) error {
	var keyjson []byte
	var err error
	if ks.cipher == CipherAes256GCM {
		keyjson, err = EncryptKeyByAes256GCM(key, auth, ks.scryptN, ks.scryptP)
		if err != nil {
			return err
		}
		return writeKeyFile(filename, keyjson)
	} else if ks.cipher == CipherAes128CTR {
		keyjson, err = EncryptKey(key, auth, ks.scryptN, ks.scryptP)
		if err != nil {
			return err
		}
	} else {
		keyjson, err = EncryptKey(key, auth, ks.scryptN, ks.scryptP)
		if err != nil {
			return err
		}
	}

	return writeKeyFile(filename, keyjson)
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
