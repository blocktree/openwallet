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

package openwallet

import (
	"crypto/ecdsa"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	//地址首字节的标识
	AddressVersion = 0x48
	//地址协议头
	AddressProtocol = "openw:"
)

type Address common.Address

// String 把地址使用base58编码成字符串格式
func (a Address) String(addProtocol ...bool) string {
	s := base58.CheckEncode(a[:common.AddressLength], AddressVersion)
	if len(addProtocol) > 0 && addProtocol[0] {
		s = AddressProtocol + s
	}
	return s
}

//PubkeyToOpenwAddress 公钥转为openw统一地址
func PubkeyToAddress(p ecdsa.PublicKey) Address {
	pubBytes := crypto.FromECDSAPub(&p)
	pkHash := btcutil.Hash160(pubBytes)
	var a common.Address
	a.SetBytes(pkHash)
	return Address(a)
}

//ExtendedKeyToAddress 扩展密钥转地址
func ExtendedKeyToAddress(k *hdkeychain.ExtendedKey) Address {
	var a Address
	pubkey, err := k.ECPubKey()
	if err != nil {
		return a
	}
	return PubkeyToAddress(ecdsa.PublicKey(*pubkey))
}
