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
	"github.com/blocktree/OpenWallet/openwallet/accounts/keystore"
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
