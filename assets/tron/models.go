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
	"fmt"

	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcdrivers/addressEncoder"
	"github.com/bytom/common"
	"github.com/tidwall/gjson"
)

type Block struct {
	/*
		 {
			 "blockID":"000000000035e1c0f60afaa8387fd17fd9b84fe4381265ff084d739f814558ea",
			 "block_header":{
				 "raw_data":{"number":3531200,
						 "txTrieRoot":"9d98f6cbbde8302774ab87c003831333e132c89e00009cb1f7da35e1e59ae8ca",
						 "witness_address":"41d1dbde8b8f71b48655bec4f6bb532a0142b88bc0",
						 "parentHash":"000000000035e1bfb3b6f244e5316ce408aa8cea4c348eabe2545247f5a4600c",
						 "version":3,
						 "timestamp":1540545240000},
				 "witness_signature":"6ceedcacd8d0111b48eb4131484de3d13f27a2f4bd8156279d03d4208690158e20a641b77d900e026ee33adc328f9ec674f6483ea7b1ca5a27fa24d7fb23964100"
			 },
			 "transactions":[
				 {"ret":[{"contractRet":"SUCCESS"}],
				  "signature":["40aa520f01cebf12948615b9c5a5df5fe7d57e1a7f662d53907b4aa14f647a3a47be2a097fdb58159d0bee7eb1ff0a15ac738f24643fe5114cab8ec0d52cc04d01"],
				  "txID":"ac005b0a195a130914821a6c28db1eec44b4ec3a2358388ceb6c87b677866f1f",
				  "raw_data":{
					  "contract":[
						 {"parameter":{"value":{"amount":1,"asset_name":"48756f6269546f6b656e","owner_address":"416b201fb7b9f2b97bbdaf5e0920191229767c30ee","to_address":"412451d09536fca47760ea6513372bbbbef8583105"},
								   "type_url":"type.googleapis.com/protocol.TransferAssetContract"},
						  "type":"TransferAssetContract"}
					  ],
					  "ref_block_bytes":"e1be",
					  "ref_block_hash":"8dbf5f0cf4c324f2",
					  "expiration":1540545294000,
					  "timestamp":1540545235358}
				 },
				 ...]
		 }
	*/
	Hash              string // 这里采用 BlockID
	tx                []*Transaction
	Previousblockhash string
	Height            uint64 `storm:"id"`
	Version           uint64
	Time              uint64
	Fork              bool
	txHash            []string
	// Merkleroot        string
	//Confirmations     uint64
}

func NewBlock(json *gjson.Result) *Block {

	header := gjson.Get(json.Raw, "block_header").Get("raw_data")
	// 解析json
	b := &Block{}
	b.Hash = gjson.Get(json.Raw, "blockID").String()
	b.Previousblockhash = header.Get("parentHash").String()
	b.Height = header.Get("number").Uint()
	b.Version = header.Get("version").Uint()
	b.Time = header.Get("timestamp").Uint()

	txs := []*Transaction{}
	for _, x := range gjson.Get(json.Raw, "transactions").Array() {
		tx := NewTransaction(&x)
		txs = append(txs, tx)
	}
	b.tx = txs
	return b
}

func (block *Block) GetHeight() uint64 {
	return block.Height
}

func (block *Block) GetBlockHashID() string {
	return block.Hash
}

func (block *Block) GetTransactions() []*Transaction {
	return block.tx
}

type Transaction struct {
	/*
		 {
			  "txID":"ac005b0a195a130914821a6c28db1eec44b4ec3a2358388ceb6c87b677866f1f",
			  "ret":[{"contractRet":"SUCCESS"}],
			  "signature":["40aa520f01cebf12948615b9c5a5df5fe7d57e1a7f662d53907b4aa14f647a3a47be2a097fdb58159d0bee7eb1ff0a15ac738f24643fe5114cab8ec0d52cc04d01"],
			  "raw_data":{
				  "contract":[
					 {"parameter":{"value":{"amount":1,"asset_name":"48756f6269546f6b656e","owner_address":"416b201fb7b9f2b97bbdaf5e0920191229767c30ee","to_address":"412451d09536fca47760ea6513372bbbbef8583105"},
							   "type_url":"type.googleapis.com/protocol.TransferAssetContract"},
					  "type":"TransferAssetContract"}
				  ],
				  "ref_block_bytes":"e1be",
				  "ref_block_hash":"8dbf5f0cf4c324f2",
				  "expiration":1540545294000,
				  "timestamp":1540545235358}
		 },
	*/
	TxID        string
	ContractRet []map[string]string // 交易合约执行状态
	IsSuccess   bool
	//Size          uint64
	//Version       uint64
	//LockTime      int64
	//Hex           string
	BlockHash   string
	BlockHeight uint64
	//Confirmations uint64
	Blocktime        uint64
	IsCoinBase       bool
	FromAccountId    string //transaction scanning 的时候对其进行赋值
	ToAccountId      string //transaction scanning 的时候对其进行赋值
	From             string
	To               string
	Amount           string
	Contract_address string
	Type             string
	// Fees          string

}

func NewTransaction(json *gjson.Result) *Transaction {

	// 交易合约执行状态
	cr := []map[string]string{}
	isSuccess := true
	for _, ret := range gjson.Get(json.Raw, "ret").Array() {
		tmp := ret.Map()
		for k := range tmp {
			value := tmp[k].String()
			cr = append(cr, map[string]string{k: value})
			if value != "SUCCESS" {
				isSuccess = false
			}
		}
	}

	rawData := gjson.Get(json.Raw, "raw_data")
	// 解析json
	b := &Transaction{}
	b.TxID = gjson.Get(json.Raw, "txID").String()
	b.ContractRet = cr
	b.IsSuccess = isSuccess
	b.BlockHash = rawData.Get("ref_block_hash").String()
	b.BlockHeight = rawData.Get("ref_block_bytes").Uint()
	b.Type = rawData.Get("contract").Array()[0].Get("type").String()
	b.Blocktime = rawData.Get("timestamp").Uint()
	FromByte, _ := hex.DecodeString(rawData.Get("contract").Array()[0].Get("parameter").Get("value").Get("owner_address").String())
	b.From = addressEncoder.AddressEncode(FromByte[1:], addressEncoder.TRON_mainnetAddress)
	if b.Type == "TransferContract" {
		ToByte, _ := hex.DecodeString(rawData.Get("contract").Array()[0].Get("parameter").Get("value").Get("to_address").String())
		b.To = addressEncoder.AddressEncode(ToByte[1:], addressEncoder.TRON_mainnetAddress)
		b.Amount = rawData.Get("contract").Array()[0].Get("parameter").Get("value").Get("amount").String()
	} else {
		b.Amount = "0" //不需要任何操作，为了防止后面出错，将b.Amount设置为0
	}
	return b
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

//BlockHeader 区块链头
func (b *Block) Blockheader() *openwallet.BlockHeader {
	obj := openwallet.BlockHeader{}
	//解析josn
	obj.Merkleroot = ""
	obj.Hash = b.Hash
	obj.Previousblockhash = b.Previousblockhash
	obj.Height = b.Height
	obj.Version = uint64(b.Version)
	obj.Time = b.Time
	obj.Symbol = Symbol
	return &obj
}
