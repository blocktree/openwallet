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

package qtum

import (
	"fmt"
	"github.com/blocktree/OpenWallet/crypto"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tidwall/gjson"
)

//BlockchainInfo 本地节点区块链信息
type BlockchainInfo struct {
	Chain                string `json:"chain"`
	Blocks               uint64 `json:"blocks"`
	Headers              uint64 `json:"headers"`
	Bestblockhash        string `json:"bestblockhash"`
	Difficulty           string `json:"difficulty"`
	Mediantime           uint64 `json:"mediantime"`
	Verificationprogress string `json:"verificationprogress"`
	Chainwork            string `json:"chainwork"`
	Pruned               bool   `json:"pruned"`
}

func NewBlockchainInfo(json *gjson.Result) *BlockchainInfo {
	b := &BlockchainInfo{}
	//解析json
	b.Chain = gjson.Get(json.Raw, "chain").String()
	b.Blocks = gjson.Get(json.Raw, "blocks").Uint()
	b.Headers = gjson.Get(json.Raw, "headers").Uint()
	b.Bestblockhash = gjson.Get(json.Raw, "bestblockhash").String()
	b.Difficulty = gjson.Get(json.Raw, "difficulty").String()
	b.Mediantime = gjson.Get(json.Raw, "mediantime").Uint()
	b.Verificationprogress = gjson.Get(json.Raw, "verificationprogress").String()
	b.Chainwork = gjson.Get(json.Raw, "chainwork").String()
	b.Pruned = gjson.Get(json.Raw, "pruned").Bool()
	return b
}

//Unspent 未花记录
type Unspent struct {

	/*
			{
		        "txid" : "d54994ece1d11b19785c7248868696250ab195605b469632b7bd68130e880c9a",
		        "vout" : 1,
		        "address" : "mgnucj8nYqdrPFh2JfZSB1NmUThUGnmsqe",
		        "account" : "test label",
		        "scriptPubKey" : "76a9140dfc8bafc8419853b34d5e072ad37d1a5159f58488ac",
		        "amount" : 0.00010000,
		        "confirmations" : 6210,
		        "spendable" : true,
		        "solvable" : true
		    }
	*/
	Key           string `storm:"id"`
	TxID          string `json:"txid"`
	Vout          uint64 `json:"vout"`
	Address       string `json:"address"`
	AccountID     string `json:"account" storm:"index"`
	ScriptPubKey  string `json:"scriptPubKey"`
	Amount        string `json:"amount"`
	Confirmations uint64 `json:"confirmations"`
	Spendable     bool   `json:"spendable"`
	Solvable      bool   `json:"solvable"`
	HDAddress     openwallet.Address
}

// AccountBalance account balance
type AccountBalance struct {
	AccountID  string `json:"account_id"`
	Alias      string `json:"account_alias"`
	AssetAlias string `json:"asset_alias"`
	AssetID    string `json:"asset_id"`
	Amount     uint64 `json:"amount"`
	Password   string
}

func NewUnspent(json *gjson.Result) *Unspent {
	obj := &Unspent{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Vout = gjson.Get(json.Raw, "vout").Uint()
	obj.Address = gjson.Get(json.Raw, "address").String()
	obj.AccountID = gjson.Get(json.Raw, "account").String()
	obj.ScriptPubKey = gjson.Get(json.Raw, "scriptPubKey").String()
	obj.Amount = gjson.Get(json.Raw, "amount").String()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	//obj.Spendable = gjson.Get(json.Raw, "spendable").Bool()
	obj.Spendable = true
	obj.Solvable = gjson.Get(json.Raw, "solvable").Bool()

	return obj
}

type UnspentSort struct {
	values     []*Unspent
	comparator func(a, b *Unspent) int
}

func (s UnspentSort) Len() int {
	return len(s.values)
}
func (s UnspentSort) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
}
func (s UnspentSort) Less(i, j int) bool {
	return s.comparator(s.values[i], s.values[j]) < 0
}

//type Address struct {
//	Address   string `json:"address" storm:"id"`
//	Account   string `json:"account" storm:"index"`
//	HDPath    string `json:"hdpath"`
//	CreatedAt time.Time
//}

type User struct {
	UserKey string `storm:"id"`     // primary key
	Group   string `storm:"index"`  // this field will be indexed
	Email   string `storm:"unique"` // this field will be indexed with a unique constraint
	Name    string // this field will not be indexed
	Age     int    `storm:"index"`
}

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

	Hash              string
	Confirmations     uint64
	Merkleroot        string
	tx                []string
	Previousblockhash string
	Height            uint64 `storm:"id"`
	Version           uint64
	Time              uint64
	Fork              bool
}

func NewBlock(json *gjson.Result) *Block {
	obj := &Block{}
	//解析json
	obj.Hash = gjson.Get(json.Raw, "hash").String()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	obj.Merkleroot = gjson.Get(json.Raw, "merkleroot").String()

	txs := make([]string, 0)
	for _, tx := range gjson.Get(json.Raw, "tx").Array() {
		txs = append(txs, tx.String())
	}

	obj.tx = txs
	obj.Previousblockhash = gjson.Get(json.Raw, "previousblockhash").String()
	obj.Height = gjson.Get(json.Raw, "height").Uint()
	obj.Version = gjson.Get(json.Raw, "version").Uint()
	obj.Time = gjson.Get(json.Raw, "time").Uint()

	return obj
}

//BlockHeader 区块链头
func (b *Block) BlockHeader() *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.Hash
	obj.Confirmations = b.Confirmations
	obj.Merkleroot = b.Merkleroot
	obj.Previousblockhash = b.Previousblockhash
	obj.Height = b.Height
	obj.Version = b.Version
	obj.Time = b.Time
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

//type Transaction struct {

/*

	{
		"txid": "c1e12febeb58aefb0b01c04360262138f4ee0faeb207276e79ea3866608ed84f",
		"hash": "c0bfbc4db1c6ed4356555c6f520df99640a42e39efa8939f4787d4c3d7aa2585",
		"version": 1,
		"size": 204,
		"vsize": 177,
		"locktime": 0,
		"vin": [{
			"coinbase": "0308ac1404a4a5525b081ffffe24dcd602000d2ff09fa498f09f988e204d722e204d6f2f",
			"sequence": 0
		}],
		"vout": [{
			"value": 0.00000000,
			"n": 0,
			"scriptPubKey": {
				"asm": "OP_RETURN aa21a9edf9be7d36da3e5fea8130b031d254b9a6ff2dd471fbceb0f460d8fcca101d27ad",
				"hex": "6a24aa21a9edf9be7d36da3e5fea8130b031d254b9a6ff2dd471fbceb0f460d8fcca101d27ad",
				"type": "nulldata"
			}
		}, {
			"value": 1.40257806,
			"n": 1,
			"scriptPubKey": {
				"asm": "OP_DUP OP_HASH160 48b5c6986b7bc6390bd1cc416154d1874fe116fd OP_EQUALVERIFY OP_CHECKSIG",
				"hex": "76a91448b5c6986b7bc6390bd1cc416154d1874fe116fd88ac",
				"reqSigs": 1,
				"type": "pubkeyhash",
				"addresses": ["mn9QhsFiX2eEXtF6zrGn5N49iS8BHXFjBt"]
			}
		}],
		"hex": "010000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff240308ac1404a4a5525b081ffffe24dcd602000d2ff09fa498f09f988e204d722e204d6f2f00000000020000000000000000266a24aa21a9edf9be7d36da3e5fea8130b031d254b9a6ff2dd471fbceb0f460d8fcca101d27ad0e2a5c08000000001976a91448b5c6986b7bc6390bd1cc416154d1874fe116fd88ac0120000000000000000000000000000000000000000000000000000000000000000000000000",
		"blockhash": "000000000000000127454a8c91e74cf93ad76752cceb7eb3bcff0c398ba84b1f",
		"confirmations": 25,
		"time": 1532143012,
		"blocktime": 1532143012
	}

*/
//}

type QRC20Unspent struct {

	/*
		{
		  "address": "91a6081095ef860d28874c9db613e7a4107b0281",
		  "executionResult": {
			"gasUsed": 23343,
			"excepted": "None",
			"newAddress": "91a6081095ef860d28874c9db613e7a4107b0281",
			"output": "000000000000000000000000000000000000000000000000016345785d8a0000",
			"codeDeposit": 0,
			"gasRefunded": 0,
			"depositSize": 0,
			"gasForDeposit": 0
		  },
		  "transactionReceipt": {
			"stateRoot": "2aeed0cff1334f1387b663079757ad1cff6fe72e8a37cfcaf051d564b7252d63",
			"gasUsed": 23343,
			"bloom": "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
					  0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
					  0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
					  0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
					  0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			"log": [
			]
		  }
		}
	*/

	Key     string `storm:"id"`
	Address string `json:"address"`
	GasUsed string `json:"gasUsed"`
	Output  string `json:"output"`
	//HDAddress     openwallet.Address

}

func NewQRC20Unspent(json *gjson.Result) *QRC20Unspent {
	obj := &QRC20Unspent{}
	//解析json
	obj.Address = gjson.Get(json.Raw, "address").String()
	obj.GasUsed = gjson.Get(json.Raw, "executionResult.gasUsed").String()
	obj.Output = gjson.Get(json.Raw, "executionResult.output").String()

	return obj
}

type Transaction struct {
	TxID            string
	Size            uint64
	Version         uint64
	LockTime        int64
	Hex             string
	BlockHash       string
	BlockHeight     uint64
	Confirmations   uint64
	Blocktime       int64
	IsCoinBase      bool
	Fees            string
	Isqrc20Transfer bool

	Vins          []*Vin
	Vouts         []*Vout
	TokenReceipts []*TokenReceipt
}

type Vin struct {
	Coinbase string
	TxID     string
	Vout     uint64
	N        uint64
	Addr     string
	Value    string
}

type Vout struct {
	N            uint64
	Addr         string
	Value        string
	ScriptPubKey string
	Type         string
}

type TokenReceipt struct {

	TxHash            string
	BlockHash       string
	BlockHeight     uint64
	Sender string
	From string
	To string
	GasUsed uint64
	ContractAddress string
	Excepted string
	Amount string
}

func newTxByCore(json *gjson.Result, isTestnet bool) *Transaction {

	/*
		{
			"txid": "6595e0d9f21800849360837b85a7933aeec344a89f5c54cf5db97b79c803c462",
			"hash": "f758cb5181d51f8bee1512b4a862faad5b51c7c85a1a11cd6092ffc1c6649bc5",
			"version": 2,
			"size": 249,
			"vsize": 168,
			"locktime": 1414190,
			"vin": [],
			"vout": [],
			"hex": "02000000000101cc8a3077023c08040e677647ad0e528564764f456b01d8519828df165ab3c4550100000017160014aa59f94152351c79b57b14a53e538a923e332468feffffff02a716167c6f00000017a914a0fe07f130a36d9c7581ccd2886895c049b0cc8287ece29c00000000001976a9148c0bceb59d452b3e077f73a420b8bfe09e0550a788ac0247304402205e667171c1798cde426282bb8bff45901866ad6bf0d209e856c1765eda65ba4802203aaa319ea3de00eccef0006e6ee2089aed4b91ada7953f420a47c9c258d424ca0121033cfda2f93d13b01d46ecc406b03ebaba3e1bd526d2148a0a5d579d52f8c7cf022e941500",
			"blockhash": "0000000040730ea7935cce346ce68bf4c07c10b137ba31960bf8a47c4f7da4ec",
			"confirmations": 20076,
			"time": 1537841342,
			"blocktime": 1537841342
		}
	*/
	obj := Transaction{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Version = gjson.Get(json.Raw, "version").Uint()
	obj.LockTime = gjson.Get(json.Raw, "locktime").Int()
	obj.BlockHash = gjson.Get(json.Raw, "blockhash").String()
	//obj.BlockHeight = gjson.Get(json.Raw, "blockheight").Uint()
	obj.Confirmations = gjson.Get(json.Raw, "confirmations").Uint()
	obj.Blocktime = gjson.Get(json.Raw, "blocktime").Int()
	obj.Size = gjson.Get(json.Raw, "size").Uint()
	//obj.Fees = gjson.Get(json.Raw, "fees").String()

	obj.Vins = make([]*Vin, 0)
	if vins := gjson.Get(json.Raw, "vin"); vins.IsArray() {
		for i, vin := range vins.Array() {
			input := newTxVinByCore(&vin)
			input.N = uint64(i)
			obj.Vins = append(obj.Vins, input)
		}
	}

	obj.Vouts = make([]*Vout, 0)
	if vouts := gjson.Get(json.Raw, "vout"); vouts.IsArray() {
		for _, vout := range vouts.Array() {
			output := newTxVoutByCore(&vout, isTestnet)
			obj.Vouts = append(obj.Vouts, output)
		}
	}

	return &obj
}

func newTxVinByCore(json *gjson.Result) *Vin {

	/*
		{
			"txid": "55c4b35a16df289851d8016b454f766485520ead4776670e04083c0277308acc",
			"vout": 1,
			"scriptSig": {
				"asm": "0014aa59f94152351c79b57b14a53e538a923e332468",
				"hex": "160014aa59f94152351c79b57b14a53e538a923e332468"
			},
			"txinwitness": ["304402205e667171c1798cde426282bb8bff45901866ad6bf0d209e856c1765eda65ba4802203aaa319ea3de00eccef0006e6ee2089aed4b91ada7953f420a47c9c258d424ca01", "033cfda2f93d13b01d46ecc406b03ebaba3e1bd526d2148a0a5d579d52f8c7cf02"],
			"sequence": 4294967294
		}
	*/
	obj := Vin{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Vout = gjson.Get(json.Raw, "vout").Uint()
	obj.Coinbase = gjson.Get(json.Raw, "coinbase").String()
	//obj.Addr = gjson.Get(json.Raw, "addr").String()
	//obj.Value = gjson.Get(json.Raw, "value").String()

	return &obj
}

func newTxVoutByCore(json *gjson.Result, isTestnet bool) *Vout {

	/*
		{
			"value": 4788.23192231,
			"n": 0,
			"scriptPubKey": {
				"asm": "OP_HASH160 a0fe07f130a36d9c7581ccd2886895c049b0cc82 OP_EQUAL",
				"hex": "a914a0fe07f130a36d9c7581ccd2886895c049b0cc8287",
				"reqSigs": 1,
				"type": "scripthash",
				"addresses": ["2N7vURMwMDjqgijLNFsErFLAWtAg58S6qNv"]
			}
		}
	*/
	obj := Vout{}
	//解析json
	obj.Value = gjson.Get(json.Raw, "value").String()
	obj.N = gjson.Get(json.Raw, "n").Uint()
	obj.ScriptPubKey = gjson.Get(json.Raw, "scriptPubKey.hex").String()

	//提取地址
	if addresses := gjson.Get(json.Raw, "scriptPubKey.addresses"); addresses.IsArray() {
		obj.Addr = addresses.Array()[0].String()
	}

	obj.Type = gjson.Get(json.Raw, "scriptPubKey.type").String()

	//if len(obj.Addr) == 0 {
	//	scriptBytes, _ := hex.DecodeString(obj.ScriptPubKey)
	//	obj.Addr, _ = ScriptPubKeyToBech32Address(scriptBytes, isTestnet)
	//}

	return &obj
}
