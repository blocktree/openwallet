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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/awnumar/memguard"
	"github.com/blocktree/openwallet/v2/crypto"
	"io"
)

// getArgon2KDFKey
func getArgon2KDFKey(cryptoJSON cryptoJSON, auth string) ([]byte, *argon2KDFParam, error) {

	argonParams := &argon2KDFParam{}

	if cryptoJSON.KDF == "argon2" { // 根据你存储的 kdf 字段值调整
		paramsBytes, err := json.Marshal(cryptoJSON.KDFParams)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal KDFParams: %w", err)
		}
		if err := json.Unmarshal(paramsBytes, argonParams); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal KDFParams into argon2KDFParam: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("unsupported KDF: %s", cryptoJSON.KDF)
	}

	authBytes := []byte(auth)

	defer ClearData(authBytes)

	salt, err := hex.DecodeString(argonParams.Salt)
	if err != nil {
		return nil, nil, err
	}

	return deriveKeyArgon2idDefault(authBytes, salt, argonParams.Time, argonParams.Memory, argonParams.Keylen, argonParams.Threads), argonParams, nil
}

// EncryptKeyByAes256GCMAndArgon2 encrypts a key using the specified scrypt parameters into a json
// blob that can be decrypted later on.
func EncryptKeyByAes256GCMAndArgon2(hdkey *HDKey, plainSeed []byte, auth string) ([]byte, error) {

	authBytes := []byte(auth)

	saltBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, saltBytes); err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}

	// 放弃旧版本的scrypt算法
	//derivedKey, err = scrypt.Key(authArray, saltBytes, scryptN, scryptR, scryptP, scryptDKLen)
	//if err != nil {
	//	return nil, err
	//}

	// 使用新版本的argon2算法
	derivedKey := DeriveKeyArgon2id(authBytes, saltBytes)

	defer ClearData(derivedKey, saltBytes, authBytes)

	salt := hex.EncodeToString(saltBytes)

	cipherText, err := AesGCMEncrypt(plainSeed, derivedKey, BuildAAD(hdkey.KeyID, hdkey.RootPath, salt, version, argon2Memory, argon2Time, argon2Threads, argon2KeyLen))
	if err != nil {
		return nil, err
	}
	if len(cipherText) == 0 {
		return nil, errors.New("aes encrypt result invalid")
	}

	kec := derivedKey[16:32]
	defer ClearData(kec)

	mac := crypto.Keccak256(kec, cipherText)

	kdfParam := argon2KDFParam{
		Memory:  argon2Memory,
		Time:    argon2Time,
		Threads: argon2Threads,
		Keylen:  argon2KeyLen,
		Salt:    salt,
	}

	cryptoStruct := cryptoJSON{
		Cipher:     CipherAes256GCM,
		CipherText: hex.EncodeToString(cipherText),
		KDF:        keyHeaderArgon2KDF,
		KDFParams:  kdfParam,
		MAC:        hex.EncodeToString(mac),
	}

	encryptedHDKeyJSON := encryptedHDKeyJSON{
		Alias:    hdkey.Alias,
		KeyID:    hdkey.KeyID,
		Crypto:   cryptoStruct,
		RootPath: hdkey.RootPath,
		Version:  version,
	}
	return json.MarshalIndent(encryptedHDKeyJSON, "", "\t")
}

// DecryptHDKeyByAes256GCMAndArgon2 decrypts a key from a json blob, returning the private key itself.
func DecryptHDKeyByAes256GCMAndArgon2(keyjson []byte, auth string) (*HDKey, error) {
	var k encryptedHDKeyJSON
	if err := json.Unmarshal(keyjson, &k); err != nil {
		return nil, err
	}

	// 1. 解密得到明文 seed
	seed, err := aesGCMAndArgon2DecryptHDKey(&k, auth)
	if err != nil {
		return nil, err
	}

	// 2. 立即使用 seed（在清零前）
	keyID := computeKeyID(seed)
	encrypted := encryptSeed(seed)

	// 3. 将加密后的种子放入锁定内存
	locker := memguard.NewBufferFromBytes(encrypted)

	// 4. 立即清零临时敏感数据
	ClearData(seed, encrypted)

	// 5. 返回安全的 HDKey
	return &HDKey{
		Alias:         k.Alias,
		RootPath:      k.RootPath,
		KeyID:         keyID,
		encryptedSeed: locker,
	}, nil
}
