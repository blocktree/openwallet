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

package keystore

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/blocktree/openwallet/log"
	"github.com/btcsuite/btcutil/hdkeychain"
)

const (
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
	SeedLen = 32
)

var (
	//ErrLocked  = accounts.NewAuthNeededError("password or unlock")
	ErrNoMatch = errors.New("no key for given address or file")
	//ErrDecrypt 机密出错
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

//HDKeystore HDKey的存粗工具类
type HDKeystore struct {
	keysDirPath string
	//MasterKey   string
	scryptN int
	scryptP int
}

// NewHDKeystore 实例化HDKeystore
func NewHDKeystore(keydir string, scryptN, scryptP int) *HDKeystore {
	keydir, _ = filepath.Abs(keydir)
	ks := &HDKeystore{keydir, scryptN, scryptP}
	return ks
}

// StoreHDKey 创建HDKey
// Deprecated: hdkeystore.StoreHDKey instead.
func StoreHDKey(dir, alias, auth string, scryptN, scryptP int) (*HDKey, string, error) {

	seed, err := hdkeychain.GenerateSeed(SeedLen)
	if err != nil {
		return nil, "", err
	}

	//extSeed, err := GetExtendSeed(seed, masterKey)
	//if err != nil {
	//	return "", err
	//}

	return StoreHDKeyWithSeed(dir, alias, auth, seed, scryptN, scryptP)
}

// StoreHDKey 创建HDKey
func StoreHDKeyWithSeed(dir, alias, auth string, seed []byte, scryptN, scryptP int) (*HDKey, string, error) {
	key, filePath, err := storeNewKey(&HDKeystore{dir, scryptN, scryptP}, alias, auth, seed)
	return key, filePath, err
}

//storeNewKey 用随机种子生成HDKey
func storeNewKey(ks *HDKeystore, alias, auth string, seed []byte) (*HDKey, string, error) {

	key, err := NewHDKey(seed, alias, OpenwCoinTypePath)
	if err != nil {
		return nil, "", err
	}
	filePath := ks.JoinPath(keyFileName(key.Alias, key.RootId) + ".key")
	log.Debugf("filepath:%v", filePath)
	ks.StoreKey(filePath, key, auth)
	return key, filePath, err
}

//GetKey 通过accountId读取钥匙
func (ks HDKeystore) GetKey(rootId, filename, auth string) (*HDKey, error) {
	// Load the key from the keystore and decrypt its contents
	keyPath := ks.JoinPath(filename)
	keyjson, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	key, err := DecryptHDKey(keyjson, auth)
	if err != nil {
		return nil, err
	}

	if len(rootId) > 0 {
		// Make sure we're really operating on the requested key (no swap attacks)
		if key.RootId != rootId {
			return nil, fmt.Errorf("key content mismatch: have account %s, want %s", key.RootId, rootId)
		}
	}

	return key, nil
}

//StoreKey 把HDKey重写加密写入到文件中
func (ks *HDKeystore) StoreKey(filename string, key *HDKey, auth string) error {
	keyjson, err := EncryptKey(key, auth, ks.scryptN, ks.scryptP)
	if err != nil {
		return err
	}
	return writeKeyFile(filename, keyjson)
}

//JoinPath 文件路径组合
func (ks *HDKeystore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	} else {
		return filepath.Join(ks.keysDirPath, filename)
	}
}

//getDecryptedKey 获取解密后的钥匙
func (ks *HDKeystore) getDecryptedKey(alias, rootId, auth string) (*HDKey, error) {
	path := ks.JoinPath(keyFileName(alias, rootId) + ".key")
	key, err := ks.GetKey(rootId, path, auth)
	return key, err
}

//GetExtendSeed 获得某个币种的扩展种子
func GetExtendSeed(seed []byte, masterKey string) ([]byte, error) {

	if len(seed) < hdkeychain.MinSeedBytes || len(seed) > hdkeychain.MaxSeedBytes {
		return nil, hdkeychain.ErrInvalidSeedLen
	}

	hmac256 := hmac.New(sha256.New, []byte(masterKey))
	hmac256.Write(seed)
	ext := hmac256.Sum(nil)
	return ext, nil
}
