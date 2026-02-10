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
	"io"
)

// getArgon2KDFKey
func getArgon2KDFKey(cryptoJSON cryptoJSON, auth *memguard.LockedBuffer) ([]byte, *argon2KDFParam, error) {

	argonParams := &argon2KDFParam{}

	if cryptoJSON.KDF == keyHeaderArgon2idKDF { // 根据你存储的 kdf 字段值调整
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

	salt, err := hex.DecodeString(argonParams.Salt)
	if err != nil {
		return nil, nil, err
	}

	return deriveKeyArgon2idDefault(auth.Data(), salt, uint32(argonParams.Time), uint32(argonParams.Memory), uint32(argonParams.Keylen), uint8(argonParams.Threads)), argonParams, nil
}

// EncryptKeyByAes256GCMAndArgon2 encrypts a key using the specified scrypt parameters into a json
// blob that can be decrypted later on.
func EncryptKeyByAes256GCMAndArgon2(hdkey *HDKey, plainSeed []byte, auth *memguard.LockedBuffer) ([]byte, error) {

	authBytes := auth.Data()

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

	defer ClearData(derivedKey, saltBytes)

	salt := hex.EncodeToString(saltBytes)

	cipherText, err := AesGCMEncrypt(plainSeed, derivedKey, BuildAAD(hdkey.KeyID, hdkey.RootPath, salt, version, argon2Memory, argon2Time, argon2Threads, argon2KeyLen))
	if err != nil {
		return nil, err
	}
	if len(cipherText) == 0 {
		return nil, errors.New("aes encrypt result invalid")
	}

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
		KDF:        keyHeaderArgon2idKDF,
		KDFParams:  kdfParam,
		//MAC:        hex.EncodeToString(mac),
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
func DecryptHDKeyByAes256GCMAndArgon2(keyjson []byte, auth *memguard.LockedBuffer, aadCall AADCall) (*HDKey, error) {
	var k encryptedHDKeyJSON
	if err := json.Unmarshal(keyjson, &k); err != nil {
		return nil, err
	}

	// 1. 解密得到 seed（在 LockedBuffer 中）
	seedBuf, err := aesGCMAndArgon2DecryptHDKey(&k, auth)
	if err != nil {
		return nil, err
	}
	defer seedBuf.Destroy()

	// 2. 在单一回调中完成所有明文操作（最佳实践）
	var (
		keyID        string
		encryptedBuf *memguard.LockedBuffer
	)

	err = func() error {
		seed := seedBuf.Data() // 仅在此作用域内使用

		// 验证 KeyID
		keyID = computeKeyID(seed)
		if keyID != k.KeyID {
			return errors.New("key ID mismatch after decryption")
		}

		var err error
		encryptedBuf, err = encryptSeed(seed, aadCall)
		return err
	}()
	if err != nil {
		return nil, err
	}

	return &HDKey{
		Alias:         k.Alias,
		RootPath:      k.RootPath,
		KeyID:         keyID,
		encryptedSeed: encryptedBuf, // ← 直接赋值
		aadCall:       aadCall,
	}, nil
}
