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

package sia

import (
	"github.com/tidwall/gjson"
	"github.com/blocktree/OpenWallet/openwallet/accounts/keystore"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common/file"
	"path/filepath"
)

//Wallet 钱包模型
type Wallet struct {

	WalletID string `json:"rootid"`
	Alias    string `json:"alias"`
	Balance  string `json:"balance"`
	Password string `json:"password"`
	RootPub  string `json:"rootpub"`
	KeyFile  string

	ConfirmBalance string `json:"confirmedsiacoinbalance"`

	OutgoingSC string `json:"unconfirmedoutgoingsiacoins"`
	IncomingSC string `json:"unconfirmedincomingsiacoins"`

	SiaFundBalance      string `json:"siafundbalance"`
	SiaCoinClaimBalance string `json:"siacoinclaimbalance"`

	Rescanning bool `json:"rescanning"`
	Unlocked   bool `json:"unlocked"`
	Encrypted  bool `json:"encrypted"`
}

//NewWallet 创建钱包
func NewWallet(json gjson.Result) *Wallet {
	w := &Wallet{}
	//解析json
	w.ConfirmBalance = gjson.Get(json.Raw, "confirmedsiacoinbalance").String()

	w.OutgoingSC = gjson.Get(json.Raw, "unconfirmedoutgoingsiacoins").String()
	w.IncomingSC = gjson.Get(json.Raw, "unconfirmedincomingsiacoins").String()

	w.SiaFundBalance = gjson.Get(json.Raw, "siafundbalance").String()
	w.SiaCoinClaimBalance = gjson.Get(json.Raw, "siacoinclaimbalance").String()

	w.Rescanning = gjson.Get(json.Raw, "rescanning").Bool()
	w.Unlocked = gjson.Get(json.Raw, "unlocked").Bool()
	w.Encrypted = gjson.Get(json.Raw, "encrypted").Bool()

	return w
}

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
	file := filepath.Join(dbPath, w.FileName()+".db")
	return storm.Open(file)

}

//FileName 该钱包定义的文件名规则
func (w *Wallet)FileName() string {
	return w.Alias+"-"+w.WalletID
}

type Account struct {
	Alias    string   `json:"alias"`
	ID       string   `json:"id"`
	KeyIndex int64    `json:"key_index"`
	Quorum   int64    `json:"quorum"`
	XPubs    []string `json:"xpubs"`
}

type Address struct {
	//Alias     string
	//AccountId string
	Address string
}

func NewAddress(json gjson.Result) *Address {

	a := &Address{}
	//解析json
	a.Address = gjson.Get(json.Raw, "address").String()
	return a
}

