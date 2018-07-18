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

package bitcoin

import (
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/keystore"
	"github.com/tidwall/gjson"
	"path/filepath"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/openwallet"
)

//Wallet 钱包模型
type Wallet struct {
	WalletID string `json:"rootid"`
	Alias    string `json:"alias"`
	Balance  string `json:"balance"`
	Password string `json:"password"`
	RootPub  string `json:"rootpub"`
	KeyFile  string
}

//NewWallet 创建钱包
//func NewWallet(json gjson.Result) *Wallet {
//	w := &Wallet{}
//	//解析json
//	w.Alias = gjson.Get(json.Raw, "alias").String()
//	w.PublicKey = gjson.Get(json.Raw, "xpub").String()
//	w.WalletID = common.NewString(w.PublicKey).SHA1()
//	return w
//}

//HDKey 获取钱包密钥，需要密码
func (w *Wallet) HDKey(password string) (*keystore.HDKey, error) {
	key, err := storage.GetKey(w.WalletID, w.KeyFile, password)
	if err != nil {
		return nil, err
	}
	return key, err
}

//openDB 打开钱包数据库
func (w *Wallet) OpenDB() (*storm.DB, error) {
	file.MkdirAll(dbPath)
	return storm.Open( w.DBFile())

}


//DBFile 数据库文件
func (w *Wallet)DBFile() string {
	return filepath.Join(dbPath, w.FileName()+".db")
}

//FileName 该钱包定义的文件名规则
func (w *Wallet)FileName() string {
	return w.Alias+"-"+w.WalletID
}

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
	WalletID       string `json:"account" storm:"index"`
	ScriptPubKey  string `json:"scriptPubKey"`
	Amount        string `json:"amount"`
	Confirmations uint64 `json:"confirmations"`
	Spendable     bool   `json:"spendable"`
	Solvable      bool   `json:"solvable"`
	HDAddress     openwallet.Address
}

func NewUnspent(json *gjson.Result) *Unspent {
	obj := &Unspent{}
	//解析json
	obj.TxID = gjson.Get(json.Raw, "txid").String()
	obj.Vout = gjson.Get(json.Raw, "vout").Uint()
	obj.Address = gjson.Get(json.Raw, "address").String()
	obj.WalletID = gjson.Get(json.Raw, "account").String()
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
