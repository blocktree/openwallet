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

package tron

import (
	"encoding/hex"
	"errors"
	"log"

	"github.com/blocktree/go-OWCrypt"
	"github.com/shengdoushi/base58"
)

func CreateAddressByPkRef(pubKey []byte) (addrBytes []byte, err error) {
	// First: calculate sha3-256 of PublicKey, get Hash as pkHash
	pkHash := owcrypt.Hash(pubKey, 0, owcrypt.HASH_ALG_KECCAK256)[12:32]
	// Second: expend 0x41 as prefix of pkHash to mark Tron
	address := append([]byte{0x41}, pkHash...)
	// Third: double sha256 to generate Checksum
	sha256_0_1 := owcrypt.Hash(address, 0, owcrypt.HASh_ALG_DOUBLE_SHA256)
	// Fourth: Append checksum to pkHash from sha256_0_1 with the last 4
	addrBytes = append(address, sha256_0_1[0:4]...)

	return addrBytes, nil
}

// Done
// Function: Create address from a specified private key string
func (wm *WalletManager) CreateAddressRef(privateKey string) (addrBase58 string, err error) {

	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		log.Println(err)
		return "", err
	}

	pubKey, res := owcrypt.GenPubkey(privateKeyBytes, owcrypt.ECC_CURVE_SECP256K1)
	if res != owcrypt.SUCCESS {
		err := errors.New("Error from owcrypt.GenPubkey: failed!")
		log.Println(err)
		return "", err
	}

	if address, err := CreateAddressByPkRef(pubKey); err != nil {
		return "", err
	} else {
		// Last: encoding with Base58(alphabet use BitcoinAlphabet)
		addrBase58 = base58.Encode(address, base58.BitcoinAlphabet)
	}

	return addrBase58, nil
}

// Done
func (wm *WalletManager) ValidateAddressRef(addrBase58 string) (err error) {

	addressBytes, err := base58.Decode(addrBase58, base58.BitcoinAlphabet)
	if err != nil {
		return err
	}

	l := len(addressBytes)
	addressBytes, checksum := addressBytes[:l-4], addressBytes[l-4:]
	sha256_0_1 := owcrypt.Hash(addressBytes, 0, owcrypt.HASh_ALG_DOUBLE_SHA256)

	if hex.EncodeToString(sha256_0_1[0:4]) != hex.EncodeToString(checksum) {
		return errors.New("Address invalid!")
	}

	return nil
}
