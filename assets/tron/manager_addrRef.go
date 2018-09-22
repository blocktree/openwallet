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

	// "github.com/blocktree/OpenWallet/assets/tron/protocol/api"

	"github.com/blocktree/go-OWCrypt"

	"github.com/shengdoushi/base58"
)

// Function: Create address from a specified password string (NOT PRIVATE KEY)
// Demo: curl -X POST http://127.0.0.1:8090/wallet/createaddress -d ‘
// 	{“value”: “7465737470617373776f7264” }’
// Parameters:
// 	value is the password, converted from ascii to hex
// Return value：
// 	value is the corresponding address for the password, encoded in hex.
// 	Convert it to base58 to use as the address.
// Warning:
// 	Please control risks when using this API. To ensure environmental security, please do not invoke APIs provided by other or invoke this very API on a public network.
func (wm *WalletManager) CreateAddressRef(passValue string) (addr string, err error) {

	privateKeyBytes, err := hex.DecodeString(passValue)
	if err != nil {

	}

	pubkey, res := owcrypt.GenPubkey(privateKeyBytes, owcrypt.ECC_CURVE_SECP256K1)
	if res != owcrypt.SUCCESS {
		return "", errors.New("Chao return Error!")
	}

	pkHash := owcrypt.Hash(pubkey, 0, owcrypt.HASH_ALG_KECCAK256)[12:32]
	address := append([]byte{0x41}, pkHash...)

	sha256_0_1 := owcrypt.Hash(address, 0, owcrypt.HASh_ALG_DOUBLE_SHA256)

	address = append(address, sha256_0_1[0:4]...)

	addr = base58.Encode(address, base58.BitcoinAlphabet)

	return addr, nil
}

func (wm *WalletManager) ValidateAddressRef(address string) (err error) {

	addressBytes, err := base58.Decode(address, base58.BitcoinAlphabet)
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
