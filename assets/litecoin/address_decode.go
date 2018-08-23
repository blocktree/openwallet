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

package litecoin

import (
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
	"github.com/blocktree/go-OWCrypt"
)

func init() {

}

var (
	AddressDecoder = &openwallet.AddressDecoder{
		PrivateKeyToWIF:    PrivateKeyToWIF,
		PublicKeyToAddress: PublicKeyToAddress,
		WIFToPrivateKey:    WIFToPrivateKey,
	}
)

//PrivateKeyToWIF 私钥转WIF
func PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error) {

	cfg := addressEncoder.LTC_mainnetPrivateWIFCompressed
	if isTestnet {
		cfg = addressEncoder.LTC_testnetPrivateWIFCompressed
	}

	//privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	//wif, err := btcutil.NewWIF(privateKey, &cfg, true)
	//if err != nil {
	//	return "", err
	//}

	wif := addressEncoder.AddressEncode(priv, cfg)

	return wif, nil

}

//PublicKeyToAddress 公钥转地址
func PublicKeyToAddress(pub []byte, isTestnet bool) (string, error) {

	cfg := addressEncoder.LTC_mainnetAddressP2PKH
	if isTestnet {
		cfg = addressEncoder.LTC_testnetAddressP2PKH
	}

	//pkHash := btcutil.Hash160(pub)
	//address, err :=  btcutil.NewAddressPubKeyHash(pkHash, &cfg)
	//if err != nil {
	//	return "", err
	//}

	pkHash := owcrypt.Hash(pub, 0, owcrypt.HASH_ALG_HASH160)

	address := addressEncoder.AddressEncode(pkHash, cfg)

	return address, nil

}

//WIFToPrivateKey WIF转私钥
func WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error) {

	cfg := addressEncoder.LTC_mainnetPrivateWIFCompressed
	if isTestnet {
		cfg = addressEncoder.LTC_testnetPrivateWIFCompressed
	}

	priv, err := addressEncoder.AddressDecode(wif, cfg)
	if err != nil {
		return nil, err
	}

	return priv, err

}
