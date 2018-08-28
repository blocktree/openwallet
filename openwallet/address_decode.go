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

type AddressDecoder struct {

	//PrivateKeyToWIF 私钥转WIF
	PrivateKeyToWIF func(priv []byte, isTestnet bool) (string, error)
	//PublicKeyToAddress 公钥转地址
	PublicKeyToAddress func(pub []byte, isTestnet bool) (string, error)
	//WIFToPrivateKey WIF转私钥
	WIFToPrivateKey func(wif string, isTestnet bool) ([]byte, error)
}
