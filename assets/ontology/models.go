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

package ontology

import (
	"fmt"

	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tidwall/gjson"
)

// type Vin struct {
// 	Coinbase string
// 	TxID     string
// 	Vout     uint64
// 	N        uint64
// 	Addr     string
// 	Value    string
// }

// type Vout struct {
// 	N            uint64
// 	Addr         string
// 	Value        string
// 	ScriptPubKey string
// 	Type         string
// }

type Block struct {

	/*

		"hash": "000000000000000127454a8c91e74cf93ad76752cceb7eb3bcff0c398ba84b1f",
		"confirmations": 2,
		"strippedsize": 191875,
		"size": 199561,
		"weight": 775186,
		"height": 1354760,
		"version": 536870912,
		"versionHex": "20000000",
		"merkleroot": "48239e76f8b37d9c8824fef93d42ac3d7c433029c1e9fa23b6416dd0356f3e57",
		"tx": ["c1e12febeb58aefb0b01c04360262138f4ee0faeb207276e79ea3866608ed84f"]
		"time": 1532143012,
		"mediantime": 1532140298,
		"nonce": 3410287696,
		"bits": "19499855",
		"difficulty": 58358570.79038175,
		"chainwork": "00000000000000000000000000000000000000000000006f68c43926cd6c2d1f",
		"previousblockhash": "00000000000000292d142fcc1ddbd9dafd4518310009f152bdca2a66cc589f97",
		"nextblockhash": "0000000000004a50ef5733ab333f718e6ef5c1995e2cfd5a7caa0875f118cd30"

	*/

	Hash             string
	Size             uint64
	Version          byte
	PrevBlockHash    string
	TransactionsRoot string
	BlockRoot        string
	Timestamp        uint64
	Height           uint64
	ConsensusData    string
	ConsensusPayload string
	NextBookkeeper   string
	Bookkeepers      []string
	SigData          []string
	Transactions     []string
}

type Transaction struct {
	TxID        string
	Version     byte
	Nonce       uint32
	GasPrice    uint64
	GasLimit    uint64
	Payer       string
	TxType      byte
	Payload     string
	BlockHeight uint64
	BlockHash   string
}

func NewBlock(json *gjson.Result) *Block {
	obj := &Block{}

	fmt.Println(json.Raw)
	// 解析
	obj.Hash = gjson.Get(json.Raw, "Hash").String()
	obj.Size = gjson.Get(json.Raw, "Size").Uint()
	obj.Version = byte(gjson.Get(json.Raw, "Header").Get("Version").Uint())
	obj.PrevBlockHash = gjson.Get(json.Raw, "Header").Get("PrevBlockHash").String()
	obj.TransactionsRoot = gjson.Get(json.Raw, "Header").Get("TransactionsRoot").String()
	obj.BlockRoot = gjson.Get(json.Raw, "Header").Get("BlockRoot").String()
	obj.Timestamp = gjson.Get(json.Raw, "Header").Get("Timestamp").Uint()
	obj.Height = gjson.Get(json.Raw, "Header").Get("Height").Uint()
	obj.ConsensusData = gjson.Get(json.Raw, "Header").Get("ConsensusData").String()
	obj.ConsensusPayload = gjson.Get(json.Raw, "Header").Get("ConsensusPayload").String()
	obj.NextBookkeeper = gjson.Get(json.Raw, "Header").Get("NextBookkeeper").String()
	bookkeepers := gjson.Get(json.Raw, "Header").Get("Bookkeepers").Array()
	for _, bk := range bookkeepers {
		obj.Bookkeepers = append(obj.Bookkeepers, bk.String())
	}
	sigdatas := gjson.Get(json.Raw, "Header").Get("SigData").Array()
	for _, sd := range sigdatas {
		obj.SigData = append(obj.SigData, sd.String())
	}

	txs := gjson.Get(json.Raw, "Transactions").Array()

	for _, tx := range txs {
		obj.Transactions = append(obj.Transactions, gjson.Get(tx.Raw, "Hash").String())
	}

	return obj
}

//BlockHeader 区块链头
func (b *Block) BlockHeader() *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.Hash
	//obj.Confirmations = b.Confirmations
	obj.Merkleroot = b.BlockRoot
	obj.Previousblockhash = b.PrevBlockHash
	obj.Height = b.Height
	obj.Version = uint64(b.Version)
	obj.Time = b.Timestamp
	obj.Symbol = Symbol

	return &obj
}

//UnscanRecords 扫描失败的区块及交易
type UnscanRecord struct {
	ID          string `storm:"id"` // primary key
	BlockHeight uint64
	TxID        string
	Reason      string
}

func NewUnscanRecord(height uint64, txID, reason string) *UnscanRecord {
	obj := UnscanRecord{}
	obj.BlockHeight = height
	obj.TxID = txID
	obj.Reason = reason
	obj.ID = common.Bytes2Hex(crypto.SHA256([]byte(fmt.Sprintf("%d_%s", height, txID))))
	return &obj
}
