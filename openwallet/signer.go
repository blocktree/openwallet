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

import "fmt"

//TransactionSigner 交易签署器
type TransactionSigner interface {

	// SignTransactionHash 交易哈希签名算法
	// required
	SignTransactionHash(msg []byte, privateKey []byte, eccType uint32) ([]byte, error)
}

type TransactionSignerBase struct {


}

// SignTransactionHash 交易哈希签名算法
// required
func (singer *TransactionSignerBase) SignTransactionHash(msg []byte, privateKey []byte, eccType uint32) ([]byte, error) {
	return nil, fmt.Errorf("SignTransactionHash not implement")
}