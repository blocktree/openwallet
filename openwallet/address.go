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

import (
	"github.com/tidwall/gjson"
	"time"
)

//Address OpenWallet地址
type Address struct {
	WalletID  string    `json:"walletID"`           //钱包ID
	Address   string    `json:"address" storm:"id"` //地址字符串
	Alias     string    `json:"alias"`              //地址别名，可绑定用户
	Tag       string    `json:"tag"`                //标签
	Index     uint64    `json:"index"`              //账户ID，索引位
	RootPath  string    `json:"rootPath"`           //地址公钥根路径
	WatchOnly bool      `json:"watchOnly"`          //是否观察地址，true的时候，Index，RootPath，Alias都没有。
	Symbol    string    `json:"coin"`               //币种类别
	Balance   string    `json:"balance"`            //余额
	IsMemo    bool      `json:"isMemo"`             //是否备注
	Memo      string    `json:"memo"`               //备注
	CreatedAt time.Time `json:"createdAt"`

	//核心地址指针
	core interface{}
}

func NewAddress(json gjson.Result) *Address {
	obj := &Address{}
	//解析json
	obj.WalletID = gjson.Get(json.Raw, "walletID").String()
	obj.Address = gjson.Get(json.Raw, "address").String()
	obj.Alias = gjson.Get(json.Raw, "alias").String()
	obj.IsMemo = gjson.Get(json.Raw, "isMemo").Bool()
	obj.Memo = gjson.Get(json.Raw, "memo").String()
	obj.Tag = gjson.Get(json.Raw, "tag").String()
	obj.Index = gjson.Get(json.Raw, "index").Uint()
	obj.RootPath = gjson.Get(json.Raw, "rootPath").String()
	obj.core = gjson.Get(json.Raw, "coin").String()
	obj.Balance = gjson.Get(json.Raw, "balance").String()

	return obj
}
