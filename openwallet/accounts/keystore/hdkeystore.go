/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package keystore

import (
	"path/filepath"
	"github.com/ethereum/go-ethereum/accounts"
	"errors"
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
)


var (
	ErrLocked  = accounts.NewAuthNeededError("password or unlock")
	ErrNoMatch = errors.New("no key for given address or file")
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

type HDKeystore struct {
	keysDirPath string
	scryptN     int
	scryptP     int
}

// NewHDKeystore 实例化HDKeystore
func NewHDKeystore(keydir string, scryptN, scryptP int) *HDKeystore {
	keydir, _ = filepath.Abs(keydir)
	ks := &HDKeystore{keydir, scryptN, scryptP}
	return ks
}

//func (ks HDKeystore) GetKey(userKey string, filename, auth string) (*openwallet.UserAccount, error) {
//	// Load the key from the keystore and decrypt its contents
//	keyjson, err := ioutil.ReadFile(filename)
//	if err != nil {
//		return nil, err
//	}
//	key, err := DecryptHDKey(keyjson, auth)
//	if err != nil {
//		return nil, err
//	}
//	// Make sure we're really operating on the requested key (no swap attacks)
//	if key.us != addr {
//		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
//	}
//	return key, nil
//}
//
//// StoreKey generates a key, encrypts with 'auth' and stores in the given directory
//func StoreKey(dir, auth string, scryptN, scryptP int) (common.Address, error) {
//	_, a, err := storeNewKey(&HDKeystore{dir, scryptN, scryptP}, crand.Reader, auth)
//	return a.Address, err
//}
//
//func (ks keyStorePassphrase) StoreKey(filename string, key *Key, auth string) error {
//	keyjson, err := EncryptKey(key, auth, ks.scryptN, ks.scryptP)
//	if err != nil {
//		return err
//	}
//	return writeKeyFile(filename, keyjson)
//}
//
//func (ks keyStorePassphrase) JoinPath(filename string) string {
//	if filepath.IsAbs(filename) {
//		return filename
//	} else {
//		return filepath.Join(ks.keysDirPath, filename)
//	}
//}
