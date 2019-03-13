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

package virtualeconomy

import (
	"fmt"

	"github.com/blocktree/openwallet/crypto"
	"github.com/blocktree/openwallet/openwallet"
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
		{
		  "version" : 1,
		  "timestamp" : 1548758380000388132,
		  "reference" : "5cz6D8YMvqgtukHDwJhE3v7rJUaN2jeqVLXiVftjwyvHLvX7gSjtxGRwwqwE47UrQ6CLsgPfrpYTbviiHG87yS6d",
		  "SPOSConsensus" : {
		    "mintTime" : 1548758380000000000,
		    "mintBalance" : 6994026823501396
		  },
		  "resourcePricingData" : {
		    "computation" : 0,
		    "storage" : 0,
		    "memory" : 0,
		    "randomIO" : 0,
		    "sequentialIO" : 0
		  },
		  "TransactionMerkleRoot" : "DwtxNGjRC737nMsu7bikMFUM5oZwB9WrykaSpViGRDHv",
		  "transactions" : [ {
		    "type" : 2,
		    "id" : "9KBoALfTjvZLJ6CAuJCGyzRA1aWduiNFMvbqTchfBVpF",
		    "fee" : 10000000,
		    "timestamp" : 1548758345562000000,
		    "proofs" : [ {
		      "proofType" : "Curve25519",
		      "publicKey" : "3w7PrCnc2ZAjinAS5Ds4rXxvbrYo3LePXFhEwJev8XnN",
		      "signature" : "RQhdWC5CpSmn3DGmu7mfVpq5ZtSaorxgPv2PY13DvwKW7MVeuAojXStXe6kDVLcvHLLx6Jz1XTa1zhbinF57QjV"
		    } ],
		    "recipient" : "AREkgFxYhyCdtKD9JSSVhuGQomgGcacvQqM",
		    "feeScale" : 100,
		    "amount" : 5000000000,
		    "attachment" : "",
		    "status" : "Success",
		    "feeCharged" : 10000000
		  }, {
		    "type" : 5,
		    "id" : "FYNmFyt93EhS6XKgDwaJfqUwLgr95nwyPKH7M4Q7hFt7",
		    "recipient" : "ARFtmybrk4MTdbu8PYXy4QFvPeY2PEH7sCq",
		    "timestamp" : 1548758380000388132,
		    "amount" : 3600000000,
		    "currentBlockHeight" : 1352447,
		    "status" : "Success",
		    "feeCharged" : 0
		  } ],
		  "generator" : "AR86RMV3YiVoMDH6JdqgQWzPVu4Rdpjo7B4",
		  "signature" : "3Uvb87ukKKwVeU6BFsZ21hy9sSbSd3Rd5QZTWbNop1d3TaY9ZzceJAT54vuY8XXQmw6nDx8ZViPV3cVznAHTtiVE",
		  "fee" : 10000000,
		  "blocksize" : 500,
		  "height" : 1352447,
		  "transaction count" : 2
		}
	*/

	Hash                  string // actually block signature in VSYS chain
	Size                  uint64
	Version               byte
	PrevBlockHash         string // actually block signature in VSYS chain
	TransactionMerkleRoot string
	Timestamp             uint64
	Height                uint64
	Transactions          []string
}

type Transaction struct {
	TxType      byte
	TxID        string
	FeeScale    uint64
	FeeCharged  uint64
	Fee         uint64
	TimeStamp   uint64
	PublicKey   string
	Recipient   string
	Amount      uint64
	BlockHeight uint64
	BlockHash   string
	Status      bool
}

func NewTransaction(json *gjson.Result) *Transaction {
	obj := &Transaction{}

	obj.TxType = byte(json.Get("type").Uint())
	obj.TxID = json.Get("id").String()
	obj.Recipient = json.Get("recipient").String()
	obj.TimeStamp = json.Get("timestamp").Uint()
	obj.Amount = json.Get("amount").Uint()
	obj.BlockHeight = json.Get("height").Uint()
	obj.FeeCharged = json.Get("feeCharged").Uint()
	if json.Get("status").String() == "Success" {
		obj.Status = true
	} else {
		obj.Status = false
	}

	if obj.TxType == 2 {
		obj.Fee = json.Get("fee").Uint()
		obj.PublicKey = json.Get("proofs").Array()[0].Get("publicKey").String()
		obj.FeeScale = json.Get("feeScale").Uint()
	}
	return obj
}

func NewBlock(json *gjson.Result) *Block {
	obj := &Block{}

	// 解析
	obj.Hash = gjson.Get(json.Raw, "signature").String()
	obj.Size = gjson.Get(json.Raw, "blocksize").Uint()
	obj.Version = byte(gjson.Get(json.Raw, "version").Uint())
	obj.PrevBlockHash = gjson.Get(json.Raw, "reference").String()
	obj.TransactionMerkleRoot = gjson.Get(json.Raw, "TransactionMerkleRoot").String()
	obj.Timestamp = gjson.Get(json.Raw, "timestamp").Uint()
	obj.Height = gjson.Get(json.Raw, "height").Uint()

	txs := gjson.Get(json.Raw, "transactions").Array()

	for _, tx := range txs {
		obj.Transactions = append(obj.Transactions, gjson.Get(tx.Raw, "id").String())
	}

	return obj
}

//BlockHeader 区块链头
func (b *Block) BlockHeader() *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.Hash
	//obj.Confirmations = b.Confirmations
	obj.Merkleroot = b.TransactionMerkleRoot
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
