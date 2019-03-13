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

package bytom

import (
	"github.com/blocktree/openwallet/common"
	"github.com/tidwall/gjson"
	"github.com/asdine/storm"
	"github.com/blocktree/openwallet/common/file"
	"path/filepath"
)

//Wallet 钱包模型
type Wallet struct {
	WalletID       string `json:"walletID"`
	Alias          string `json:"alias"`
	Balance        string `json:"balance"`
	AccountsNumber uint64 `json:"accountsNumber"`
	Password       string `json:"password"`
	PublicKey      string `json:"xpub"`
}

//NewWallet 创建钱包
func NewWallet(json gjson.Result) *Wallet {
	w := &Wallet{}
	//解析json
	w.Alias = gjson.Get(json.Raw, "alias").String()
	w.PublicKey = gjson.Get(json.Raw, "xpub").String()
	w.WalletID = common.NewString(w.PublicKey).SHA1()
	return w
}

type Account struct {
	Alias    string   `json:"alias"`
	ID       string   `json:"id"`
	KeyIndex int64    `json:"key_index"`
	Quorum   int64    `json:"quorum"`
	XPubs    []string `json:"xpubs"`
}

func NewAccount(json gjson.Result) *Account {
	/*
		{
			"alias": "alice",
			"id": "08FO663C00A02",
			"key_index": 1,
			"quorum": 1,
			"xpubs": [
			"02581f1a2099e1696c498901c0049a22cc3e7f85db71c4ebb78f238d3ef8b323d2fd5c33b6f634aacdd25eb5e09c0c803077c521ef0524e4cc64d1a4420c8bc6"
			]
		}
	*/

	a := &Account{}
	//解析json
	a.Alias = gjson.Get(json.Raw, "alias").String()
	a.ID = gjson.Get(json.Raw, "id").String()
	a.KeyIndex = gjson.Get(json.Raw, "key_index").Int()
	a.Quorum = gjson.Get(json.Raw, "quorum").Int()

	a.XPubs = make([]string, 0)
	gjson.Get(json.Raw, "xpubs").ForEach(func(key, value gjson.Result) bool {
		a.XPubs = append(a.XPubs, value.String())
		return true
	})

	return a
}

//openDB 打开钱包数据库
func (a *Account) OpenDB() (*storm.DB, error) {
	file.MkdirAll(dbPath)
	file := filepath.Join(dbPath, a.FileName()+".db")
	return storm.Open(file)

}

//FileName 该钱包定义的文件名规则
func (w *Account)FileName() string {
	return w.Alias+"-"+w.ID
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

func NewAccountBalance(json gjson.Result) *AccountBalance {
	/*
		{
		      "account_alias": "default",
		      "account_id": "0BDQ9AP100A02",
		      "amount": 35508000000000,
		      "asset_alias": "BTM",
		      "asset_id": "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		    }
	*/

	a := &AccountBalance{}
	//解析json
	a.Alias = gjson.Get(json.Raw, "account_alias").String()
	a.AccountID = gjson.Get(json.Raw, "account_id").String()
	a.AssetAlias = gjson.Get(json.Raw, "asset_alias").String()
	a.Amount = gjson.Get(json.Raw, "amount").Uint()
	a.AssetID = gjson.Get(json.Raw, "asset_id").String()

	return a
}

type Address struct {
	Alias     string
	AccountId string
	Address   string
}

func NewAddress(accountID, alias string, json gjson.Result) *Address {

	/*
		{
		    "address": "bm1q5u8u4eldhjf3lvnkmyl78jj8a75neuryzlknk0",
		    "control_program": "0014a70fcae7edbc931fb276d93fe3ca47efa93cf064"
		}
	*/

	a := &Address{}
	//解析json
	a.Address = gjson.Get(json.Raw, "address").String()
	if len(accountID) == 0 {
		a.AccountId = gjson.Get(json.Raw, "account_id").String()
	} else {
		a.AccountId = accountID
	}

	if len(alias) == 0 {
		a.Alias = gjson.Get(json.Raw, "account_alias").String()
	} else {
		a.Alias = alias
	}

	return a
}
