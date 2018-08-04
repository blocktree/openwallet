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

type BlockHeader struct {
	Hash              string
	Confirmations     uint64
	Merkleroot        string
	Previousblockhash string
	Height            uint64
	Version           uint64
	Time              uint64
	Fork              bool
	Symbol            string
}

//BlockScanNotify 新区块扫描完成通知
//@param  txs 每扫完区块链，与地址相关的交易到
type BlockScanNotify func(header *BlockHeader)
