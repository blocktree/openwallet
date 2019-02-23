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

package virtualeconomy

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/blocktree/go-owcdrivers/addressEncoder"
	owcrypt "github.com/blocktree/go-owcrypt"
)

type addressDecoder struct {
	wm *WalletManager //钱包管理者
}

//NewAddressDecoder 地址解析器
func NewAddressDecoder(wm *WalletManager) *addressDecoder {
	decoder := addressDecoder{}
	decoder.wm = wm
	return &decoder
}

//PrivateKeyToWIF 私钥转WIF
func (decoder *addressDecoder) PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error) {
	return "", nil
}

//PublicKeyToAddress 公钥转地址
func (decoder *addressDecoder) PublicKeyToAddress(pub []byte, isTestnet bool) (string, error) {

	fmt.Println(hex.EncodeToString(pub))
	xpub, err := owcrypt.CURVE25519_convert_Ed_to_X(pub)
	fmt.Println(hex.EncodeToString(xpub))
	if err != nil {
		return "", errors.New("Invalid public key!")
	}

	cfg := addressEncoder.VSYS_mainnetAddress
	if decoder.wm.Config.IsTestNet {
		cfg = addressEncoder.VSYS_testnetAddress
	}

	pkHash := owcrypt.Hash(owcrypt.Hash(xpub, 32, owcrypt.HASH_ALG_BLAKE2B), 32, owcrypt.HASH_ALG_KECCAK256)[:20]

	address := addressEncoder.AddressEncode(pkHash, cfg)

	return address, nil
}

//RedeemScriptToAddress 多重签名赎回脚本转地址
func (decoder *addressDecoder) RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error) {
	return "", nil
}

//WIFToPrivateKey WIF转私钥
func (decoder *addressDecoder) WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error) {
	return nil, nil

}
