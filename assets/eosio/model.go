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

package eosio

import (
	openwallet "github.com/blocktree/openwallet/openwallet"
	eos "github.com/eoscanada/eos-go"
)

// Block model
type Block eos.SignedBlock

//UnscanRecord 扫描失败的区块及交易
type UnscanRecord struct {
	ID          string `storm:"id"` // primary key
	BlockHeight uint64
	TxID        string
	Reason      string
}

// OWHeader 区块链头
func (b *Block) OWHeader() *openwallet.BlockHeader {
	obj := openwallet.BlockHeader{}

	hash, _ := b.BlockID()
	height := b.BlockNumber()

	//解析josn
	obj.Merkleroot = b.TransactionMRoot.String()
	obj.Hash = hash.String()
	obj.Previousblockhash = b.Previous.String()
	obj.Height = uint64(height)
	obj.Time = uint64(b.Timestamp.Unix())
	obj.Symbol = Symbol
	return &obj
}
