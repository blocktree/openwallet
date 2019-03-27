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

package tron

import (
	"encoding/hex"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/blocktree/go-owcrypt"
)

//AddressDecoderStruct for Interface AddressDecoder
type AddressDecoderStruct struct {
	wm *WalletManager //钱包管理者
}

//NewAddressDecoder 地址解析器
func NewAddressDecoder(wm *WalletManager) *AddressDecoderStruct {
	decoder := AddressDecoderStruct{}
	decoder.wm = wm
	return &decoder
}

//PrivateKeyToWIF 私钥转WIF
func (decoder *AddressDecoderStruct) PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error) {
	return "", nil
}

//PublicKeyToAddress 公钥转地址
func (decoder *AddressDecoderStruct) PublicKeyToAddress(pub []byte, isTestnet bool) (string, error) {
	pubkey := owcrypt.PointDecompress(pub, owcrypt.ECC_CURVE_SECP256K1)
	hash := owcrypt.Hash(pubkey[1:], 0, owcrypt.HASH_ALG_KECCAK256)
	address, err := decoder.wm.CreateAddressRef(hash[12:], false) // isPrivate == false
	if err != nil {
		decoder.wm.Log.Info("creat address failed;unexpected error:%v", err)
		return "", err
	}
	return address, nil
}

//RedeemScriptToAddress 多重签名赎回脚本转地址
func (decoder *AddressDecoderStruct) RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error) {
	return "", nil
}

//WIFToPrivateKey WIF转私钥
func (decoder *AddressDecoderStruct) WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error) {

	return nil, nil

}

func DecodeAddress(addr string, isTestnet bool) (string, []byte, error) {
	codeType := addressEncoder.TRON_mainnetAddress
	if isTestnet {
		codeType = addressEncoder.TRON_testnetAddress
	}

	toAddressBytes, err := addressEncoder.AddressDecode(addr, codeType)
	if err != nil {
		return "", nil, err
	}
	toAddressBytes = append(codeType.Prefix(), toAddressBytes...)
	return hex.EncodeToString(toAddressBytes), toAddressBytes, nil
}

func EncodeAddress(hexStr string, isTestnet bool) (string, error) {

	codeType := addressEncoder.TRON_mainnetAddress
	if isTestnet {
		codeType = addressEncoder.TRON_testnetAddress
	}

	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	if len(b) > 20 {
		b = b[1:]
	}

	addr := addressEncoder.AddressEncode(b, codeType)
	return addr, nil
}
